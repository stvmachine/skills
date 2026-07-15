# Stevmachine Skills — Design Document

> **Scope:** This document describes the *current* architecture of the `stevmachine-skills` distribution tool. It is intentionally trimmed to what exists today: a Go CLI that installs skills into Claude Code and manages encrypted credentials via `dotenvx`.

## 1. Overview

`stevmachine-skills` is a single Go binary that distributes a small set of personal AI coding skills to Claude Code. It does three things:

1. **Copies skill packages** into `~/.claude/skills/`.
2. **Merges MCP server configurations** from each skill into `~/.claude/.mcp.json`, keeping environment variables as literal `${VAR}` placeholders.
3. **Manages an encrypted credential vault** at `~/.stevmachine-skills/` using `dotenvx`.

Skills are embedded into the binary with `//go:embed`, so the tool works standalone once built.

## 2. Repository Structure

```
stevmachine-skills/
├── cmd/stevmachine-skills/        # CLI entry point and subcommands
│   ├── main.go                    # Command routing
│   ├── cmd_install.go             # TUI + fallback install flow
│   ├── cmd_list.go                # List installed vs embedded skills
│   ├── cmd_env.go                 # Vault read/write commands
│   ├── cmd_doctor.go              # Health check
│   └── ui.go                      # Shared lipgloss styles
├── internal/
│   ├── vault/vault.go             # dotenvx wrapper
│   └── mcp/mcp.go                 # SKILL.md frontmatter parser + .mcp.json writer
├── packages/
│   ├── embed.go                   # //go:embed all:claude-plugin/skills
│   └── claude-plugin/
│       ├── .claude-plugin/plugin.json
│       └── skills/                # Embedded skill packages
│           ├── stevmachine-jira/
│           ├── stevmachine-jira-markup/
│           ├── stevmachine-jira-ticket-transition/
│           ├── commit/
│           └── run-stevmachine-skills/
├── scripts/
│   ├── install.sh                 # Clone + build helper
│   └── verify-install.sh          # Fresh-clone verification
├── docs/                          # ADRs and installation guides
├── README.md
├── DESIGN.md                      # This document
├── go.mod
└── go.sum
```

## 3. Architecture

### 3.1 CLI (`cmd/stevmachine-skills`)

A plain `flag`-free subcommand dispatcher using `os.Args`:

```bash
stevmachine-skills install [skill ...]   # default: all embedded skills
stevmachine-skills list
stevmachine-skills doctor
stevmachine-skills env set KEY VALUE
stevmachine-skills env list
stevmachine-skills env encrypt|decrypt|rotate|setup
```

- `install` uses a Bubble Tea spinner TUI; if the TTY is unavailable it falls back to plain text output.
- `env` is a thin wrapper around the `dotenvx` CLI for encrypt/decrypt/set/list/rotate operations.
- `doctor` reports the state of Claude Code, `dotenvx`, the vault, `~/.claude/.mcp.json`, and optional tools (`beads`, `rtk`).

### 3.2 Skill Packages (`packages/`)

Skills are embedded from `packages/claude-plugin/skills` via `//go:embed all:claude-plugin/skills`. The `all:` prefix preserves hidden files (e.g., `.mcp.json` inside skill directories).

A skill directory contains:

- `SKILL.md` — skill instructions plus a YAML frontmatter block.
- Optional `rules/*.md` files for skill-specific conventions.
- Optional shell scripts or sample files for skill demos.

`packages/claude-plugin/.claude-plugin/plugin.json` lists the bundled skills and is the Claude Code plugin manifest.

### 3.3 MCP Config (`internal/mcp`)

Each `SKILL.md` frontmatter declares the MCP servers it needs:

```yaml
---
name: stevmachine-jira
description: Connect code work to Jira tickets
mcp_servers:
  - name: mcp-atlassian
    type: stdio
    command: uvx
    args: [mcp-atlassian, --transport, stdio]
    env:
      JIRA_URL: ${JIRA_URL}
      JIRA_USERNAME: ${JIRA_USERNAME}
      JIRA_API_TOKEN: ${JIRA_API_TOKEN}
---
```

On `install`, the tool:

1. Parses the frontmatter.
2. Copies the skill to `~/.claude/skills/<name>`.
3. Builds a `map[string]map[string]any` of MCP servers via `mcp.BuildMcpServers`.
4. Merges those servers into `~/.claude/.mcp.json` (and `~/.claude.json` for compatibility) without resolving the variables.

`WriteMcpConfig` preserves existing entries in the target file, overwriting only servers with the same name.

### 3.4 Vault (`internal/vault`)

The vault is stored under `~/.stevmachine-skills/`:

- `.env` — encrypted ciphertext.
- `.env.keys` — decryption key (local-only, never committed).
- Optional `.env.<environment>` files for per-environment vaults.

`vault.Manager` shells out to `dotenvx` for all operations:

- `Set`, `Get`, `ListVars` — read/write variables.
- `Encrypt`, `Decrypt`, `Rotate` — manage the ciphertext and keys.
- `DestroyPlaintext` — securely shreds a decrypted `.env` file.

The directory is created with `0o700`, and `.env` files are created with `0o600`.

## 4. Installation Flow

The repo is public and the distribution model is **clone + build**, not a remote `curl | bash` or `go install` flow:

```bash
git clone git@github.com:stvmachine/skills.git
cd stevmachine-skills
go build -o ~/.local/bin/stevmachine-skills ./cmd/stevmachine-skills
stevmachine-skills install
stevmachine-skills env setup
```

Optional tools are suggested but not required:

- `beads` (`bd`) — enables bead-based Jira ticket storage; otherwise the `stevmachine-jira` skill falls back to a local `./.todo/` directory.
- `rtk` — token-optimized wrappers for Claude Code commands.

## 5. Security Model

1. **No secrets in the repo.** The binary embeds only skill instructions and public manifests.
2. **No resolved secrets in MCP config.** `~/.claude/.mcp.json` stores only `${VAR}` placeholders; Claude Code resolves them from its own process environment at MCP server startup.
3. **Encrypted vault.** Credentials are encrypted with `dotenvx` (ECIES + AES-256-GCM) and decrypted only when launching Claude Code via `dotenvx run`.
4. **Least privilege.** The tool writes only to user directories (`~/.claude`, `~/.stevmachine-skills`).
5. **Safe plaintext lifecycle.** `env decrypt` warns the user to re-encrypt before leaving the shell; `DestroyPlaintext` overwrites the decrypted file before deletion.

## 6. Roadmap Alignment

The project roadmap is defined in `.planning/ROADMAP.md`. This design document reflects the current state after Phase 1 and the starting point for Phase 2.

| Phase | Status | Focus |
|-------|--------|-------|
| 1 | Complete | Rename and publish as `stevmachine-skills` |
| 2 | In progress | Add OpenCode support |
| 3 | Planned | Tooling improvements: dependency resolver, updates, CI |

### Phase 2 — OpenCode Support (planned)

The current implementation is hardcoded to Claude Code. Phase 2 will introduce a platform abstraction so the same embedded skills can be installed into OpenCode's skill directory and merged into OpenCode's MCP config file while preserving the `dotenvx` vault.

### Phase 3 — Tooling Improvements (planned)

- Parse `dependencies:` from `SKILL.md` frontmatter and install skills in topological order.
- Add an `update` command to refresh installed skills from the embedded source.
- Add GitHub Actions for build, test, and release artifacts.


## 7. Appendix

### A.1 Skill Manifest (frontmatter)

```yaml
---
name: example-skill
description: Short description for list output
mcp_servers:
  - name: mcp-example
    type: stdio
    command: npx
    args: [-y, "@modelcontextprotocol/server-example"]
    env:
      EXAMPLE_TOKEN: ${EXAMPLE_TOKEN}
---
```

### A.2 Environment Variables

| Variable | Purpose | Required by |
|----------|---------|-------------|
| `JIRA_URL` | Jira base URL | `stevmachine-jira` |
| `JIRA_USERNAME` | Jira user email | `stevmachine-jira` |
| `JIRA_API_TOKEN` | Jira API token | `stevmachine-jira` |
| `CONFLUENCE_URL` | Confluence wiki URL | `stevmachine-jira` |
| `CONFLUENCE_USERNAME` | Confluence user email | `stevmachine-jira` |
| `CONFLUENCE_API_TOKEN` | Confluence API token | `stevmachine-jira` |
| `GITHUB_TOKEN` | GitHub PAT | `stevmachine-jira` |
| `FIGMA_API_KEY` | Figma API token | `stevmachine-jira` |
| `CONTEXT7_API_KEY` | Context7 API key | `stevmachine-jira` |
| `STEVMACHINE_TICKET_DIR` | Local ticket fallback directory | `stevmachine-jira` |
| `STEVMACHINE_VAULT` | Alternative vault `.env` path | `env` commands |

---

*Document version: 2.0.0*  
*Last updated: 2026-07-15*  
*Source of truth: `.planning/ROADMAP.md`, `.planning/PROJECT.md`, and the Go source tree.*
