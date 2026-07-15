#!/usr/bin/env bash
# Medtasker Skills installer.
#
# Run from a fresh clone (the repo is private — no `curl | bash` over the web):
#   git clone git@github.com:nimblic/medtasker-skills.git
#   cd medtasker-skills && ./scripts/install.sh

set -euo pipefail

# Must be run from the repo root (the directory above scripts/).
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"
if [ ! -f go.mod ] || ! grep -q '^module github.com/nimblic/medtasker-skills' go.mod; then
    echo "ERROR: run ./scripts/install.sh from the medtasker-skills repo root." >&2
    exit 1
fi

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[1;33m'
BLUE=$'\033[0;34m'
NC=$'\033[0m'

info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
fail()  { echo -e "${RED}[ERROR]${NC} $1"; }

cmd_exists() { command -v "$1" >/dev/null 2>&1; }

require_go() {
    if ! cmd_exists go; then
        fail "Go is required. Install from https://go.dev/dl/"
        exit 1
    fi
    ok "Go: $(go version)"
}

install_dotenvx() {
    if cmd_exists dotenvx; then
        ok "dotenvx already installed"
        return
    fi
    info "Installing dotenvx"
    if cmd_exists npm; then
        npm install -g @dotenvx/dotenvx && return
    fi
    if cmd_exists curl; then
        curl -sfS https://dotenvx.sh/install.sh | sh && return
    fi
    fail "Could not install dotenvx. See https://dotenvx.com/docs/install"
    exit 1
}

install_cli() {
    info "Building medtasker-skills"
    # Build from the cloned repo (go install from network is impossible — repo
    # is private). Default install dir: ~/.local/bin. Override with INSTALL_DIR.
    local install_dir="${INSTALL_DIR:-$HOME/.local/bin}"
    mkdir -p "$install_dir"
    go build -o "$install_dir/medtasker-skills" ./cmd/medtasker-skills
    ok "medtasker-skills built at $install_dir/medtasker-skills"
    case ":$PATH:" in
        *":$install_dir:"*) ;;
        *) warn "$install_dir is not on PATH — add it to your shell rc, or move the binary to a PATH dir." ;;
    esac
}

install_skills() {
    if ! cmd_exists medtasker-skills; then
        warn "medtasker-skills not on PATH. Restart your shell, then run: medtasker-skills install"
        return
    fi
    medtasker-skills install
}

post_install() {
    cat <<EOF

========================================
  Medtasker Skills installed
========================================

Next steps:
  1. Run the interactive setup wizard:
       medtasker-skills env setup

     Or set credentials manually:
       medtasker-skills env set JIRA_HOST https://yourcompany.atlassian.net
       medtasker-skills env set JIRA_USERNAME you@example.com
       medtasker-skills env set JIRA_API_TOKEN <token>

  2. Launch Claude Code with the vault decrypted in-process:
       dotenvx run -f ~/.medtasker-skills/.env -- claude

  3. Verify with: medtasker-skills doctor

EOF
}

main() {
    echo "Medtasker Skills installer"
    require_go
    install_dotenvx
    install_cli
    install_skills
    post_install
}

case "${1:-}" in
    --help|-h)
        cat <<'EOF'
Usage: ./scripts/install.sh

Run from a fresh clone of the medtasker-skills repo. The repo is private,
so go install / curl | bash from the public web don't work.

Env overrides:
  INSTALL_DIR=<dir>   Where to put the binary (default: ~/.local/bin)
EOF
        exit 0
        ;;
esac

main
