# Technology Stack

**Analysis Date:** 2026-07-15

## Languages

**Primary:**
- Go 1.22 (declared in `go.mod`), Go 1.26.5 (observed runtime) — entire CLI and embedded package distribution.

**Secondary:**
- Markdown — skill definitions (SKILL.md with YAML frontmatter).
- Bash — installer, verification, smoke tests, and asset generation scripts.
- JSON — plugin manifest and generated MCP config.
- TOML — RTK filter config (currently empty template).

## Runtime

**Environment:**
- Go standard library runtime; no custom server or container.
- Targets end-user workstations (macOS primary; Linux noted but less-tested; Windows not supported by `install.sh`).

**Package Manager:**
- Go modules (`go.mod` / `go.sum`).
- Lockfile: `go.sum` present.
- No vendored dependencies detected.

## Frameworks

**Core:**
- Standard library only for business logic (`flag`, `os/exec`, `path/filepath`, `embed`, `io/fs`, `encoding/json`, `testing`).

**TUI / UI:**
- `github.com/charmbracelet/huh` v0.6.0 — `env setup` wizard forms.
- `github.com/charmbracelet/bubbletea` v1.1.0 (indirect via `bubbles`) — install spinner/progress TUI.
- `github.com/charmbracelet/bubbles` v0.20.0 (indirect) — spinner component.
- `github.com/charmbracelet/lipgloss` v1.0.0 — color styles and status rendering.

**Parsing:**
- `gopkg.in/yaml.v3` v3.0.1 — SKILL.md YAML frontmatter.

**Build/Dev:**
- `go build` — binary compilation.
- `scripts/install.sh` — full installer (dotenvx + binary + skills).
- `scripts/verify-install.sh` — post-install sanity check.
- `tapes/generate.sh` — README asset generation (requires `freeze`, `vhs`, `ffmpeg`, `ttyd`).
- `packages/claude-plugin/skills/run-medtasker-skills/smoke.sh` — smoke tests for non-interactive surfaces.
- `packages/claude-plugin/skills/run-medtasker-skills/wizard.sh` — TUI wizard driver via tmux.

## Key Dependencies

**Critical:**
- `github.com/charmbracelet/huh` v0.6.0 — Interactive credential setup form.
- `github.com/charmbracelet/lipgloss` v1.0.0 — Terminal styling.
- `gopkg.in/yaml.v3` v3.0.1 — Skill frontmatter parsing.

**Infrastructure (external binaries, not Go modules):**
- `dotenvx` (`@dotenvx/dotenvx`) — Encryption/decryption of credential vault. Required at runtime.
- `claude` / Claude Code (`@anthropic-ai/claude-code`) — Target agent platform; `install` and `doctor` check for `~/.claude`.
- `npx` / Node.js — Used by MCP servers declared in skill frontmatter (e.g., GitHub, Figma).
- `uv` / `uvx` — Used to run `mcp-atlassian` Python MCP server.
- `beads-mcp` (`bd`) — Optional ticket storage backend.
- `rtk` — Optional token-output optimizer suggested by `doctor`/`install`.
- `tmux` — Required only by `wizard.sh` TUI driver.

**Indirect:**
- `github.com/charmbracelet/bubbles`, `bubbletea`, `x/ansi`, `x/term`, `x/exp/strings` — Bubble Tea ecosystem.
- `github.com/mitchellh/hashstructure/v2` — used by Bubble Tea/huh internals.
- `golang.org/x/sync`, `x/sys`, `x/text` — standard transitive dependencies.

## Configuration

**Environment:**
- No `.env` files read by the Go binary itself.
- `dotenvx` manages the user's credential vault at `~/.medtasker-skills/.env`.
- `MEDTASKER_TICKET_DIR` — optional directory for filesystem ticket storage backend.
- `MEDTASKER_FORCE_FILESYSTEM=1` — forces filesystem backend even if `bd` is installed.
- `MEDTASKER_VAULT` — optional custom vault path mentioned in README wrapper examples (not consumed by Go code directly).
- `MEDTASKER_REPO` — used by `smoke.sh` and `wizard.sh` to locate source repo.

**Build:**
- `go.mod` — module declaration (`github.com/nimblic/medtasker-skills`), Go version 1.22.
- `go.sum` — dependency checksums.
- No `Makefile`, `Taskfile`, or CI configuration detected.
- `.rtk/filters.toml` — empty RTK filter template.

## Platform Requirements

**Development:**
- Go 1.22+.
- Node.js + npm for `dotenvx` and Claude Code.
- SSH access to `github.com/nimblic/medtasker-skills` (private repo).
- Optional: `tmux`, `beads-mcp`, `rtk`, `uv`, `vhs`, `freeze`, `ffmpeg`, `ttyd`.

**Production:**
- No server-side production deployment; this is a client-side CLI.
- Distribution is "clone + build" from a private repo; `go install` and `curl | bash` are intentionally unsupported.

---

*Stack analysis: 2026-07-15*
