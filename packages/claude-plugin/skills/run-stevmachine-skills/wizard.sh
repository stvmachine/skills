#!/usr/bin/env bash
# Drive `medtasker-skills env setup` (the huh wizard) end-to-end via tmux,
# choosing one of three canned profiles. Verifies the wizard actually wrote
# encrypted ciphertext to ~/.medtasker-skills/.env in the isolated $HOME.
#
# Usage:
#   wizard.sh                        # decline-all profile (no creds entered)
#   wizard.sh jira                   # configure Jira with dummy values
#   wizard.sh all                    # configure Jira + GitHub with dummy values
#   PROFILE=all DEBUG=1 wizard.sh    # dump every frame
#
# Why tmux: huh refuses to run without a TTY. tmux gives us a PTY plus
# capture-pane so we can verify each frame before sending the next key.
#
# Requires: tmux, go, dotenvx.
set -euo pipefail

PROFILE="${1:-${PROFILE:-decline}}"

# Same repo-finder as smoke.sh — walks up looking for the medtasker-skills go.mod.
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
    if [[ -f "$d/go.mod" ]] && grep -q '^module github.com/nimblic/medtasker-skills' "$d/go.mod"; then
      echo "$d"; return 0
    fi
    d="$(dirname "$d")"
  done
  return 1
}
REPO_ROOT="$(find_repo || true)"
[[ -z "$REPO_ROOT" ]] && REPO_ROOT="${MEDTASKER_REPO:-}"
if [[ -z "$REPO_ROOT" || ! -f "$REPO_ROOT/go.mod" ]]; then
  echo "wizard.sh: medtasker-skills source repo not found." >&2
  echo "           Run from inside the repo, or set MEDTASKER_REPO=/path/to/medtasker-skills." >&2
  exit 1
fi
cd "$REPO_ROOT"

command -v tmux >/dev/null || { echo "tmux required (brew install tmux)" >&2; exit 1; }

WIZ_DIR="$(mktemp -d -t mts-wizard.XXXXXX)"
trap '[[ "${KEEP:-0}" == "1" ]] || { tmux kill-session -t mts-wiz 2>/dev/null || true; rm -rf "$WIZ_DIR"; }' EXIT
WIZ_HOME="$WIZ_DIR/home"
mkdir -p "$WIZ_HOME/.claude" "$WIZ_HOME/.medtasker-skills"
chmod 700 "$WIZ_HOME/.medtasker-skills"

BIN="$WIZ_DIR/medtasker-skills"
go build -o "$BIN" ./cmd/medtasker-skills

# tmux session helpers -------------------------------------------------------
SESSION=mts-wiz
tmux kill-session -t "$SESSION" 2>/dev/null || true
tmux new-session -d -s "$SESSION" -x 200 -y 60 \
  "HOME=$WIZ_HOME $BIN env setup; echo ___WIZARD_EXITED___; sleep 5"

frame() { tmux capture-pane -t "$SESSION" -p; }
expect() {
  # Wait up to ~3s for the screen to contain $1
  local needle="$1" tries=0
  while (( tries < 30 )); do
    if grep -qF -- "$needle" <(frame); then
      [[ "${DEBUG:-0}" == "1" ]] && { echo "--- expect OK: $needle ---"; frame | sed 's/^/  | /'; }
      return 0
    fi
    sleep 0.1
    tries=$((tries+1))
  done
  echo "TIMEOUT waiting for: $needle" >&2
  frame | sed 's/^/  | /' >&2
  return 1
}
send() {
  [[ "${DEBUG:-0}" == "1" ]] && echo "--- send: $* ---"
  tmux send-keys -t "$SESSION" "$@"
  sleep 0.2
}
type_text() { send -l "$1"; }     # literal text (no key-name translation)
hit_enter() { send Enter; }

# Drive ---------------------------------------------------------------------
expect "Medtasker Skills Setup" || exit 1
hit_enter                                       # dismiss the intro Note

case "$PROFILE" in
  decline)
    expect "Configure Jira?";       send n
    expect "Configure GitHub?";     send n
    expect "Configure Confluence?"; send n
    expect "Configure Figma?";      send n
    expect "Configure Context7?";   send n
    ;;
  jira)
    expect "Configure Jira?";       send y
    expect "JIRA_URL";              type_text "https://example.atlassian.net"; hit_enter
    expect "JIRA_USERNAME";         type_text "smoke@example.com";             hit_enter
    expect "JIRA_API_TOKEN";        type_text "dummy-jira-token-not-real";     hit_enter
    expect "Configure GitHub?";     send n
    expect "Configure Confluence?"; send n
    expect "Configure Figma?";      send n
    expect "Configure Context7?";   send n
    ;;
  all)
    expect "Configure Jira?";       send y
    expect "JIRA_URL";              type_text "https://example.atlassian.net"; hit_enter
    expect "JIRA_USERNAME";         type_text "smoke@example.com";             hit_enter
    expect "JIRA_API_TOKEN";        type_text "dummy-jira-token-not-real";     hit_enter
    expect "Configure GitHub?";     send y
    expect "GITHUB_TOKEN";          type_text "ghp_dummy_github_token_xxx";    hit_enter
    expect "Configure Confluence?"; send n
    expect "Configure Figma?";      send n
    expect "Configure Context7?";   send n
    ;;
  *) echo "unknown profile: $PROFILE (decline | jira | all)" >&2; exit 2 ;;
esac

# Wait for the binary to exit and final summary to print
expect "___WIZARD_EXITED___" || { echo "wizard did not exit cleanly" >&2; exit 1; }
final=$(frame)
tmux kill-session -t "$SESSION" 2>/dev/null || true

# Verify ---------------------------------------------------------------------
echo
if [[ "$PROFILE" == "decline" ]]; then
  if grep -qF "Stored 0 variable(s)" <<<"$final"; then
    echo "✓ decline profile stored 0 variables"
  else
    echo "✗ decline profile: expected 'Stored 0 variable(s)'"; exit 1
  fi
else
  if grep -qE "Stored [1-9][0-9]* variable\(s\)" <<<"$final"; then
    n=$(grep -oE "Stored [0-9]+ variable" <<<"$final" | grep -oE '[0-9]+')
    echo "✓ profile=$PROFILE stored $n variables"
  else
    echo "✗ wizard summary missing 'Stored N variable(s)'"; exit 1
  fi
  if grep -q '^JIRA_URL="encrypted:' "$WIZ_HOME/.medtasker-skills/.env"; then
    echo "✓ JIRA_URL written as ciphertext"
  else
    echo "✗ JIRA_URL not encrypted in .env"
    cat "$WIZ_HOME/.medtasker-skills/.env" | sed 's/^/  | /'
    exit 1
  fi
fi
[[ "${KEEP:-0}" == "1" ]] && echo "kept: $WIZ_DIR"
