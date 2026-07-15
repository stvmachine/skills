# Roadmap: stevmachine-skills

**Core Value:** Steve can own, ship, and update his personal AI assistant skills from one repo without exposing secrets.

| Phase | Status | Plans | Goal |
|-------|--------|-------|------|
| 1 | ○ | 5/5 | Rename and publish as `stevmachine-skills` |
| 2 | ○ | 0/4 | Add OpenCode support |
| 3 | ○ | 0/3 | Tooling improvements (v2) |

## Phase 1: Rename and Publish

**Goal:** Remove all employer branding, rename to `stevmachine-skills`, keep Claude Code behavior intact, and push to GitHub.

**Covers requirements:** RENAME-01 through RENAME-06

**Plans:** 5 plans

- [ ] `01-01-PLAN.md` — String sweep and code rename
- [ ] `01-02-PLAN.md` — Package and asset rename
- [ ] `01-03-PLAN.md` — Build and smoke test
- [ ] `01-04-PLAN.md` — README and documentation rewrite
- [ ] `01-05-PLAN.md` — GitHub publish

**Success criteria:**

- `go build -o stevmachine-skills ./cmd/stevmachine-skills` succeeds.
- `./stevmachine-skills install` copies `stevmachine-*` skills to `~/.claude/skills/`.
- `~/.claude/.mcp.json` is written with `stevmachine-*` server entries.
- `dotenvx` vault under `~/.stevmachine-skills/` works for `env set/list`.
- `medtasker` and `nimblic` no longer appear in source or docs.
- Repo is pushed to GitHub and can be cloned/fresh-installed.

## Phase 2: OpenCode Support

**Goal:** Add OpenCode as a second platform without breaking Claude Code.

**Covers requirements:** OPENCODE-01 through OPENCODE-03

**Plans:**

1. **Platform abstraction**
   - Introduce a `Platform` or `Installer` interface in `internal/`.
   - Refactor `cmd/` to route install/list/doctor through the platform abstraction.
   - Keep Claude Code as the default platform.

2. **OpenCode package layout**
   - Add `packages/opencode-package/` with skills and manifest.
   - Embed OpenCode assets alongside Claude assets.
   - Decide whether to share canonical skill source or duplicate platform-specific packages.

3. **OpenCode install path**
   - Implement install into the OpenCode skills directory (e.g., `~/.agents/skills/`).
   - Merge MCP servers into the OpenCode MCP config file.
   - Keep `dotenvx` vault shared between platforms.

4. **Test and verify**
   - Smoke test OpenCode install on a local OpenCode setup.
   - Ensure Claude Code install still works after the refactor.

**Success criteria:**

- `stevmachine-skills install --platform opencode` installs skills into OpenCode directory.
- Claude Code install still works.
- `dotenvx` credentials are shared and resolve correctly for both platforms.
- No hardcoded platform paths remain in command logic.

## Phase 3: Tooling Improvements (v2)

**Goal:** Harden the distribution tool with dependency resolution, updates, and CI.

**Covers requirements:** TOOL-01 through TOOL-03, PLAT-01, SKILL-01

**Plans:**

1. **Dependency resolver**
   - Parse `dependencies:` from `SKILL.md` frontmatter.
   - Build a directed graph and install in topological order.
   - Report cycles and missing dependencies.

2. **Update command**
   - Add `stevmachine-skills update` to refresh installed skills.
   - Compare embedded skill versions with installed copies.

3. **CI/CD pipeline**
   - GitHub Actions workflow for build, test, and release.
   - Lint and format checks.
   - Release binary artifacts.

**Success criteria:**

- Installing a skill with dependencies auto-installs those dependencies.
- `update` refreshes installed skills without manual reinstall.
- Releases are built automatically on tags.

---
*Roadmap created: 2026-07-15*
