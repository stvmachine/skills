# Requirements: stevmachine-skills

**Defined:** 2026-07-15
**Core Value:** Steve can own, ship, and update his personal AI assistant skills from one repo without exposing secrets.

## v1 Requirements

### Rename & Ownership

- [ ] **RENAME-01**: Rename the repository, Go module, binary, vault directory, and user-facing strings from `medtasker` to `stevmachine`.
- [ ] **RENAME-02**: Remove or replace all references to `nimblic`, `Medtasker`, and employer-specific branding in code, docs, scripts, and assets.
- [ ] **RENAME-03**: Update default skill names and package paths from `medtasker-*` to `stevmachine-*`.
- [ ] **RENAME-04**: Keep existing Claude Code installation behavior working end-to-end after the rename.
- [ ] **RENAME-05**: Build, run the smoke test, and verify the renamed CLI still installs skills and writes MCP config correctly.
- [ ] **RENAME-06**: Push the renamed repo to GitHub as `stevmachine/skills` (or `stevmachine-skills` if the org/slug requires it).

### OpenCode Support

- [ ] **OPENCODE-01**: Add OpenCode as a second target platform, installing skills into the appropriate OpenCode skills directory (likely `~/.agents/skills/` or via symlink).
- [ ] **OPENCODE-02**: Add OpenCode-specific package layout and manifest handling without breaking Claude Code support.
- [ ] **OPENCODE-03**: Preserve the `dotenvx` credential model for OpenCode MCP server variables.

## v2 Requirements

### Tooling

- **TOOL-01**: Implement a dependency resolver so installing a skill automatically installs its declared dependencies in topological order.
- **TOOL-02**: Add an `update` command to refresh installed skills from the latest repo state.
- **TOOL-03**: Add CI/CD pipeline (GitHub Actions) for build, test, and release.

### Skills

- **SKILL-01**: Author new personal skills beyond the existing repurposed set.

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| RENAME-01 | Phase 1 | Pending |
| RENAME-02 | Phase 1 | Pending |
| RENAME-03 | Phase 1 | Pending |
| RENAME-04 | Phase 1 | Pending |
| RENAME-05 | Phase 1 | Pending |
| RENAME-06 | Phase 1 | Pending |
| OPENCODE-01 | Phase 2 | Pending |
| OPENCODE-02 | Phase 2 | Pending |
| OPENCODE-03 | Phase 2 | Pending |

**Coverage:**
- v1 requirements: 9 total
- Mapped to phases: 9
- Unmapped: 0 ✓

---
*Requirements defined: 2026-07-15*
*Last updated: 2026-07-15*
