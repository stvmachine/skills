# stevmachine-skills

## What This Is

A personal skill distribution system for AI coding assistants. It packages skills and MCP server configurations into a single Go CLI, installs them into Claude Code (and later OpenCode), and keeps credentials encrypted via `dotenvx`. Owned and maintained by Steve as a personal tool, repurposed from a former employer-specific distribution system.

## Core Value

Steve can own, ship, and update his personal AI assistant skills from one repo without exposing secrets.

## Requirements

### Validated

- ✓ Self-contained Go binary with embedded skill assets — existing
- ✓ Copy skill packages into `~/.claude/skills/` — existing
- ✓ Parse `SKILL.md` YAML frontmatter for MCP servers — existing
- ✓ Merge MCP servers into `~/.claude/.mcp.json` with literal `${VAR}` placeholders — existing
- ✓ Encrypted credential vault via `dotenvx` under `~/.stevmachine-skills/` — existing
- ✓ `install`, `list`, `doctor`, and `env` subcommands — existing

### Active

- [ ] **RENAME-01**: Rename the repository, Go module, binary, vault directory, and user-facing strings from `medtasker` to `stevmachine`.
- [ ] **RENAME-02**: Remove or replace all references to `nimblic`, `Medtasker`, and employer-specific branding in code, docs, scripts, and assets.
- [ ] **RENAME-03**: Update default skill names and package paths from `medtasker-*` to `stevmachine-*`.
- [ ] **RENAME-04**: Keep existing Claude Code installation behavior working end-to-end after the rename.
- [ ] **RENAME-05**: Build, run the smoke test, and verify the renamed CLI still installs skills and writes MCP config correctly.
- [ ] **RENAME-06**: Push the renamed repo to GitHub as `stevmachine/skills` (or `stevmachine-skills` if the org/slug requires it).
- [ ] **OPENCODE-01**: Add OpenCode as a second target platform, installing skills into the appropriate OpenCode skills directory (likely `~/.agents/skills/` or via symlink).
- [ ] **OPENCODE-02**: Add OpenCode-specific package layout and manifest handling without breaking Claude Code support.
- [ ] **OPENCODE-03**: Preserve the `dotenvx` credential model for OpenCode MCP server variables.

## Context

- Repurposed from a private employer-specific skill distribution tool.
- The codebase is a Go CLI with embedded skill packages, a YAML frontmatter parser, an MCP config merger, and a thin `dotenvx` wrapper.
- Only Claude Code is currently implemented; OpenCode packages and the dependency resolver are described in `ROADMAP.md` but not implemented.
- The repo was previously located inside a home-directory git worktree; it has been moved to a standalone repo at `/Users/estvmachine/Projects/personal/stevmachine-skills`.
- Existing code has hardcoded paths to `~/.claude`, `~/.stevmachine-skills`, and `~/.claude.json`, plus scattered `stevmachine` / `Claude Code` strings.

## Constraints

- **Tech stack**: Go single binary with embedded assets and `dotenvx` credential vault — keep the proven architecture.
- **Compatibility**: Existing Claude Code install flow must remain working after the rename.
- **Security**: Credential files and `.env.keys` stay out of git; only literal `${VAR}` placeholders are written to MCP config.
- **Ownership**: No employer branding, names, or proprietary references can remain before the repo is pushed to GitHub.
- **Tooling**: Build and smoke test must pass before any phase is considered complete.
- **Git**: The standalone repo must be initialized and pushed under the new name.
- **Platforms**: Only Claude Code (v1) and OpenCode (v2) are targeted. Other agent platforms are not planned.
- **Dependency resolver**: Skills install in the order given today. Topological ordering is planned for Phase 3.
- **Update command**: Not implemented today; planned for Phase 3.
- **Marketplace**: This is a personal repo, not a public store.
- **New skill content**: The immediate goal is to own and repackage existing skills, not author new ones.
- **Credential model**: `dotenvx` stays; a native Go vault is a future concern only if the Node dependency becomes unacceptable.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Rename to `stevmachine-skills` | Steve owns the code and wants personal branding. | — Pending |
| Stage OpenCode after the rename | Get the existing tool owned and published first, then expand platforms. | — Pending |
| Keep the Go single-binary approach | It is already self-contained and easy to distribute. | — Pending |
| Keep `dotenvx` for credentials | Avoid rewriting the vault; works for both Claude and OpenCode. | — Pending |
| Move project to a standalone repo | Required to push to GitHub as a separate project. | ✓ Good |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Remove or defer with reason.
2. Requirements validated? → Move to Validated with phase reference.
3. New requirements emerged? → Add to Active.
4. Decisions to log? → Add to Key Decisions.
5. "What This Is" still accurate? → Update if drifted.

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections.
2. Core Value check — still the right priority?
3. Audit deferred items — reasons still valid?
4. Update Context with current state.

---
*Last updated: 2026-07-15 after project initialization*
