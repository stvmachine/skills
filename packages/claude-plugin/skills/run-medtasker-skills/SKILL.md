---
name: run-medtasker-skills
description: Build, run, smoke-test, and drive the `medtasker-skills` Go CLI and its huh-based `env setup` TUI wizard. Use to verify a change to the CLI, install/list/doctor subcommands, MCP config writer, or vault wrapper, and to capture wizard screenshots.
---

# Run / drive medtasker-skills

`medtasker-skills` is a single Go binary that copies embedded skill packages into `~/.claude/skills/`, writes `~/.claude/.mcp.json`, and wraps `dotenvx` for an encrypted credential vault at `~/.medtasker-skills/`. Two surfaces matter for verifying a change:

1. **Subcommands** — `list`, `doctor`, `install`, `env set/list/encrypt/decrypt/rotate`. Exit codes + stdout. Drive with `smoke.sh`.
2. **`env setup`** — a `huh` form wizard that needs a TTY. Drive with `wizard.sh` under tmux.

The skill lives at `packages/claude-plugin/skills/run-medtasker-skills/` so it gets embedded in the distribution binary; `.claude/skills/run-medtasker-skills/` is a symlink to it for Claude Code auto-discovery in this repo. Either path works for invocation. Paths below are relative to the repo root.

## Prerequisites

Already installed on the host that authored this skill:

```bash
go version       # go1.26.3+ (tested with darwin/arm64)
dotenvx --version # 1.66.0 — see Gotchas
tmux -V          # 3.6a — required for the wizard driver
```

If `tmux` is missing: `brew install tmux`. The smoke script itself does **not** need tmux; only `wizard.sh` does.

## Run (agent path)

### Smoke the non-interactive surfaces

```bash
.claude/skills/run-medtasker-skills/smoke.sh
```

Builds the binary into a temp dir, sets up an isolated `$HOME` under `/tmp/mts-smoke.*`, then runs and asserts:

- `medtasker-skills` (no args) prints usage
- `list` shows the embedded skills
- `doctor` reports the isolated `.claude` dir
- `install medtasker-jira` copies the skill **and** writes `~/.claude/.mcp.json`
- `.mcp.json` is valid JSON **and** holds `${JIRA_API_TOKEN}` as a literal placeholder (the security invariant — resolved secrets in `.mcp.json` would be a regression)
- `env set` writes `~/.medtasker-skills/.env` with `encrypted:` ciphertext
- `go test ./internal/mcp ./cmd/...` passes

On success: `16 passed, 0 failed`. The temp `$HOME` is wiped on exit; pass `KEEP=1` to inspect.

```bash
KEEP=1 .claude/skills/run-medtasker-skills/smoke.sh
# prints "kept: /tmp/mts-smoke.XXXXXX" at the end
```

### Drive the env setup wizard

```bash
.claude/skills/run-medtasker-skills/wizard.sh decline   # walks past every prompt with 'n'
.claude/skills/run-medtasker-skills/wizard.sh jira      # types dummy Jira creds
.claude/skills/run-medtasker-skills/wizard.sh all       # Jira + GitHub
```

Each profile spawns the wizard inside a detached tmux session, waits for each prompt with `capture-pane`, sends the canned keystrokes, and finally verifies that:

- The summary line `Stored N variable(s)` matches the profile (0 / 3 / 4).
- The vault `.env` file contains `JIRA_URL="encrypted:..."` for the non-decline profiles — i.e. the wizard's plaintext input made it through dotenvx encryption.

`DEBUG=1` dumps every frame as it passes. `KEEP=1` leaves the temp `$HOME` for inspection.

### Capture a wizard frame ("screenshot")

The wizard is a text TUI, so frames are captured as text. Quick recipe (this is the actual command used to capture the current `samples/wizard-jira-prompt.txt`):

```bash
WIZ=$(mktemp -d); mkdir -p "$WIZ/home/.claude" "$WIZ/home/.medtasker-skills"
chmod 700 "$WIZ/home/.medtasker-skills"
go build -o "$WIZ/bin" ./cmd/medtasker-skills
tmux new-session -d -s shot -x 100 -y 40 "HOME=$WIZ/home $WIZ/bin env setup; sleep 5"
sleep 0.8; tmux send-keys -t shot Enter; sleep 0.5
tmux capture-pane -t shot -p > .claude/skills/run-medtasker-skills/samples/wizard-jira-prompt.txt
tmux kill-session -t shot
```

## Run (human path)

For a real user with credentials they actually want to keep:

```bash
go build -o /opt/homebrew/bin/medtasker-skills ./cmd/medtasker-skills
medtasker-skills install        # installs into your real ~/.claude/skills/
medtasker-skills env setup      # interactive — fill in your tokens
dotenvx run -f ~/.medtasker-skills/.env -- claude
```

Don't run this against `$HOME` while iterating on the code; use the driver scripts above instead, which sandbox everything under `/tmp`.

## Direct invocation (subset of changes)

When a PR touches only one of the internal packages, you don't need the full smoke. Targeted runs:

```bash
go test ./internal/mcp                          # MCP frontmatter parser + config writer
go test ./cmd/...                               # CLI tests
go run ./cmd/medtasker-skills list              # quickest sanity check
```

The `internal/vault` package wraps the `dotenvx` CLI. Its tests are excluded from `smoke.sh` by default — see Gotchas.

## Gotchas

- **`.mcp.json` placeholders must stay literal.** The whole point of this tool is that `${JIRA_API_TOKEN}` never gets resolved to the actual token in `~/.claude/.mcp.json`. `smoke.sh` checks this; if you change `internal/mcp` and break it, the smoke fails with `JIRA_API_TOKEN missing or resolved -- this would leak secrets`. Don't bypass that assertion.
- **`env set` needs the vault dir to exist.** `vault.Manager.Set` runs `dotenvx` with `cmd.Dir = ~/.medtasker-skills`, but never `MkdirAll`s it (only `InitVault` does, which isn't called from `cmdEnvSet` or `cmdEnvSetup`). On a fresh `$HOME`, both `set` and the wizard fail with `chdir … no such file or directory` until the dir exists. The smoke scripts pre-create it. The CLI itself is missing this step — file an issue rather than working around it in tests.
- **`medtasker-skills env list` says "No vault initialized" even when it is.** `vault.IsInitialized` checks for `.env.keys`, but dotenvx ≥ 1.66 stores the private key in the system keyring (or `DOTENV_PRIVATE_KEY` env var) and **never writes `.env.keys`** by default. The data is there — `cat ~/.medtasker-skills/.env` will show the `encrypted:` lines.
- **`internal/vault` tests fail by default.** Same root cause: they call `vault.Get`, which needs a private key dotenvx 1.66 doesn't have lying around. The smoke script excludes them; set `RUN_VAULT_TESTS=1 ./smoke.sh` only if you've separately arranged `DOTENV_PRIVATE_KEY` in the environment.
- **`install` emits a bubbletea TUI with ANSI escapes even when stdout is a pipe.** It runs `tea.NewProgram(...).Run()` first; the non-TTY fallback only kicks in if `Run()` returns an error. In practice the output you capture has escape sequences like `[?25l` mixed in — `grep -F` for keywords still works, but don't try to byte-compare it against a fixture.
- **The huh wizard advances on single-key Yes/No.** `y` and `n` *both* accept the value and move on (huh's behavior — see the bottom hint `y Yes • n No`). `enter` advances using the currently-highlighted choice. The wizard driver relies on this; if you change the form to require an explicit `enter` after `y/n` it'll desync.
- **`scripts/install.sh`'s post-install banner used to print `.env.vault`.** That's the legacy dotenvx-v0 layout; the actual file is `.env` (encrypted in place). Already fixed in this repo, but if you see it elsewhere or in a downstream fork, swap it.

## Troubleshooting

| Symptom | Fix |
|---|---|
| `chdir … /home/.medtasker-skills: no such file or directory` | The vault dir doesn't exist. `mkdir -p ~/.medtasker-skills && chmod 700 ~/.medtasker-skills`. The smoke scripts already do this. |
| `MISSING_PRIVATE_KEY` from `env decrypt` or `env list` | dotenvx ≥ 1.66 needs the private key in env. `export DOTENV_PRIVATE_KEY=$(dotenvx keypair -f ~/.medtasker-skills/.env DOTENV_PRIVATE_KEY)` or accept that `env list` won't decrypt — the smoke skips this check. |
| `wizard.sh` hangs with `TIMEOUT waiting for: <text>` | The form has been re-ordered or re-worded. Run with `DEBUG=1` to see the actual frame, then update the matching `expect "..."` call in `wizard.sh`. |
| `tmux: command not found` | `brew install tmux`. Only `wizard.sh` needs it; `smoke.sh` doesn't. |
| `go: cannot find main module` | Run the driver scripts from anywhere — they `cd` to the repo root themselves via `$(cd "$(dirname "$0")/../../.." && pwd)`. If you copied the script elsewhere, fix that path. |
