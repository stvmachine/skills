# Installation Guide

## Prerequisites

- [Go](https://go.dev/dl/)
- Node.js and npm (for `dotenvx` and Claude Code)
- Claude Code (`npm install -g @anthropic-ai/claude-code`)
- SSH access to `git@github.com:stvmachine/skills.git` — the repo is private, so `go install …@latest` and `curl | bash` do not work over the public web.

## Install

```bash
# 1. Clone (SSH required — anonymous HTTPS returns 404)
git clone git@github.com:stvmachine/skills.git
cd stevmachine-skills

# 2. Install dotenvx if missing (skip if you already have it)
npm install -g @dotenvx/dotenvx

# 3. Build the CLI onto your PATH
go build -o ~/.local/bin/stevmachine-skills ./cmd/stevmachine-skills
# (or any other PATH dir: /usr/local/bin, /opt/homebrew/bin)

# 4. Install skills + write ~/.claude/.mcp.json
stevmachine-skills install

# 5. Set credentials (interactive wizard)
stevmachine-skills env setup
```

`scripts/install.sh` automates steps 2–5 once you have the repo cloned: `./scripts/install.sh` from the repo root.

## Launching Claude Code

Claude Code must be started with the vault decrypted into its env. See the [README](../README.md#launching-claude-code) for shell aliases and PATH wrappers.

The canonical command:

```bash
dotenvx run -f ~/.stevmachine-skills/.env -- claude
```

## Verify

```bash
stevmachine-skills doctor
./scripts/verify-install.sh
```

## Uninstall

```bash
rm $(which stevmachine-skills)
rm -rf ~/.claude/skills/stevmachine-*
rm -rf ~/.stevmachine-skills   # WARNING: deletes all credentials
```
