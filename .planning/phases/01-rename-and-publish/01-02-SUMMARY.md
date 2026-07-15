---
phase: 01-rename-and-publish
plan: 01-02
subsystem: packaging
tags: [claude-plugin, skills, rename, plugin]

requires:
  - phase: 01-rename-and-publish
    provides: 01-01
provides:
  - Renamed skill directories under packages/claude-plugin/skills/
  - Updated skill frontmatter, invocation paths, and env vars
  - Updated plugin.json manifest
  - Updated embedded smoke.sh and wizard.sh test scripts
affects:
  - 01-03
  - 01-04

tech-stack:
  added: []
  patterns: []

key-files:
  created: []
  modified:
    - packages/claude-plugin/.claude-plugin/plugin.json
    - packages/claude-plugin/skills/stevmachine-jira/SKILL.md
    - packages/claude-plugin/skills/stevmachine-jira/rules/ticket-storage.md
    - packages/claude-plugin/skills/stevmachine-jira-markup/SKILL.md
    - packages/claude-plugin/skills/stevmachine-jira-ticket-transition/SKILL.md
    - packages/claude-plugin/skills/run-stevmachine-skills/SKILL.md
    - packages/claude-plugin/skills/run-stevmachine-skills/smoke.sh
    - packages/claude-plugin/skills/run-stevmachine-skills/wizard.sh

key-decisions:
  - "Used generic placeholders (your-project, example.atlassian.net, github.com/your-org/your-repo) for employer-specific references"

patterns-established:

requirements-completed:
  - RENAME-01
  - RENAME-02
  - RENAME-03

coverage:
  - id: D1
    description: "Embedded skill directories renamed to stevmachine-* and run-stevmachine-skills"
    requirement: RENAME-01
    verification:
      - kind: other
        ref: "ls packages/claude-plugin/skills/"
        status: pass
    human_judgment: false
  - id: D2
    description: "Skill frontmatter, invocation paths, and environment variables renamed"
    requirement: RENAME-02
    verification:
      - kind: other
        ref: "grep -R 'medtasker|nimblic' --include='*.md' packages/claude-plugin/skills/"
        status: pass
    human_judgment: false
  - id: D3
    description: "Plugin manifest and embedded test scripts updated"
    requirement: RENAME-03
    verification:
      - kind: other
        ref: "grep -R 'medtasker|nimblic' --include='*.json' --include='*.sh' packages/claude-plugin/"
        status: pass
    human_judgment: false

duration: 20min
completed: 2026-07-15
status: complete
---

# Plan 01-02: Package and Asset Rename Summary

**Renamed the embedded Claude Code skill packages from medtasker-* to stevmachine-*, updated their frontmatter and internal references, and aligned the plugin manifest and test scripts.**

## Performance

- **Duration:** 20 min
- **Started:** 2026-07-15T00:00:00Z
- **Completed:** 2026-07-15T00:00:00Z
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments
- Renamed four skill directories under `packages/claude-plugin/skills/`
- Updated skill `name:` frontmatter to match new directory basenames
- Replaced skill invocation paths `/medtasker-*` with `/stevmachine-*`
- Renamed `MEDTASKER_*` environment variables to `STEVMACHINE_*`
- Replaced employer-specific references with generic placeholders
- Updated `plugin.json` name, description, and skills list
- Updated `smoke.sh` and `wizard.sh` to build and verify `stevmachine-skills`

## Task Commits

Each task was committed atomically:

1. **Task 1: Rename skill directories** - `bf34e66` (feat)
2. **Task 2: Update skill frontmatter and internal references** - `e8e9888` (feat)
3. **Task 3: Update plugin manifest and embedded test scripts** - `b95bb44` (feat)

**Plan metadata:** `bf34e66`

## Files Created/Modified
- `packages/claude-plugin/skills/stevmachine-jira/` - Renamed from medtasker-jira
- `packages/claude-plugin/skills/stevmachine-jira-markup/` - Renamed from medtasker-jira-markup
- `packages/claude-plugin/skills/stevmachine-jira-ticket-transition/` - Renamed from medtasker-jira-ticket-transition
- `packages/claude-plugin/skills/run-stevmachine-skills/` - Renamed from run-medtasker-skills
- `packages/claude-plugin/.claude-plugin/plugin.json` - Updated manifest
- `packages/claude-plugin/skills/run-stevmachine-skills/smoke.sh` - Updated smoke test
- `packages/claude-plugin/skills/run-stevmachine-skills/wizard.sh` - Updated wizard driver

## Decisions Made
- Used generic placeholders for employer-specific references (`your-project`, `example.atlassian.net`, `github.com/your-org/your-repo`) to keep the skills reusable.

## Deviations from Plan

None - plan executed exactly as written, with the module path adjusted to `github.com/stvmachine/skills` to match the user's GitHub account.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Skill package rename complete. Ready for Plan 01-03 (build and smoke test) and Plan 01-04 (documentation rewrite).
- No blockers.

---
*Phase: 01-rename-and-publish*
*Completed: 2026-07-15*
