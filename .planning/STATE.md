# STATE.md

## Project Reference

See: `.planning/PROJECT.md` (updated 2026-07-15)

**Core value:** Steve can own, ship, and update his personal AI assistant skills from one repo without exposing secrets.
**Current focus:** Phase 1 — Rename and publish as `stevmachine-skills`

## Current Status

- Project initialized from a repurposed employer-specific skill distribution repo.
- Codebase mapped and analyzed in `.planning/codebase/`.
- Repo moved from a nested home-directory worktree to a standalone repo at `/Users/estvmachine/Projects/personal/stevmachine-skills`.
- Standalone git repo initialized on `main` branch.

## Next Action

Run `/gsd-plan-phase 1` to plan Phase 1 execution.

## Important Notes

- Only Claude Code is currently implemented; OpenCode is Phase 2.
- `dotenvx` remains the credential mechanism.
- All `medtasker` / `nimblic` / `Medtasker` references must be removed before GitHub push.
- The project is a Go single binary with embedded skill packages.

## Active Workflows

- Mode: `yolo`
- Granularity: `standard`
- Parallelization: `true`
- Research: `false` (domain is known)
- Plan check: `true`
- Verifier: `true`
- Drift guard: `true`

## Files to Keep in Mind

- `go.mod` — module path must be renamed
- `cmd/medtasker-skills/` — binary and command handlers must be renamed
- `internal/vault/vault.go` — vault directory must be renamed
- `packages/embed.go` — embed paths may need updating
- `scripts/install.sh`, `scripts/verify-install.sh`, `scripts/smoke.sh` — scripts must be updated
- `README.md`, `DESIGN.md`, `CLAUDE.md` — docs must be rewritten
- `docs/` — assets may contain branding to replace
- `tapes/` — demo scripts may contain old branding

---
*Last updated: 2026-07-15 after project initialization*

## Session Continuity

- Last session: 2026-07-15
- Stopped at: Session resumed, awaiting next action selection
- Resume file: none
- Next action per STATE.md: `/gsd-plan-phase 1`
