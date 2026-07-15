# ADR-003: RTK as a Suggested (Not Bundled) Optimization

## What RTK is

RTK ("Rust Token Killer") is a shell-level CLI proxy that filters/compresses the output of common dev commands before it lands in an agent's context. Examples from its own docs: `git status` becomes compact, `git log` drops timestamps the agent doesn't need, `npm install` strips progress chatter, `cargo test` returns only failures, etc. Typical savings: **60–90%** on the operations agents run hundreds of times per session.

It works via a Claude Code hook that transparently rewrites `git status` → `rtk git status` and pipes through filters. No code changes needed in the agent; no MCP server; no API key. `brew install rtk` and a one-time hook install.

## What we do about it

The `doctor` command checks for `rtk` on `PATH`. If missing, both `doctor` and `install` print:

```
rtk not installed — Claude Code commands won't be token-optimized.
  → Install: brew install rtk
```

That's the entire integration. RTK is not bundled, not auto-installed, not declared in any skill's frontmatter.

## Why suggest instead of bundle

We considered three postures and rejected two:

1. **Bundle / require.** Rejected. RTK isn't an MCP; it's a shell-environment modification (PATH, hooks). The `medtasker-skills install` flow is about copying skill packages + writing `.mcp.json`. Pulling in a shell hook would mean we're now also managing the user's shell config — a far larger surface area, with rollback semantics we'd need to design from scratch. Different problem, different tool.

2. **Ignore entirely.** Rejected. The token savings are real and compound across every shell-using skill. A user who installs medtasker-skills, runs into a Jira ticket fetch that pulls 30k tokens of `gh pr view` output, and never knew RTK existed has been failed by us. Saying nothing is a bug.

3. **Suggest in `doctor` and `install`.** Picked. Cost: ~10 lines in `cmd_doctor.go` (the `printSuggestions` helper). Value: every fresh-install user gets a one-line nudge with the exact `brew install` command. They can ignore it. They can install it later. The installation, configuration, and removal of RTK stays entirely in the user's hands — which is the right place for a global shell modification.

## Why not list it as a hard prereq

RTK is a **personal-productivity** tool. Some teams may have policies against shell hooks; some users prefer raw output. The skills work fine without it — they just use more tokens. Treating it as a prereq would block users from a tool they don't actually need.

The same argument applies to beads-mcp (ADR-004): suggested, not bundled.

## When the suggestion is wrong

- **CI environments.** `brew install rtk` is suggested even on a CI runner that just wants to run `medtasker-skills install` once. The suggestion is harmless (it's `stderr`, doesn't change exit code), but worth knowing the suggestion is human-targeted.
- **Linux without brew.** The suggestion command is `brew install rtk`. On a Linux box without Homebrew, that line is dead. Acceptable for now — the project's primary platform is macOS — but if Linux usage grows, the suggestion text should detect platform and print the right install path.
- **Name collision warning.** Per `~/.claude/RTK.md`: if `rtk gain` fails, you may have `reachingforthejack/rtk` (Rust Type Kit) installed instead. We don't currently detect this; if a user reports the install hint not helping, that's the first thing to check.

## Implementation pointer

The suggestion lives in `cmd/medtasker-skills/cmd_doctor.go::printSuggestions`. Adding another suggested tool is one `if missing { fmt.Println(...) }` block — same shape as the `bd` (beads-mcp) suggestion right next to it.
