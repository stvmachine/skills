# Quality & Testing Patterns

**Analysis Date:** 2026-07-15

## Testing Framework

**Runner:**
- Go's built-in `testing` package.
- Config: none; tests run with `go test ./...`.

**Run Commands:**
```bash
rtk go test ./...              # Run all tests
go test ./internal/mcp         # MCP parser/writer tests
go test ./internal/vault       # Vault tests (may skip if dotenvx missing)
go test ./cmd/...              # No tests currently exist for cmd package
```

**Coverage:**
- No coverage target enforced; no `go test -cover` CI gate detected.

## Test File Organization

**Location:** Co-located with source code (`*_test.go` next to implementation).

**Files:**
- `internal/mcp/mcp_test.go` — 149 lines, 4 test functions.
- `internal/vault/vault_test.go` — 126 lines, 7 test functions.

**Naming:**
- `TestParseSkillMcpConfig`, `TestBuildMcpServers`, `TestWriteMcpConfig`, `TestManagerPaths`, `TestListVars`, etc.

## Test Structure

**Suite Organization:**
```go
func TestParseSkillMcpConfig(t *testing.T) {
    dir := t.TempDir()

    // No SKILL.md
    _, err := ParseSkillMcpConfig(dir)
    if err == nil {
        t.Error("expected error for missing SKILL.md")
    }

    // ... further scenarios in same function
}
```

**Patterns:**
- Use `t.TempDir()` for isolated filesystem fixtures.
- Write small fixture files inline with `os.WriteFile`.
- Skip tests when external CLI dependency (`dotenvx`) is unavailable (`vault_test.go:57-63`).
- Tests are mostly sequential; no parallel tests or benchmarks detected.

## Mocking

**Framework:** None.

**Patterns:**
- Tests do not mock `dotenvx`; they rely on the real binary being on PATH or skip.
- `internal/mcp` tests use temporary files and directories instead of mocking the filesystem.
- `internal/vault` tests set a custom `EnvDir` to a temp directory (`vault_test.go:41`).

**What to Mock:**
- External CLI calls (`dotenvx`, `exec.Command`) should be mocked or injected for hermetic CI runs.

**What NOT to Mock:**
- Standard library filesystem operations are acceptable to exercise directly in Go tests.

## Fixtures and Factories

**Test Data:**
- Inline YAML frontmatter strings in `mcp_test.go:29-38` and `mcp_test.go:55-62`.
- Inline `.env` files in `vault_test.go:89`, `vault_test.go:106`, `vault_test.go:121`.

**Location:** Embedded in test files; no shared `testdata/` directory.

## Test Types

**Unit Tests:**
- `internal/mcp` — pure parser/builder/writer logic.
- `internal/vault` — path logic, initialization, masking, and plaintext destruction.
- `cmd` package — no tests exist.

**Integration Tests:**
- `smoke.sh` (`packages/claude-plugin/skills/run-medtasker-skills/smoke.sh`) exercises the full CLI build/install/list/doctor/env path against an isolated `$HOME`.
- `wizard.sh` drives the `huh` TUI end-to-end through tmux.

**E2E Tests:**
- Not used in CI; `scripts/verify-install.sh` is a manual post-install sanity check.

## Common Patterns

**Async Testing:**
- Not applicable; the code is synchronous.

**Error Testing:**
- Error cases are checked by comparing `err != nil` or by asserting substrings in error messages.

## Coding Conventions

**Files:**
- Go files use lowercase with hyphens for command files: `cmd_install.go`, `cmd_env.go`, `cmd_doctor.go`.
- Tests use `_test.go` suffix.

**Functions:**
- Exported functions use PascalCase (e.g., `ParseSkillMcpConfig`, `BuildMcpServers`).
- Unexported helpers use camelCase (e.g., `parseFrontmatterBytes`, `findDotenvx`).
- Command entry functions are named `cmd<Name>` (e.g., `cmdInstall`, `cmdEnvSet`).

**Variables:**
- Standard Go camelCase.
- Global lipgloss styles are grouped in `ui.go` under `var (...)`.

**Types:**
- Structs are exported (`ServerConfig`, `SkillFrontmatter`, `Manager`, `installModel`).

## Code Style

**Formatting:**
- `gofmt` / `go fmt` expected; no custom formatter config.

**Linting:**
- No `.golangci.yml`, `golangci-lint` config, or pre-commit hooks detected.
- No `go vet` CI gate detected.

## Import Organization

**Order:**
1. Standard library (`fmt`, `os`, `path/filepath`, etc.).
2. Blank line.
3. Third-party (`github.com/charmbracelet/...`, `gopkg.in/yaml.v3`).
4. Blank line.
5. Project internal (`github.com/nimblic/medtasker-skills/internal/mcp`, `.../packages`).

**Path Aliases:** None.

## Error Handling

**Patterns:**
- Return `error` from internal packages; commands print and `os.Exit(1)`.
- Some command code ignores non-fatal errors (e.g., `mcp.WriteMcpConfig` return values in `cmd_install.go:153,223`).
- `dotenvx` errors include combined stdout/stderr for diagnostics.

## Logging

**Framework:** `fmt` only.

**Patterns:**
- Direct `fmt.Println` / `fmt.Fprintln` for user-facing output.
- No structured logs or debug levels.

## Comments

**When to Comment:**
- Function-level comments for exported functions and helpers (e.g., `ui.go:18-21`, `cmd_install.go:197-198`).
- ADRs in `docs/` explain design decisions.

**JSDoc/TSDoc:** Not applicable (Go project).

## Function Design

**Size:**
- Functions are moderately sized; `cmd_install.go` and `cmd_env.go` are the largest command handlers.
- `cmdEnvSetup` (136 lines in `cmd_env.go:123-226`) builds a large `huh` form; consider decomposing into a builder function.

**Parameters:**
- Command handlers accept `args []string` from `os.Args`.
- Internal functions take filesystem paths and environment names explicitly.

**Return Values:**
- Internal functions return `(error)` or typed results.
- `doInstallOne` returns a boolean tuple `(ok bool, servers []string, errMsg string)`.

## Module Design

**Exports:**
- `internal/mcp` exports parsing and config-building functions.
- `internal/vault` exports `Manager` and its methods.
- `packages` exports only `SkillsFS`.

**Barrel Files:**
- None; each package has a small surface area.

## Notable Quality Issues

1. **No tests for `cmd` package.** The CLI surface (parsing, exit codes, env subcommand dispatch) is untested by Go tests; only `smoke.sh` covers it.
2. **Tests depend on external `dotenvx`.** `vault_test.go` skips or fails when `dotenvx` is unavailable, making CI non-deterministic unless the tool is pre-installed.
3. **No linting or CI gate.** There is no automated style, vet, or static-analysis check.
4. **Large generated skill files are embedded.** Skill Markdown files are long and contain hardcoded organization-specific references (e.g., `medtasker.atlassian.net`, `medtasker-frontends`), which may leak context if not reviewed.

---

*Quality analysis: 2026-07-15*
