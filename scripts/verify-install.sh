#!/usr/bin/env bash
# Post-install sanity check for Medtasker Skills.
# Usage: ./verify-install.sh

set -euo pipefail

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[1;33m'
NC=$'\033[0m'

PASS=0; FAIL=0; WARN=0
pass() { echo -e "${GREEN}[PASS]${NC} $1"; PASS=$((PASS+1)); }
fail() { echo -e "${RED}[FAIL]${NC} $1"; FAIL=$((FAIL+1)); }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; WARN=$((WARN+1)); }

check_cmd() {
    if command -v "$1" >/dev/null 2>&1; then pass "$1 on PATH"
    else fail "$1 not on PATH"; fi
}

echo "Medtasker Skills install verification"

check_cmd go || true
check_cmd dotenvx
check_cmd medtasker-skills

VAULT_DIR="$HOME/.medtasker-skills"
if [ -d "$VAULT_DIR" ]; then
    pass "Vault directory exists: $VAULT_DIR"
    PERMS=$(stat -c "%a" "$VAULT_DIR" 2>/dev/null || stat -f "%Lp" "$VAULT_DIR" 2>/dev/null || echo "?")
    if [ "$PERMS" = "700" ]; then pass "Vault permissions 700"
    else warn "Vault permissions $PERMS (expected 700)"; fi
else
    warn "Vault directory missing: $VAULT_DIR"
fi

if [ -f "$VAULT_DIR/.env.vault" ]; then pass "Encrypted vault present"
else warn "No .env.vault yet — run: medtasker-skills env setup"; fi

if [ -d "$HOME/.claude" ]; then pass "Claude Code directory present"
else fail "Claude Code not detected (~/.claude missing)"; fi

if [ -f "$HOME/.claude/.mcp.json" ]; then pass "MCP config present"
else warn "No .mcp.json — run: medtasker-skills install"; fi

if command -v medtasker-skills >/dev/null 2>&1; then
    medtasker-skills doctor || warn "medtasker-skills doctor reported issues"
fi

echo ""
echo "Results: $PASS passed, $FAIL failed, $WARN warnings"
[ "$FAIL" -eq 0 ]
