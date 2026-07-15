# Architecture

**Analysis Date:** 2026-07-15

## System Overview

```text
┌─────────────────────────────────────────────────────────────────────┐
│                     CLI Entry Point (cmd/stevmachine-skills)         │
│  install │ list │ doctor │ env set/list/encrypt/decrypt/rotate/setup │
└────────────────────────────────┬──────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Embedded Skill Packages                          │
│     packages/embed.go //go:embed all:claude-plugin/skills           │
│     (copied at install time to ~/.claude/skills/)                   │
└────────────────────────────────┬──────────────────────────────────────┘
                                 │
         ┌───────────────────────┴───────────────────────┐
         │                                               │
         ▼                                               ▼
┌─────────────────────┐                    ┌─────────────────────┐
│   internal/mcp      │                    │   internal/vault    │
│  Parse SKILL.md     │                    │  dotenvx CLI wrapper│
│  frontmatter, build │                    │  ~/.stevmachine-skills│
│  ~/.claude/.mcp.json│                    │  .env / .env.keys   │
└─────────────────────┘                    └─────────────────────┘
```

`stevmachine-skills` is a single Go binary that distributes Claude Code skills. Skills are embedded with `//go:embed` so the binary is self-contained. Install copies them to `~/.claude/skills/`, parses each skill's `SKILL.md` YAML frontmatter for `mcp_servers`, and merges those servers into `~/.claude/.mcp.json` (and `~/.claude.json`) while keeping `${VAR}` references literal. Credential management is delegated to the `dotenvx` Node.js CLI; the `internal/vault` package is a thin wrapper around it.

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| CLI main | Parse subcommands, dispatch to handlers | `cmd/stevmachine-skills/main.go` |
| Install command | Copy embedded skills, drive TUI, merge MCP configs | `cmd/stevmachine-skills/cmd_install.go` |
| Env command | Set/list/encrypt/decrypt/rotate vault vars, run setup wizard | `cmd/stevmachine-skills/cmd_env.go` |
| List command | Show available vs installed skills | `cmd/stevmachine-skills/cmd_list.go` |
| Doctor command | Check Claude Code dir, dotenvx, vault, optional deps | `cmd/stevmachine-skills/cmd_doctor.go` |
| UI styles | Lipgloss styles and status-line helpers | `cmd/stevmachine-skills/ui.go` |
| MCP manager | Parse SKILL.md frontmatter, build and write `.mcp.json` | `internal/mcp/mcp.go` |
| Vault manager | Wrap `dotenvx` CLI for encrypted env files | `internal/vault/vault.go` |
| Packages | Embed skill assets into the binary | `packages/embed.go` |

## Pattern Overview

**Overall:** Minimalist CLI with embedded assets, external-tool orchestration, and file-based config merging.

**Key Characteristics:**
- Single binary with no runtime dependency on the repo being cloned.
- Skills are embedded via `embed.FS` from `packages/claude-plugin/skills/`.
- MCP config is merged by shallow `map[string]any` overwrite into the user's existing `~/.claude/.mcp.json`.
- Secrets are never stored in `.mcp.json`; only `${VAR}` placeholders are written.
- All credential operations shell out to the `dotenvx` CLI.

## Layers

**CLI Layer:**
- Purpose: Dispatch commands and render output.
- Location: `cmd/stevmachine-skills/`
- Contains: Subcommand handlers, Bubble Tea TUI for install, `huh` form for env setup.
- Depends on: `internal/mcp`, `internal/vault`, `packages`.
- Used by: End user.

**Business Logic Layer:**
- Purpose: Skill frontmatter parsing and MCP config assembly.
- Location: `internal/mcp/`
- Contains: `ServerConfig`, `SkillFrontmatter`, `BuildMcpServers`, `WriteMcpConfig`.
- Depends on: `gopkg.in/yaml.v3`, `encoding/json`, `io/fs`.
- Used by: `cmd_install.go`, `cmd_list.go`.

**Infrastructure Layer:**
- Purpose: Encrypted vault operations.
- Location: `internal/vault/`
- Contains: `Manager`, `findDotenvx`, `runDotenvx`.
- Depends on: `os/exec` for `dotenvx` CLI.
- Used by: `cmd_env.go`, `cmd_doctor.go`.

**Asset Layer:**
- Purpose: Provide static skill packages at runtime.
- Location: `packages/embed.go`
- Contains: `//go:embed all:claude-plugin/skills`.
- Used by: `cmd_install.go`, `cmd_list.go`.

## Data Flow

### Primary Install Path

1. `main.go` dispatches `install` to `cmdInstall()` (`cmd/stevmachine-skills/cmd_install.go:134`).
2. Compute target dirs: `~/.claude/skills/` and MCP config paths `~/.claude/.mcp.json` plus `~/.claude.json` (`cmd_install.go:136-140`).
3. If no args are provided, use `defaultSkills()` (`cmd_install.go:204-206`).
4. For each skill, `doInstallOne()` reads the embedded directory from `packages.SkillsFS` (`cmd_install.go:208-230`).
5. `copyFS()` writes files to the user's filesystem (`cmd_install.go:232-255`).
6. `mcp.ParseSkillMcpConfig()` reads `SKILL.md` frontmatter (`internal/mcp/mcp.go:62-74`).
7. `mcp.BuildMcpServers()` converts YAML structs to the JSON shape Claude Code expects (`internal/mcp/mcp.go:84-114`).
8. `mcp.WriteMcpConfig()` merges the new servers into existing `.mcp.json` by shallow map overwrite (`internal/mcp/mcp.go:116-137`).
9. The TUI (`installModel`) reports results; on non-TTY it falls back to plain output (`cmd_install.go:162-193`).

### Env Command Path

1. `main.go` dispatches `env <sub>` to `cmdEnvSet/List/Encrypt/Decrypt/Rotate/Setup` (`cmd/stevmachine-skills/main.go:40-55`).
2. `vault.New()` creates a manager whose `EnvDir` is `~/.stevmachine-skills` (`internal/vault/vault.go:16-21`).
3. Each operation calls `runDotenvx()` which finds the `dotenvx` binary and runs it with `cmd.Dir = EnvDir` (`internal/vault/vault.go:71-79`).
4. `cmdEnvSetup()` builds a `huh` form with hardcoded integrations and stores values via `vault.Manager.Set()` (`cmd/stevmachine-skills/cmd_env.go:123-226`).

### List Command Path

1. `cmdList()` reads installed directories from `~/.claude/skills/` (`cmd/stevmachine-skills/cmd_list.go:14-25`).
2. It iterates embedded `claude-plugin/skills/` and parses each `SKILL.md` for descriptions (`cmd_list.go:34-45`).
3. Prints a table with install status indicators.

## Key Abstractions

**SkillFrontmatter / ServerConfig:**
- Purpose: Represent the YAML frontmatter in `SKILL.md` and the MCP server entries.
- Examples: `internal/mcp/mcp.go:13-27`.
- Pattern: Plain structs with `yaml` tags; no validation beyond "name must be present".

**Manager (Vault):**
- Purpose: Encapsulate all dotenvx interactions.
- Examples: `internal/vault/vault.go:12-15`.
- Pattern: Struct with `EnvDir`, methods per dotenvx subcommand.

## Entry Points

**Binary:**
- Location: `cmd/stevmachine-skills/main.go`
- Triggers: Direct invocation, `scripts/install.sh`, `go run ./cmd/stevmachine-skills`.
- Responsibilities: Parse `os.Args`, dispatch subcommands, print usage.

**Install Script:**
- Location: `scripts/install.sh`
- Triggers: Run from a fresh clone.
- Responsibilities: Install `dotenvx`, build the binary, run `stevmachine-skills install`, print next steps.

## Architectural Constraints

- **Single-platform support:** Only Claude Code is implemented. The repo embeds `packages/claude-plugin/skills`; `packages/opencode-package/` from early design drafts does not exist.
- **Hardcoded target directories:** `~/.claude`, `~/.stevmachine-skills`, and `~/.claude.json` are literals in source.
- **External CLI dependency:** `internal/vault` shells out to `dotenvx` and fails if it is missing or not on PATH.
- **No dependency graph:** `ROADMAP.md` plans a dependency resolver for Phase 3, but the current CLI installs skills in the order given without resolving `dependencies:` from frontmatter.
- **No config validation:** `.mcp.json` write path uses `map[string]any` and JSON marshal/unmarshal; malformed existing configs are silently overwritten.
- **TUI fallback relies on error:** `tea.NewProgram(...).Run()` error is used to detect non-TTY, which means piped output may contain ANSI escape sequences before the fallback runs.

## Anti-Patterns

### Hardcoded User Directories in Multiple Files

**What happens:** `~/.claude`, `~/.claude/skills`, `~/.stevmachine-skills`, and `~/.claude.json` are constructed inline in `cmd_install.go`, `cmd_list.go`, `cmd_doctor.go`, and `cmd_env.go`.
**Why it's wrong:** Renaming the product, supporting a second platform (OpenCode), or allowing custom install roots requires editing many files and risks inconsistency.
**Do this instead:** Centralize all path constants in a `pkg/paths` or `internal/platform` package and inject a platform abstraction into the commands.

### Shallow MCP Config Merge

**What happens:** `WriteMcpConfig` merges top-level `mcpServers` entries by overwriting the entire server map (`internal/mcp/mcp.go:129-131`).
**Why it's wrong:** If two skills define the same server with different `env` or `args`, the last one wins unconditionally, wiping the earlier config without warning.
**Do this instead:** Deep-merge server entries (env, args, headers) or report conflicts and require user resolution.

### Missing Directory Creation in Vault Set

**What happens:** `vault.Manager.Set` runs `dotenvx` with `cmd.Dir = EnvDir` but never creates the directory (`internal/vault/vault.go:105-112`).
**Why it's wrong:** On a fresh machine `stevmachine-skills env set` fails with `chdir … no such file or directory`.
**Do this instead:** Ensure `EnvDir` exists in `Set`, `Get`, and `ListVars`, or call `InitVault` lazily.

## Error Handling

**Strategy:** Print to `os.Stderr` and `os.Exit(1)` for fatal errors; return `error` from internal packages.

**Patterns:**
- Commands swallow non-fatal errors (e.g., `cmdInstall` ignores `mcp.WriteMcpConfig` return values).
- `internal/vault` wraps `dotenvx` combined output in the error message.
- Tests skip when `dotenvx` is unavailable.

## Cross-Cutting Concerns

**Logging:** No structured logging; only `fmt.Println` / `fmt.Fprintln` output.
**Validation:** Input validation is minimal; `huh` form provides basic UI constraints, but no backend contract validation exists.
**Authentication:** Handled entirely by `dotenvx` and the user's shell wrapper; the Go code never touches plaintext secrets.

---

*Architecture analysis: 2026-07-15*
