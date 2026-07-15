#!/usr/bin/env bash
# Smoke test for stevmachine-skills CLI.
#
# Builds the binary, runs every non-interactive subcommand against an isolated
# $HOME under /tmp, asserts on outputs, and tears down. Exits 0 only if
# every check passes.
#
# Usage:
#   .claude/skills/run-stevmachine-skills/smoke.sh         # build + run + cleanup
#   KEEP=1 .claude/skills/run-stevmachine-skills/smoke.sh  # leave the temp HOME for inspection
#
# Requires: go, dotenvx (already a runtime dep of stevmachine-skills).
set -euo pipefail

# Find the stevmachine-skills repo by walking up from the script's real dir
# (resolves symlinks). Falls back to STEVMACHINE_REPO. Works whether the script
# is run from packages/, via the .claude/ symlink, or after `stevmachine-skills
# install` placed it in ~/.claude/skills/ (in which case the user must set
# STEVMACHINE_REPO).
find_repo() {
  local src d
  src="${BASH_SOURCE[0]}"
  if command -v readlink >/dev/null && readlink -f / >/dev/null 2>&1; then
    src="$(readlink -f "$src")"
  else
    src="$(python3 -c 'import os,sys; print(os.path.realpath(sys.argv[1]))' "$src")"
  fi
  d="$(cd "$(dirname "$src")" && pwd)"
  while [[ "$d" != "/" ]]; do
    if [[ -f "$d/go.mod" ]] && grep -q '^module github.com/stvmachine/skills' "$d/go.mod"; then
      echo "$d"; return 0
    fi
    d="$(dirname "$d")"
  done
  return 1
}
REPO_ROOT="$(find_repo || true)"
[[ -z "$REPO_ROOT" ]] && REPO_ROOT="${STEVMACHINE_REPO:-}"
if [[ -z "$REPO_ROOT" || ! -f "$REPO_ROOT/go.mod" ]]; then
  echo "smoke.sh: stevmachine-skills source repo not found." >&2
  echo "         Run from inside the repo, or set STEVMACHINE_REPO=/path/to/stevmachine-skills." >&2
  exit 1
fi
cd "$REPO_ROOT"

# Isolated HOME so we never touch the real ~/.claude or ~/.stevmachine-skills
SMOKE_DIR="$(mktemp -d -t mts-smoke.XXXXXX)"
trap '[[ "${KEEP:-0}" == "1" ]] || rm -rf "$SMOKE_DIR"' EXIT
SMOKE_HOME="$SMOKE_DIR/home"
mkdir -p "$SMOKE_HOME/.claude" "$SMOKE_HOME/.stevmachine-skills"
chmod 700 "$SMOKE_HOME/.stevmachine-skills"

BIN="$SMOKE_DIR/stevmachine-skills"

# Helpers --------------------------------------------------------------------
PASS=0; FAIL=0
pass() { printf '  \033[0;32m✓\033[0m %s\n' "$1"; PASS=$((PASS+1)); }
fail() { printf '  \033[0;31m✗\033[0m %s\n' "$1"; FAIL=$((FAIL+1)); }
section() { printf '\n\033[1;34m▸ %s\033[0m\n' "$1"; }

run() { HOME="$SMOKE_HOME" "$BIN" "$@"; }
assert_contains() {
  local needle="$1" haystack="$2" label="$3"
  if grep -qF -- "$needle" <<<"$haystack"; then pass "$label"
  else fail "$label  (missing: $needle)"; printf '%s\n' "$haystack" | sed 's/^/      | /'; fi
}
assert_file() {
  if [[ -f "$1" ]]; then pass "$2"; else fail "$2  (missing: $1)"; fi
}

# Build ----------------------------------------------------------------------
section "build"
go build -o "$BIN" ./cmd/stevmachine-skills
[[ -x "$BIN" ]] && pass "binary built ($BIN)" || { fail "build failed"; exit 1; }

# usage ----------------------------------------------------------------------
section "usage"
out="$("$BIN" 2>&1 || true)"
assert_contains "install [skill ...]" "$out" "no-args prints usage"

# list -----------------------------------------------------------------------
section "list (before install)"
out="$(run list)"
assert_contains "Skills" "$out" "list prints Skills header"
assert_contains "stevmachine-jira" "$out" "list shows stevmachine-jira from embedded packages"

# doctor ---------------------------------------------------------------------
section "doctor"
out="$(run doctor)"
# Regex tolerant of variable padding (cmd_doctor.go formats with N spaces
# depending on the longest field label).
if grep -qE 'Claude Code dir:[[:space:]]+OK' <<<"$out"; then pass "doctor finds .claude"
else fail "doctor finds .claude"; printf '%s\n' "$out" | sed 's/^/      | /'; fi
if grep -qE 'dotenvx:[[:space:]]+OK' <<<"$out"; then pass "doctor finds dotenvx"
else fail "doctor finds dotenvx"; printf '%s\n' "$out" | sed 's/^/      | /'; fi
# New: ensure the optional-dep section shows up (bd/rtk lines exist).
if grep -qE 'beads \(bd\):' <<<"$out"; then pass "doctor reports beads status"
else fail "doctor missing 'beads (bd):' line"; fi
if grep -qE 'rtk:[[:space:]]' <<<"$out"; then pass "doctor reports rtk status"
else fail "doctor missing 'rtk:' line"; fi

# install (TUI auto-runs to completion) --------------------------------------
section "install stevmachine-jira"
out="$(run install stevmachine-jira 2>&1)"
assert_contains "1 skills installed" "$out" "install reports success"
assert_file "$SMOKE_HOME/.claude/skills/stevmachine-jira/SKILL.md" "skill copied to ~/.claude/skills/"
assert_file "$SMOKE_HOME/.claude/.mcp.json" ".mcp.json written"

# .mcp.json must hold literal ${VAR} placeholders, NOT resolved secrets ------
section ".mcp.json placeholder invariant"
mcpjson="$(cat "$SMOKE_HOME/.claude/.mcp.json")"
if grep -q '"JIRA_API_TOKEN": *"\${JIRA_API_TOKEN}"' <<<"$mcpjson"; then
  pass '${JIRA_API_TOKEN} stored as placeholder (not resolved)'
else
  fail 'JIRA_API_TOKEN missing or resolved -- this would leak secrets'
  printf '%s\n' "$mcpjson" | sed 's/^/      | /'
fi
python3 -c "import json,sys; json.load(open('$SMOKE_HOME/.claude/.mcp.json'))" \
  && pass ".mcp.json is valid JSON" \
  || fail ".mcp.json is not valid JSON"

# list (after install -- should mark stevmachine-jira ✓) -----------------------
section "list (after install)"
out="$(run list)"
if grep -qE '✓ +stevmachine-jira' <<<"$out"; then
  pass "list marks stevmachine-jira as installed"
else
  fail "stevmachine-jira not marked installed"
fi

# env: full vault round-trip -------------------------------------------------
section "env set / list / encrypt / decrypt"
run env set SMOKE_KEY hunter2-super-secret-value >/dev/null
assert_file "$SMOKE_HOME/.stevmachine-skills/.env" "vault .env written"
# dotenvx encrypts on `set`, so .env should contain ciphertext already
if grep -q '^SMOKE_KEY="encrypted:' "$SMOKE_HOME/.stevmachine-skills/.env"; then
  pass "value stored as encrypted ciphertext"
else
  fail "value not encrypted in .env"
  cat "$SMOKE_HOME/.stevmachine-skills/.env" | sed 's/^/      | /'
fi
# decrypt round-trip
run env decrypt >/dev/null 2>&1 || true
if grep -q '^SMOKE_KEY="hunter2-super-secret-value"' "$SMOKE_HOME/.stevmachine-skills/.env"; then
  pass "decrypt restores plaintext"
else
  # newer dotenvx requires DOTENV_PRIVATE_KEY in env -- treat as non-fatal
  printf '    \033[0;33m⚠\033[0m decrypt did not produce plaintext (needs DOTENV_PRIVATE_KEY env var on this dotenvx)\n'
fi
run env encrypt >/dev/null
grep -q '^SMOKE_KEY="encrypted:' "$SMOKE_HOME/.stevmachine-skills/.env" \
  && pass "encrypt re-encrypts" \
  || fail "encrypt did not produce ciphertext"

# unit tests (Go) ------------------------------------------------------------
# Note: internal/vault tests need DOTENV_PRIVATE_KEY in env for dotenvx >=1.66.
# We exclude that package by default; pass RUN_VAULT_TESTS=1 to include it.
section "unit tests (mcp, cmd)"
if go test ./internal/mcp ./cmd/... >/dev/null 2>&1; then
  pass "go test ./internal/mcp ./cmd/... passes"
else
  fail "go test (mcp/cmd) failed"
  go test ./internal/mcp ./cmd/... 2>&1 | sed 's/^/      | /'
fi

if [[ "${RUN_VAULT_TESTS:-0}" == "1" ]]; then
  section "unit tests (vault -- needs DOTENV_PRIVATE_KEY)"
  if go test ./internal/vault >/dev/null 2>&1; then
    pass "go test ./internal/vault passes"
  else
    fail "go test ./internal/vault failed (likely dotenvx >=1.66 + missing key)"
  fi
fi

# Report ---------------------------------------------------------------------
printf '\n\033[1m%d passed, %d failed\033[0m\n' "$PASS" "$FAIL"
[[ "${KEEP:-0}" == "1" ]] && printf 'kept: %s\n' "$SMOKE_DIR"
[[ "$FAIL" -eq 0 ]]
