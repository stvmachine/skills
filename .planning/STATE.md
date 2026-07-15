# STATE.md

## Project Reference

See: `.planning/PROJECT.md` (updated 2026-07-15)

**Core value:** Steve can own, ship, and update his personal AI assistant skills from one repo without exposing secrets.
**Current focus:** Phase 2 — Add OpenCode support

## Current Status

- Phase 1 complete: all `medtasker` / `nimblic` / `Medtasker` references removed from source and docs.
- Go module, CLI binary, vault directory, and embedded skill packages renamed to `stevmachine-*`.
- Smoke test and install verification pass with the renamed binary.
- Repo published to public GitHub: https://github.com/stvmachine/skills
- Fresh clone builds and passes install verification.

## Next Action

Run `/gsd-plan-phase 2` to plan Phase 2 (OpenCode support).

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
*Last updated: 2026-07-15 after Phase 1 execution*

## Session Continuity

- Last session: 2026-07-15
- Stopped at: Phase 1 complete, awaiting Phase 2 planning
- Resume file: none
- Next action per STATE.md: `/gsd-plan-phase 2`
