# Codebase Concerns

**Analysis Date:** 2026-07-15

## Tech Debt

**Claude-Only Platform Coupling:**
- Issue: The entire distribution is hardcoded to Claude Code. `packages/claude-plugin/` is the only implemented package; `packages/opencode-package/` from early design drafts does not exist.
- Files: `cmd/stevmachine-skills/cmd_install.go`, `cmd_list.go`, `cmd_doctor.go`, `packages/embed.go`.
- Impact: Supporting OpenCode or other platforms requires a significant refactor, not just adding new packages.
- Fix approach: Introduce a platform interface (`PlatformInstaller`) and per-platform embed directories; route `install`/`list`/`doctor` through the platform abstraction.

**Hardcoded Branding Strings:**
- Issue: `stevmachine`, `claude`, and `Claude Code` strings are scattered throughout the codebase, making rename or white-label expensive.
- Files:
  - `cmd/stevmachine-skills/cmd_install.go:88` — `Stevmachine Skills` title.
  - `cmd/stevmachine-skills/cmd_install.go:136-140` — `~/.claude`, `~/.claude.json`.
  - `cmd/stevmachine-skills/cmd_install.go:144` — `Claude Code not detected`.
  - `cmd/stevmachine-skills/cmd_install.go:205` — default skill names all prefixed `stevmachine-`.
  - `cmd/stevmachine-skills/cmd_doctor.go:14-15,27` — `Claude Code dir` and `~/.claude/.mcp.json`.
  - `cmd/stevmachine-skills/cmd_env.go:224-225` — launch Claude command.
  - `internal/vault/vault.go:19` — `~/.stevmachine-skills`.
  - `packages/embed.go:5` — `//go:embed all:claude-plugin/skills`.
  - `scripts/install.sh`, `scripts/verify-install.sh`, `tapes/setup-flow.sh`, `docs/INSTALL.md` — numerous references.
- Impact: A product rename or dual-platform release would require a wide, error-prone string sweep.
- Fix approach: Centralize product name, target directory, and binary name constants in a single package and generate user-facing strings from them.

**Shallow MCP Config Merge:**
- Issue: `WriteMcpConfig` merges top-level `mcpServers` entries by overwriting the entire server value. If two skills define the same MCP server with different `env` or `args`, the last one wins silently.
- Files: `internal/mcp/mcp.go:116-137`.
- Impact: Conflicts between skills are not detected; partial config loss is possible.
- Fix approach: Deep-merge server entries (env, args, headers) or detect conflicts and prompt/fail.

**Missing Vault Directory Initialization:**
- Issue: `vault.Manager.Set` and `cmdEnvSetup` do not ensure `~/.stevmachine-skills` exists before invoking `dotenvx` with `cmd.Dir = EnvDir`.
- Files: `internal/vault/vault.go:105-112`, `cmd/stevmachine-skills/cmd_env.go:188-200`.
- Impact: On a fresh machine, `env set` or `env setup` fails with `chdir … no such file or directory`.
- Fix approach: Call `os.MkdirAll(EnvDir, 0o700)` at the start of every public vault method, or create a lazy initialization helper.

**Vault Initialization Checks Wrong File:**
- Issue: `IsInitialized` checks for the existence of `.env.keys`, but modern `dotenvx` (≥1.66) stores the private key in the system keyring and does not write `.env.keys` by default.
- Files: `internal/vault/vault.go:37-40`, `cmd/stevmachine-skills/cmd_doctor.go:29`.
- Impact: `stevmachine-skills env list` reports `No vault initialized` even when the vault is populated and encrypted.
- Fix approach: Check for the existence of `.env` (encrypted) or test `dotenvx get` instead of relying on the legacy keys file.

**Dependency Resolution Not Implemented:**
- Issue: `ROADMAP.md` plans a dependency resolver for Phase 3, but the CLI does not yet parse `dependencies:` from skill frontmatter or resolve installation order.
- Files: `cmd/stevmachine-skills/cmd_install.go:157-160`.
- Impact: `stevmachine-jira-ticket-transition` depends on `commit`, `stevmachine-jira`, and `stevmachine-jira-markup`, but installing it alone will not install those dependencies.
- Fix approach: Parse `dependencies` from `SkillFrontmatter`, build a graph, and install in topological order before the requested skill.

## Known Bugs

**TUI Emits ANSI in Piped Output:**
- Symptoms: `stevmachine-skills install` runs `tea.NewProgram(...).Run()` first; only if that returns an error does it fall back to plain output. Piped output therefore contains escape sequences.
- Files: `cmd/stevmachine-skills/cmd_install.go:162-184`.
- Trigger: Run `stevmachine-skills install | cat` or capture output in a script.
- Workaround: The `smoke.sh` test greps for keywords rather than comparing byte-for-byte.
- Fix approach: Detect non-TTY before starting the TUI and skip the Bubble Tea program entirely.

**MCP Config Written to Two Files with No Precedence Logic:**
- Symptoms: `install` writes to both `~/.claude/.mcp.json` and `~/.claude.json`. The ADR says `~/.claude.json` takes precedence, but the code only loops and writes to both.
- Files: `cmd/stevmachine-skills/cmd_install.go:138-140`, `internal/mcp/mcp.go:116-137`.
- Trigger: Every install.
- Workaround: Users must manually reconcile if the two files diverge.
- Fix approach: Write to a single canonical location or read the precedence file first and merge into it, then copy to the fallback if needed.

## Security Considerations

**Credential Vault Relies on External Node Tool:**
- Risk: `dotenvx` is a Node.js dependency; a compromised or breaking release could expose or lock away credentials.
- Files: `internal/vault/vault.go:52-79`, `scripts/install.sh:39-53`.
- Current mitigation: Pins are not used; `npm install -g @dotenvx/dotenvx` installs latest, and `findDotenvx` finds whatever is on PATH.
- Recommendations: Pin `dotenvx` version in install docs and verify checksums; consider vendoring a known-good binary.

**No Validation of MCP Server Commands:**
- Risk: Skill frontmatter can declare arbitrary `command` and `args` that are written directly into `.mcp.json` and executed by Claude Code.
- Files: `internal/mcp/mcp.go:84-114`.
- Current mitigation: Skills are bundled from the same repo; no third-party skill loading is implemented.
- Recommendations: Add an allow-list of known MCP server commands or signatures before writing to the user's config.

**`.env.keys` Not Always Written:**
- Risk: The install docs assume `.env.keys` exists, but dotenvx ≥1.66 may store keys in the system keyring. Users following legacy docs may commit or mishandle keys.
- Files: `.gitignore:1-7`, `docs/ADR-001-dotenvx.md`, `internal/vault/vault.go`.
- Current mitigation: `.gitignore` correctly excludes `.env.keys` and `.env.*.keys`.
- Recommendations: Update all docs to reflect the current dotenvx behavior and the `DOTENV_PRIVATE_KEY` workflow.

## Performance Bottlenecks

**Synchronous Sequential Installs:**
- Problem: `install` processes skills one at a time in a single goroutine.
- Files: `cmd/stevmachine-skills/cmd_install.go:54-62`.
- Cause: Each skill is installed and its MCP config written before the next begins.
- Improvement path: With only a handful of small skills, this is fine. If the skill catalog grows, consider copying files in parallel and merging MCP configs once at the end.

**Large Embedded SKILL.md Files:**
- Problem: Skill Markdown files are hundreds of lines and include detailed workflow rules, increasing binary size and compile time modestly.
- Files: `packages/claude-plugin/skills/stevmachine-jira/SKILL.md`, `stevmachine-jira-ticket-transition/SKILL.md`.
- Cause: Skills are embedded verbatim.
- Improvement path: Compress or split rule files; only embed what the CLI needs at runtime (frontmatter can be extracted at build time).

## Fragile Areas

**Skill Discovery via Embedded Directory Name:**
- Files: `cmd/stevmachine-skills/cmd_install.go:209`, `cmd_list.go:34`.
- Why fragile: Skills are discovered by walking `claude-plugin/skills/` inside the embedded FS. Renaming the package directory or adding a new platform requires updating every discovery site.
- Safe modification: Add a `const skillsDir = "claude-plugin/skills"` and use it everywhere.
- Test coverage: Only `smoke.sh` covers discovery; no unit tests.

**Default Skill List Hardcoded:**
- Files: `cmd/stevmachine-skills/cmd_install.go:204-206`.
- Why fragile: New skills must be added to this slice manually or they will not be installed by `stevmachine-skills install` without explicit arguments.
- Safe modification: Derive the default list from the embedded FS at runtime.
- Test coverage: `smoke.sh` checks that `stevmachine-jira` is in `list`, but not the default install set.

**RTK and Beads Suggestions Are macOS-Centric:**
- Files: `cmd/stevmachine-skills/cmd_doctor.go:73-89`, `scripts/install.sh` (implicit), `docs/ADR-003-rtk.md`, `docs/ADR-004-beads.md`.
- Why fragile: Suggestions use `brew install rtk` and a curl-based beads installer, which do not apply to all platforms.
- Safe modification: Detect `runtime.GOOS` and print appropriate install instructions.
- Test coverage: No tests for suggestion output.

## Scaling Limits

**Skill Catalog:**
- Current capacity: ~4 default skills, all embedded.
- Limit: Every skill is copied and parsed at install time; the TUI currently runs one install at a time.
- Scaling path: Parallelize copies or build a manifest index so the CLI does not walk the entire embedded FS.

**MCP Config File Size:**
- Current capacity: Small JSON file with a few servers.
- Limit: Shallow merge and repeated JSON marshal/unmarshal become expensive and lossy as more servers are added.
- Scaling path: Implement a typed MCP config model with deep merge and conflict detection.

## Dependencies at Risk

**dotenvx:**
- Risk: Node.js-based external CLI; breaking changes or supply-chain issues directly impact credential operations.
- Impact: All `env` commands and the launch workflow stop working.
- Migration plan: Pin version and checksums; evaluate a native Go implementation of ECIES vault operations if the Node dependency becomes unacceptable.

**Bubble Tea / huh:**
- Risk: Major version or API changes in charmbracelet libraries.
- Impact: TUI build breaks.
- Migration plan: Pin minor versions in `go.mod`; current versions are v0/v1 and relatively stable.

## Missing Critical Features

**Multi-Platform Support:**
- Problem: Only Claude Code is implemented today. OpenCode is planned for Phase 2.
- Blocks: Distribution to non-Claude users.

**Dependency Resolver:**
- Problem: `dependencies:` in skill frontmatter is ignored.
- Blocks: Skills that rely on other skills cannot be installed safely with a single command.

**Update Command:**
- Problem: The roadmap includes an `update` command, but the CLI does not implement it yet.
- Blocks: No first-class upgrade path for installed skills.

**CI/CD Pipeline:**
- Problem: No GitHub Actions, GitLab CI, or similar detected.
- Blocks: Automated testing, linting, and asset generation on PRs.

## Test Coverage Gaps

**Untested CLI surface:**
- What's not tested: Subcommand parsing, exit codes, install TUI fallback, env setup form, doctor output.
- Files: `cmd/stevmachine-skills/*.go`.
- Risk: Refactoring commands can break user-facing behavior without failing tests.
- Priority: High.

**Untested MCP conflict scenarios:**
- What's not tested: Same MCP server defined by two skills with different env/args; malformed existing `.mcp.json`.
- Files: `internal/mcp/mcp.go`.
- Risk: Silent config loss or corruption.
- Priority: Medium.

**Untested vault edge cases:**
- What's not tested: Missing `dotenvx` error handling, custom environment files, directory permission failures, keyring-based dotenvx versions.
- Files: `internal/vault/vault.go`.
- Risk: Users on fresh machines or newer dotenvx versions hit unhandled paths.
- Priority: High.

---

*Concerns audit: 2026-07-15*
