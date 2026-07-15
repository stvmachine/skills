# ADR-004: Beads as an Optional Storage Backend

## What beads is

Beads (`gastownhall/beads-mcp`, CLI `bd`) is a local SQLite-backed issue tracker designed as "memory upgrade for your coding agent." It stores tickets, labels, status, priority, and notes in `~/.beads/`, queryable from the shell with `bd query "label=X"`.

## What we use it for

When a user runs `/stevmachine-jira inbox MT-1234`, that ticket — description, acceptance criteria, related links, recent comments — has to be stored locally so downstream skills (`/stevmachine-jira-ticket-transition qa`, future planners) can read it without going back to Jira every time. We need a place to put it.

## What we tried first

Originally we tried beads-as-required: all skills wrote to `bd`, queried from `bd`, and assumed `bd` was on `PATH`. We discovered three problems:

1. **Install friction.** `bd` is a separate dependency on a separate release cycle, with its own update lifecycle. A user installing stevmachine-skills had to install another tool just to fetch their first Jira ticket. The skills failed cryptically (`chdir … no such file or directory`) when `bd` was missing.

2. **Opacity.** Bead notes live in SQLite. There's no `cat ./.todo/MT-1234/TICKET_DESCRIPTION.md` to glance at a ticket. Every inspection goes through `bd show <id>`, which is fine when you're driving via Claude but annoying when you just want to see what you've got.

3. **Surface coupling.** Every skill that touched a ticket ended up with `bd query "label=..."` hardcoded into its workflow, then `bd update` for status changes. The skills weren't really about Jira workflow — they were about Jira-workflow-via-beads. Swapping the backend later would have required editing every skill body.

## What we picked

**Filesystem default, beads optional, auto-detect at call time.**

| Condition | Backend | Storage |
|---|---|---|
| `command -v bd` succeeds AND `STEVMACHINE_FORCE_FILESYSTEM` unset | Beads | `~/.beads/` |
| Otherwise | Filesystem | `${STEVMACHINE_TICKET_DIR:-./.todo}/<TICKET-ID>/TICKET_DESCRIPTION.md` |

Both backends carry the same schema (status values, priority mapping, notes template). The full contract is in `packages/claude-plugin/skills/stevmachine-jira/rules/ticket-storage.md`. Downstream skills (`stevmachine-jira-ticket-transition`) check both, prefer beads when present.

## Why this shape

- **Default = filesystem.** A user who just ran `stevmachine-skills install` and `/stevmachine-jira inbox` should get a working ticket without installing anything else. They will, because `./.todo/MT-1234/TICKET_DESCRIPTION.md` requires nothing beyond `mkdir`.

- **`bd` is an optimization, not a prereq.** When beads is present, the skill uses it transparently — same skill code, different backend. The user gets the structured-query benefits (`bd query "label=jira" --status open` is genuinely useful) without those benefits being a prerequisite for anyone else.

- **`STEVMACHINE_FORCE_FILESYSTEM=1` is the escape hatch.** For CI runners, ephemeral shells, or when you want to keep `bd` clean while testing something locally, you can force filesystem mode without uninstalling `bd`.

- **`STEVMACHINE_TICKET_DIR` is configurable, defaults sensibly.** `./.todo/` puts ticket context next to the code it relates to — branch and repo together, easy to `.gitignore`. Setting `STEVMACHINE_TICKET_DIR=~/.stevmachine-tickets` makes context global if you'd rather have it follow you across repos.

## What we lose vs. beads-required

- **Structured queries across all your tickets.** `bd query "priority>=4 status=open"` doesn't have a filesystem equivalent without writing a custom grep. We accept this — when beads is installed, the query works; when it isn't, the user wasn't going to run that command anyway.
- **Dedup and relation primitives.** Beads has built-in support for linked issues, blocking relationships, etc. The filesystem backend stores related-ticket info as markdown lines, which is enough for display but not for navigation. If a future skill needs real graph queries, it should either require beads or add its own index — not invent half a beads in markdown.
- **One source of truth.** With auto-detect, two users on the same repo could end up with one keeping notes in beads and the other on filesystem. Per-machine state divergence is acceptable here because the ticket records aren't shared — they're personal working context.

## What we do about it

The `doctor` command checks for `bd`. If missing, it prints:

```
beads (bd) not installed — Jira tickets will use filesystem (./.todo/) instead of beads.
  → Install: curl -fsSL https://raw.githubusercontent.com/gastownhall/beads/main/integrations/beads-mcp/install.sh | bash
```

The framing is deliberate: missing `bd` is not an error, it's a different mode. Users who try beads and don't like it can remove `bd` from their PATH (or set `STEVMACHINE_FORCE_FILESYSTEM=1`) and lose nothing.

## Related

- ADR-003 (RTK): same posture — suggest, don't bundle, host-level tool managed by the user.
- `packages/claude-plugin/skills/stevmachine-jira/rules/ticket-storage.md`: the full schema for both backends.
