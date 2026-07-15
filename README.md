# Medtasker Skills

A minimal distribution system for Medtasker [skills](https://docs.claude.com/en/docs/claude-code/skills) and [MCP servers](https://modelcontextprotocol.io) in Claude Code, with credentials backed by [dotenvx](https://github.com/dotenvx/dotenvx).

![demo](docs/assets/demo.gif)

The whole tool does three things:

1. Copies skill packages to `~/.claude/skills/`.
2. Merges each skill's MCP server config into `~/.claude/.mcp.json` — keeping `${VAR}` references literal, not resolved values.
3. Manages an encrypted `.env` at `~/.medtasker-skills/` (with keys in `.env.keys`) that supplies those vars when you launch Claude Code under `dotenvx run`.

## Quick Start

![setup flow](docs/assets/setup-flow.png)

> **Note:** this repo is private. The install path is **clone + build**, not the public `curl | bash` or `go install` flows you may have seen on other tools. Anonymous HTTPS access to `github.com/nimblic/medtasker-skills` returns 404, so `go install …@latest` and a raw-content `curl` both fail before they start.

Prereqs: [Go](https://go.dev/dl/), Node.js (for `dotenvx`), Claude Code (`npm install -g @anthropic-ai/claude-code`), and SSH access to the nimblic GitHub org.

```bash
git clone git@github.com:nimblic/medtasker-skills.git
cd medtasker-skills

# Build the CLI into a directory on your PATH
go build -o ~/.local/bin/medtasker-skills ./cmd/medtasker-skills
# (or anywhere on PATH: /usr/local/bin, /opt/homebrew/bin, etc.)

# Copy skill packages to ~/.claude/skills/ + write ~/.claude/.mcp.json
medtasker-skills install

# Interactive credential setup
medtasker-skills env setup
```

If you'd rather not put the binary on PATH, every `medtasker-skills <cmd>` line below is equivalent to `go run ./cmd/medtasker-skills <cmd>` from inside the repo.

## CLI

```bash
medtasker-skills install [SKILL ...]   # copy skills + write .mcp.json (defaults to all)
medtasker-skills list                  # list installed skills
medtasker-skills doctor                # check Claude Code / dotenvx / vault state

medtasker-skills env set KEY VALUE     # store an encrypted variable
medtasker-skills env list              # show variables (masked)
medtasker-skills env encrypt           # encrypt .env
medtasker-skills env decrypt           # decrypt .env (don't leave it sitting)
medtasker-skills env rotate            # rotate the encryption key
medtasker-skills env setup             # interactive TUI wizxxard

# All env commands take --environment/-e for per-env files (e.g. .env.production)
```

## Credential Management

Secrets live encrypted in `~/.medtasker-skills/.env` (ciphertext) with decryption keys in `~/.medtasker-skills/.env.keys`. This is the sole store — no OS keychain, no plaintext fallback. See [docs/ADR-001-dotenvx.md](docs/ADR-001-dotenvx.md) for the rationale.

Use the interactive wizard for first-time setup:

```bash
medtasker-skills env setup
```

Or set variables manually:

```bash
medtasker-skills env set JIRA_HOST https://yourcompany.atlassian.net
medtasker-skills env set JIRA_USERNAME you@example.com
medtasker-skills env set JIRA_API_TOKEN <token>
```

### Supported MCP Servers

| Server | Required Env Vars |
|---|---|
| Jira (mcp-atlassian) | `JIRA_HOST`, `JIRA_USERNAME`, `JIRA_API_TOKEN` |
| GitHub | `GITHUB_TOKEN` |
| Confluence | `CONFLUENCE_URL`, `CONFLUENCE_USERNAME`, `CONFLUENCE_API_TOKEN` |
| Figma | `FIGMA_ACCESS_TOKEN` |

### Generated `.mcp.json`

The generated `~/.claude/.mcp.json` stores only `${VAR}` references. Claude Code expands them from its process env at MCP server startup.


```json
{
  "mcpServers": {
    "mcp-atlassian": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-atlassian"],
      "env": {
        "JIRA_HOST": "${JIRA_HOST}",
        "JIRA_USERNAME": "${JIRA_USERNAME}",
        "JIRA_API_TOKEN": "${JIRA_API_TOKEN}"
      }
    }
  }
}
```

## Launching Claude Code

`claude` must be started with the vault's variables already in its env. The canonical command:

```bash
dotenvx run -f ~/.medtasker-skills/.env -- claude
```

`dotenvx run` decrypts the vault in memory using `~/.medtasker-skills/.env.keys`, injects values into the `claude` child process env, and exits. No plaintext is ever written to disk.

### Wrapping `claude`

Typing the full command every time is tedious. Pick one.

**Shell function — `fish`** (`~/.config/fish/functions/claude.fish`):

```fish
function claude --description 'Run claude with medtasker vault decrypted'
    dotenvx run -f $HOME/.medtasker-skills/.env -- command claude $argv
end
```

**Shell function — `bash` / `zsh`** (in `~/.bashrc` or `~/.zshrc`):

```bash
claude() {
  dotenvx run -f "$HOME/.medtasker-skills/.env" -- command claude "$@"
}
```

`command claude` bypasses the function recursion and calls the real binary. Reload with `exec $SHELL -l`.

**PATH wrapper** (works for GUI / IDE launches, not just terminal). Save as `~/.local/bin/claude`, `chmod +x`, and ensure `~/.local/bin` is ahead of the real binary's directory in `PATH`:

```bash
#!/usr/bin/env bash
set -euo pipefail
REAL_CLAUDE="$(PATH="${PATH#*$HOME/.local/bin:}" command -v claude)"
exec dotenvx run -f "$HOME/.medtasker-skills/.env" -- "$REAL_CLAUDE" "$@"
```

**Per-environment vaults**: parameterize the wrapper with `MEDTASKER_VAULT`:

```bash
claude() {
  local vault="${MEDTASKER_VAULT:-$HOME/.medtasker-skills/.env}"
  dotenvx run -f "$vault" -- command claude "$@"
}
```

Then `MEDTASKER_VAULT=~/.medtasker-skills/.env.production claude`.

## Available Skills

| Skill | Description |
|---|---|
| `medtasker-jira` | Jira ticket workflow via mcp-atlassian |
| `medtasker-jira-markup` | Jira comment formatting helper |
| `medtasker-jira-ticket-transition` | Advance a ticket — `qa` (ship to QA) or `review` (send for peer review). PR + Jira transition + bead update. |
| `commit` | Conventional commits with gitmoji |
| `run-medtasker-skills` | Build, run, and smoke-test this CLI (developer tool — drives the `env setup` wizard via tmux) |

## Architecture

Three Go packages:

- `internal/vault` — wraps the `dotenvx` CLI.
- `internal/mcp` — parses skill `SKILL.md` frontmatter and writes `~/.claude/.mcp.json`.
- `cmd/medtasker-skills` — CLI entry point with `flag`-based subcommands.

Skill packages are embedded into the binary via `//go:embed` so `medtasker-skills install` works without the repo cloned.

## Security

- ECIES (secp256k1 + AES-256-GCM) via dotenvx.
- `.env` is encrypted and safe to commit. `.env.keys` is local-only.
- `~/.medtasker-skills/` has `chmod 700`.
- `.mcp.json` never contains resolved secrets — only `${VAR}` references.

## Development

```bash
go build ./cmd/medtasker-skills
go test ./...
```

## License

MIT
