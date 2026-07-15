#!/usr/bin/env bash
# Stevmachine Skills installer.
#
# Run from a fresh clone (this is a public repo):
#   git clone git@github.com:stvmachine/skills.git
#   cd skills && ./scripts/install.sh

set -euo pipefail

# Must be run from the repo root (the directory above scripts/).
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"
if [ ! -f go.mod ] || ! grep -q '^module github.com/stvmachine/skills' go.mod; then
    echo "ERROR: run ./scripts/install.sh from the stevmachine-skills repo root." >&2
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
    info "Building stevmachine-skills"
    # Build from the cloned repo.
    # Default install dir: ~/.local/bin. Override with INSTALL_DIR.
    local install_dir="${INSTALL_DIR:-$HOME/.local/bin}"
    mkdir -p "$install_dir"
    go build -o "$install_dir/stevmachine-skills" ./cmd/stevmachine-skills
    ok "stevmachine-skills built at $install_dir/stevmachine-skills"
    case ":$PATH:" in
        *":$install_dir:"*) ;;
        *) warn "$install_dir is not on PATH — add it to your shell rc, or move the binary to a PATH dir." ;;
    esac
}

install_skills() {
    if ! cmd_exists stevmachine-skills; then
        warn "stevmachine-skills not on PATH. Restart your shell, then run: stevmachine-skills install"
        return
    fi
    stevmachine-skills install
}

post_install() {
    cat <<EOF

========================================
  Stevmachine Skills installed
========================================

Next steps:
  1. Run the interactive setup wizard:
       stevmachine-skills env setup

     Or set credentials manually:
       stevmachine-skills env set JIRA_HOST https://yourcompany.atlassian.net
       stevmachine-skills env set JIRA_USERNAME you@example.com
       stevmachine-skills env set JIRA_API_TOKEN <token>

  2. Launch Claude Code with the vault decrypted in-process:
       dotenvx run -f ~/.stevmachine-skills/.env -- claude

  3. Verify with: stevmachine-skills doctor

EOF
}

main() {
    echo "Stevmachine Skills installer"
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

Run from a fresh clone of the stevmachine-skills repo.

Env overrides:
  INSTALL_DIR=<dir>   Where to put the binary (default: ~/.local/bin)
EOF
        exit 0
        ;;
esac

main
