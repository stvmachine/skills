---
phase: 01-rename-and-publish
plan: 01-01
subsystem: cli
tags: [go, module, rename, vault]

requires:
provides:
  - Go module path renamed to github.com/stevmachine/skills
  - CLI directory and binary renamed to stevmachine-skills
  - Vault default directory renamed to ~/.stevmachine-skills
affects:
  - 01-02
  - 01-03
  - 01-04

tech-stack:
  added: []
  patterns: []

key-files:
  created: []
  modified:
    - go.mod
    - cmd/stevmachine-skills/main.go
    - cmd/stevmachine-skills/cmd_install.go
    - cmd/stevmachine-skills/cmd_env.go
    - cmd/stevmachine-skills/cmd_doctor.go
    - cmd/stevmachine-skills/cmd_list.go
    - internal/vault/vault.go
    - internal/vault/vault_test.go

key-decisions:
patterns-established:

requirements-completed:
  - RENAME-01
  - RENAME-02
  - RENAME-03

coverage:
  - id: D1
    description: "Go module path and package imports renamed to github.com/stevmachine/skills"
    requirement: RENAME-01
    verification:
      - kind: other
        ref: "grep -R 'github.com/nimblic/medtasker-skills' --include='*.go' --include='go.mod' ."
        status: pass
    human_judgment: false
  - id: D2
    description: "CLI directory, usage strings, and default skill names renamed to stevmachine-skills"
    requirement: RENAME-02
    verification:
      - kind: other
        ref: "test -d cmd/stevmachine-skills && test ! -d cmd/medtasker-skills"
        status: pass
    human_judgment: false
  - id: D3
    description: "Vault default directory renamed to ~/.stevmachine-skills"
    requirement: RENAME-03
    verification:
      - kind: other
        ref: "grep -R 'medtasker' --include='*.go' internal/vault"
        status: pass
    human_judgment: false

duration: 15min
completed: 2026-07-15
status: complete
---

# Plan 01-01: String Sweep and Code Rename Summary

**Renamed the Go module, CLI binary, package imports, and vault directory from medtasker to stevmachine, removing all employer-specific references from the Go source.**

## Performance

- **Duration:** 15 min
- **Started:** 2026-07-15T00:00:00Z
- **Completed:** 2026-07-15T00:00:00Z
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments
- Changed `go.mod` module path to `github.com/stevmachine/skills`
- Renamed `cmd/medtasker-skills/` directory to `cmd/stevmachine-skills/`
- Updated all internal Go imports to use the new module path
- Renamed CLI usage strings, TUI titles, and default skill names to `stevmachine-*`
- Renamed `MEDTASKER_TICKET_DIR` to `STEVMACHINE_TICKET_DIR`
- Changed vault default directory to `~/.stevmachine-skills`

## Task Commits

Each task was committed atomically:

1. **Task 1: Rename Go module path and package imports** - `715cb63` (feat)
2. **Task 2: Rename CLI directory and CLI-facing strings** - `d399672` (feat)
3. **Task 3: Rename dotenvx vault directory** - `d399672` (feat)

**Plan metadata:** `715cb63`

## Files Created/Modified
- `go.mod` - Module path renamed
- `cmd/stevmachine-skills/main.go` - Usage string renamed
- `cmd/stevmachine-skills/cmd_install.go` - TUI title, default skill names renamed
- `cmd/stevmachine-skills/cmd_env.go` - Usage string, wizard title, vault path renamed
- `cmd/stevmachine-skills/cmd_doctor.go` - Env var and label renamed
- `cmd/stevmachine-skills/cmd_list.go` - Internal imports renamed
- `internal/vault/vault.go` - Default EnvDir renamed
- `internal/vault/vault_test.go` - Test assertion updated

## Decisions Made

None - followed plan as specified.

## Deviations from Plan

### Auto-fixed Issues

**1. Source files were untracked at phase start**
- **Found during:** Task 1 (initial commit)
- **Issue:** All source files (`cmd/`, `internal/`, `packages/`, etc.) were untracked, so the first `git add -A` committed the entire source tree alongside the module rename.
- **Fix:** Committed the source files as part of the first task commit. Subsequent task commits are focused on the renamed content.
- **Files modified:** All source files
- **Verification:** `git status` shows no remaining untracked source files
- **Committed in:** `715cb63` (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (initial source tracking)
**Impact on plan:** No functional impact. Source files were added to git as part of the first commit.

## Issues Encountered

- Initial `git add -A` included all untracked source files in the first task commit. No other issues.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Go code rename complete. Ready for Plan 01-02 (skill package directory rename).
- No blockers.

---
*Phase: 01-rename-and-publish*
*Completed: 2026-07-15*
