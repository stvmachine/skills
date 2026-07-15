# Phase 2: OpenCode Support - Context

**Gathered:** 2026-07-15
**Status:** Ready for planning
**Source:** Clarified user intent during `/gsd-plan-phase 2`

## Phase Boundary

Add OpenCode as a second supported platform for `stevmachine-skills` **without** building OpenCode-specific plugin infrastructure. The CLI remains focused on:

1. **MCP server configuration** for OpenCode (via secrets / `dotenvx` vault).
2. **Sharable skills** — skills can be installed and consumed by OpenCode users.
3. **No OpenCode plugin/client config** — OpenCode plugins, UI settings, and client-side tech decisions are not part of this phase.

Claude Code behavior must remain the default and must not break.

## Implementation Decisions

### In Scope
- Add a `--platform opencode` flag to the install command (Claude Code remains the default).
- Install skills into the OpenCode skills directory (e.g., `~/.opencode/skills/` or equivalent).
- Write/merge MCP server entries into OpenCode's MCP config file (e.g., `~/.opencode/mcp.json` or equivalent) so skills register as MCP servers.
- Keep the `dotenvx` vault shared between platforms (`~/.stevmachine-skills/`).
- Ensure skill packages are platform-agnostic and "sharable" — the same embedded skill source works for both Claude Code and OpenCode.
- Refactor any hardcoded Claude Code paths into a platform-aware abstraction so the CLI can route install/list/doctor based on the target platform.

### Not in This Phase
- OpenCode plugin system or client plugin configuration.
- OpenCode UI themes, custom agents, or client-specific settings.
- OpenCode account/auth configuration beyond MCP server registration.
- Renaming or duplicating skills for OpenCode branding.

### Platform Default
- Claude Code remains the default platform when no `--platform` flag is provided.
- Existing `~/.claude/skills/` and `~/.claude/.mcp.json` behavior must continue to work unchanged.

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project docs and decisions
- `.planning/ROADMAP.md` — Phase 2 goal and success criteria.
- `.planning/STATE.md` — Current project status and constraints.
- `.planning/REQUIREMENTS.md` — Requirements OPENCODE-01 through OPENCODE-03.
- `DESIGN.md` — High-level design of the skills distribution tool.

### Existing code to inspect
- `cmd/stevmachine-skills/` — CLI command handlers and install logic.
- `internal/mcp/mcp.go` — MCP config parsing and writing.
- `internal/vault/vault.go` — `dotenvx` vault directory and secrets handling.
- `packages/embed.go` — embedded skill packages.
- `packages/claude-plugin/skills/` — current Claude Code skill packages and `plugin.json`.
- `scripts/install.sh`, `scripts/verify-install.sh` — current install and verification scripts.
- `packages/claude-plugin/skills/run-stevmachine-skills/smoke.sh` — smoke test expectations.

## Specific Ideas

- OpenCode's MCP config path is likely `~/.opencode/mcp.json`, but this must be confirmed by research.
- OpenCode's skills directory is likely `~/.opencode/skills/` or `~/.agents/skills/`; confirm by research.
- The same `SKILL.md` + embedded source package should work for both platforms; avoid duplicating skill source.
- The platform abstraction should be small: a `Platform` interface with `SkillsDir()`, `MCPConfigPath()`, and `WriteMCPConfig(servers)` methods.
- Verification should include a smoke test for `--platform opencode` install and a regression test for default Claude Code install.

## Deferred Ideas

- Full OpenCode plugin marketplace integration.
- OpenCode-specific skill UI or manifest extensions.
- Auto-detection of the running AI client (Claude Code vs OpenCode) — could be useful later, but not required now.

---

*Phase: 02-opencode-support*
*Context gathered: 2026-07-15 via clarified user intent*
