# Ticket Storage

How fetched Jira tickets are stored locally. There are two backends — the skill picks one **automatically** at runtime:

| Condition | Backend | Storage location |
|---|---|---|
| `bd` (beads-mcp) is installed (`command -v bd`) | **Beads** — a SQLite-backed `bd` issue | The bead DB (typically `~/.beads/`) |
| Otherwise | **Filesystem** | `${STEVMACHINE_TICKET_DIR:-./.todo}/<TICKET-ID>/` |

The user can also override by setting `STEVMACHINE_FORCE_FILESYSTEM=1` to skip beads even when installed (useful in CI or one-off shells where `bd` exists but you don't want to write to its DB).

To check which backend is active: `command -v bd && [ -z "$STEVMACHINE_FORCE_FILESYSTEM" ] && echo beads || echo filesystem`.

Both backends store the **same information**. Downstream consumers (stevmachine-jira-ticket-transition, future skills) must check for both and prefer beads if found.

---

## Beads backend

Used when `bd` (beads-mcp) is installed and `STEVMACHINE_FORCE_FILESYSTEM` is not set.

### Bead schema

A fetched ticket becomes one bead. Fields:

| Bead field | Source | Notes |
|---|---|---|
| `title` | `$TICKET_ID: $SUMMARY` | E.g. `MT-9661: Bad date format on /reports page` |
| `description` | Short Jira description excerpt | Full description goes in `notes`. |
| `labels` | `jira`, `$TICKET_ID`, ticket type | Type labels are e.g. `bug`, `quick-fix`. |
| `status` | `open` / `in_progress` / `closed` | See status mapping below. |
| `priority` | Jira priority → 1–5 | See priority mapping below. |
| `notes` | Full ticket context | Metadata table, description, repro, acceptance criteria, attachments, related tickets, last 3 comments. |

### Status mapping

- **open** — fetched, ready for planning. Use this on initial inbox.
- **in_progress** — work has started (set by `/stevmachine-jira-ticket-transition review`).
- **closed** — work completed and shipped to QA (set by `/stevmachine-jira-ticket-transition qa`).

---

## Filesystem backend

Used when `bd` (beads-mcp) is missing, or when `STEVMACHINE_FORCE_FILESYSTEM=1`.

### Directory layout

```
${STEVMACHINE_TICKET_DIR:-./.todo}/
  <TICKET-ID>/
    TICKET_DESCRIPTION.md   # Single source-of-truth file (replaces bead notes)
    attachments/            # Downloaded attachments, if any
```

The default `./.todo/` is **project-local** (one ticket dir per branch/repo). Add `.todo/` to `.gitignore` — these are working notes, not source.

`STEVMACHINE_TICKET_DIR` can be an absolute path (`~/.stevmachine-tickets/`) for global storage that follows you across repos. Persist by storing it in the dotenvx vault:

```bash
stevmachine-skills env set STEVMACHINE_TICKET_DIR /Users/you/.stevmachine-tickets
```

### TICKET_DESCRIPTION.md schema

```markdown
# <TICKET-ID>: <Summary>

## Metadata
| Field | Value |
|---|---|
| **Status** | $STATUS (open | in_progress | closed) |
| **Type** | $ISSUE_TYPE |
| **Assignee** | $ASSIGNEE |
| **Priority** | P$PRIORITY |
| **Labels** | $LABELS |
| **Jira** | https://example.atlassian.net/browse/$TICKET_ID |

## Description
$RENDERED_DESCRIPTION

## Acceptance Criteria
$ACCEPTANCE_CRITERIA

## Attachments
- attachments/<filename> ($SIZE_KB KB)

## Related Tickets
- $LINK_TYPE: $LINKED_KEY — $LINKED_SUMMARY [$LINKED_STATUS]

## Subtasks
- [ ] $SUBTASK_KEY — $SUBTASK_SUMMARY [$SUBTASK_STATUS]

## Recent Comments
**[$DATE] $AUTHOR:**
$COMMENT_BODY
```

The `**Status**` field in the metadata table is the local lifecycle state, mirroring the bead status values (`open`, `in_progress`, `closed`). Downstream skills (stevmachine-jira-ticket-transition) update this field in place when transitioning.

---

## Common: Priority mapping

| Jira priority | Local priority |
|---|---:|
| Highest | 5 |
| High | 4 |
| Medium | 3 |
| Low | 2 |
| Lowest | 1 |

Same on both backends (`bd priority` field; `**Priority**: P3` line in the metadata table).

## Common: Fields to extract from the MCP response

Call `mcp__mcp-atlassian__jira_get_issue` with `fields="*all"` and `expand="renderedFields"`. Extract:

- `key` → ticket ID
- `fields.summary` → title
- `fields.description` (raw) and `renderedFields.description` (markdown) → description
- `fields.status.name` → metadata
- `fields.issuetype.name` → metadata + label
- `fields.assignee.displayName` → metadata (may be null)
- `fields.priority.name` → priority mapping above
- `fields.labels` → labels
- `fields.comment.comments` → last 3 in "Recent Comments"
- `fields.issuelinks` → "Related Tickets"
- `fields.subtasks` → "Subtasks"
- `fields.attachment` → reference URLs (download only if user asks for them)

## Common: Readiness contract

A ticket is "ready for planning" when **all** of:

1. A record exists for the ticket — either a bead with `label=<TICKET-ID>`, or a `${STEVMACHINE_TICKET_DIR:-./.todo}/<TICKET-ID>/TICKET_DESCRIPTION.md` file.
2. The record has non-empty content (notes / file body) with at least the metadata table and description.
3. Status is `open` or `in_progress` (not `closed`).

Downstream skills check both backends in this order, returning the first hit:

```bash
# Beads first (if available)
if command -v bd >/dev/null 2>&1 && [ -z "$STEVMACHINE_FORCE_FILESYSTEM" ]; then
    BEAD=$(bd query "label=$TICKET_ID" --sort updated --reverse --json | python3 -c "import sys,json; r=json.load(sys.stdin); print(r[0]['id'] if r else '')")
    if [ -n "$BEAD" ]; then
        echo "Found bead: $BEAD"
        exit 0
    fi
fi

# Filesystem fallback
TICKET_DIR="${STEVMACHINE_TICKET_DIR:-./.todo}/$TICKET_ID"
if [ -f "$TICKET_DIR/TICKET_DESCRIPTION.md" ]; then
    echo "Found ticket file: $TICKET_DIR/TICKET_DESCRIPTION.md"
    exit 0
fi

echo "ERROR: no record for $TICKET_ID — run /stevmachine-jira inbox $TICKET_ID"
exit 1
```

## Common: Update / re-fetch

Re-fetching is **destructive overwrite**, not merge. On both backends:

- **Beads**: `bd update <bead-id> --notes ""` to clear, then re-run fetch.
- **Filesystem**: `rm "$TICKET_DIR/TICKET_DESCRIPTION.md"` (preserve `attachments/`), then re-run fetch.

Use when the ticket changed in Jira (new comments, edited description, status moved) and you want the local copy to reflect current state. Don't merge old and new content — too much room for drift.
