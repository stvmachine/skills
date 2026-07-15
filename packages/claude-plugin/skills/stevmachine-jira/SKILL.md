---
name: medtasker-jira
description: Connect code work to Jira tickets — inbox for gathering full ticket context, read context, update status, link PRs, and auto-detect ticket IDs from branch names.
allowed-tools: mcp__mcp-atlassian__*, Bash(git *), Bash(curl *), Bash(file *), Bash(ls *), Bash(bd *), Read
mcp_servers:
  - name: mcp-atlassian
    type: stdio
    command: uvx
    args: [mcp-atlassian, --transport, stdio]
    env:
      JIRA_URL: ${JIRA_URL}
      JIRA_USERNAME: ${JIRA_USERNAME}
      JIRA_API_TOKEN: ${JIRA_API_TOKEN}
      CONFLUENCE_URL: ${CONFLUENCE_URL}
      CONFLUENCE_USERNAME: ${CONFLUENCE_USERNAME}
      CONFLUENCE_API_TOKEN: ${CONFLUENCE_API_TOKEN}
  - name: beads-mcp
    type: stdio
    command: beads-mcp
  - name: github
    type: stdio
    command: npx
    args: [-y, "@modelcontextprotocol/server-github"]
    env:
      GITHUB_TOKEN: ${GITHUB_TOKEN}
  - name: context7
    type: http
    url: https://mcp.context7.com/mcp
    headers:
      CONTEXT7_API_KEY: ${CONTEXT7_API_KEY}
---

# Jira Integration Skill

Connects code work to Jira tickets using the mcp-atlassian MCP server.

## Capabilities

### 0. Ticket Inbox (NEW)
Fetch a Jira ticket with full context and store it locally. The skill **auto-detects** the storage backend:
- If `bd` (beads-mcp) is installed → store as a bead with full ticket context in `notes`.
- Otherwise → store at `${MEDTASKER_TICKET_DIR:-./.todo}/<TICKET-ID>/TICKET_DESCRIPTION.md` (markdown file).

See `rules/ticket-storage.md` for both backends' schemas, field mappings, and re-fetch semantics.
See `rules/mcp-patterns.md` for retry policy, response validation, and credential extraction.

### 1. Read ticket context
Fetch ticket details (summary, description, acceptance criteria, status, assignee) to understand what needs to be done. When the ticket is already stored locally (from a prior inbox), read from there. Otherwise, fetch directly per `rules/ticket-storage.md` and store via the active backend.

### 2. Update ticket with work
After completing work, update the Jira ticket:
- Add a comment summarizing what was done
- Link the PR URL
- Transition the ticket status (e.g. In Progress → In Review)

### 3. Auto-link from branch names
Extract Jira ticket IDs from git branch names automatically. Common patterns:
- `fix/MT-9548_description` → `MT-9548`
- `feat/MT-1234-description` → `MT-1234`
- `MT-1234_description` → `MT-1234`
- `feature/PROJ-123-thing` → `PROJ-123`

## Workflow

### When invoked with `inbox` (`/medtasker-jira inbox MT-9548`)
1. Validate ticket key format (pattern: `[A-Z]+-\d+`)
2. **Detect the storage backend** (see `rules/ticket-storage.md`):
   ```bash
   if command -v bd >/dev/null 2>&1 && [ -z "$MEDTASKER_FORCE_FILESYSTEM" ]; then
       BACKEND=beads
   else
       BACKEND=filesystem
       TICKET_DIR="${MEDTASKER_TICKET_DIR:-./.todo}/<TICKET-ID>"
   fi
   ```
3. **Check for an existing record:**
   - **Beads:** `bd query "label=<TICKET-ID>" --sort updated --reverse`
   - **Filesystem:** test for `$TICKET_DIR/TICKET_DESCRIPTION.md`
   - If a record **already exists**:
     - Display the existing summary (title, status)
     - **Prompt the user:** "Pull fresh updates from Jira? [y/n]"
     - If **yes** → destructive clear per `rules/ticket-storage.md` (`bd update <id> --notes ""` OR `rm $TICKET_DIR/TICKET_DESCRIPTION.md`) and re-run steps 4–9, then jump to step 10
     - If **no** → display the existing summary and stop
   - If **no record exists** → continue with steps 4–10
4. Fetch the ticket via `mcp__mcp-atlassian__jira_get_issue` with `fields="*all"` and `expand="renderedFields"` — apply retry/validation per `rules/mcp-patterns.md`. Read credentials from `~/.claude/mcp.json` per `rules/mcp-patterns.md` (never hardcode).
5. Write the record using the active backend's schema in `rules/ticket-storage.md`:
   - **Beads:** `bd create --title="..." --label="jira,<TICKET-ID>,..." --priority=N`; then `bd update <id> --notes "$NOTES_BODY"`
   - **Filesystem:** `mkdir -p "$TICKET_DIR" && cat > "$TICKET_DIR/TICKET_DESCRIPTION.md" <<EOF ... EOF`
6. Reference attachment URLs in the record (don't download unless the user asks; large attachments stay in Jira)
7. Scan description and comments for Confluence URLs (`grep -oE 'https://[^/]+\.atlassian\.net/wiki/[^ )\]"]+'`) — if any, fetch each via `mcp__mcp-atlassian__confluence_get_page` and append to a "Related Confluence" section in the record
8. Scan for GitHub URLs (`grep -oE 'https://github\.com/[^ )\]"]+'`) — if any, fetch PR/issue details via `gh pr view --json …` or `gh issue view --json …` and append to a "Related GitHub" section
9. If using the filesystem backend, ensure `.todo/` (or the configured dir if inside the repo) is gitignored — these are local working notes, not source.
10. Display summary of what was fetched and what failed (per `rules/mcp-patterns.md` partial-context resilience), including which backend was used

### When invoked with `inbox update` (`/medtasker-jira inbox update MT-9548`)
1. Detect backend (same as above).
2. Find the existing record. Clear it destructively (per `rules/ticket-storage.md`) and re-run steps 4–8 of the `inbox` workflow.

### When invoked without arguments (`/medtasker-jira`)
1. Detect the current git branch
2. Extract ticket ID from branch name (pattern: `[A-Z]+-\d+`)
3. If found, fetch and display the ticket details
4. If not found, ask the user for a ticket ID

### When invoked with a ticket ID (`/medtasker-jira MT-9548`)
1. Fetch and display the ticket details

### When invoked with `update` (`/medtasker-jira update`)
1. Detect ticket ID from current branch
2. Gather context: recent commits, current PR (if any via `gh pr view`)
3. Post a comment to the ticket summarizing the work done
4. Ask user if they want to transition the ticket status

### When invoked with `link` (`/medtasker-jira link`)
1. Detect ticket ID from current branch
2. Find the current PR via `gh pr view --json url`
3. Add the PR URL as a comment on the Jira ticket
4. Format: `PR: <url> — <pr title>`

### When invoked with `start` (`/medtasker-jira start MT-9548`)
1. Transition the ticket to "In Progress"
2. Display the ticket details for context

## Extracting Ticket ID

Use this regex on the current git branch name:
```
([A-Z][A-Z0-9]+-\d+)
```

Run: `git branch --show-current` then extract the match.

## IMPORTANT: Always Fetch Full Ticket Context

When displaying any ticket, ALWAYS fetch and show the complete context automatically. Never require the user to ask for additional details. This includes:

### Attachments & Images
1. Fetch the ticket with `expand=renderedFields` to get attachment references
2. Download attachments via `jira_download_attachments` MCP tool (handles authentication automatically)
3. Large attachments (>50MB) remain in Jira — reference their URLs in bead notes
4. List attachments with filenames and sizes in bead notes

```bash
# Check if bead exists locally
bd query "label=MT-XXXX" --sort updated --reverse
```

Read credentials from `~/.claude/mcp.json`:
```bash
python3 -c "import json; c=json.load(open('$HOME/.claude/mcp.json'))['mcpServers']['mcp-atlassian']['env']; print(c['JIRA_USERNAME']); print(c['JIRA_API_TOKEN'])"
```

### Issue Links
Show all linked issues (blocks, is blocked by, relates to, etc.) with their key, summary, and status.

### Comments
Show the last 3 comments with author, date, and content.

### Subtasks
List all subtasks with key, summary, and status.

### Full Display Format

```
**<KEY>: <Summary>**

| Field | Value |
|-------|-------|
| **Status** | ... |
| **Type** | ... |
| **Assignee** | ... |
| **Priority** | ... |
| **Sprint** | ... |
| **Labels** | ... |

**Description:**
<rendered description text>

**Attachments:**
<display each image inline, list other files>

**Links:**
- blocks: MT-1234 - Summary [Status]
- relates to: MT-5678 - Summary [Status]

**Subtasks:**
- MT-9999 - Summary [Status]

**Recent Comments:**
[date] Author: comment text
```

## Comment Formatting

When posting comments to Jira, use **Markdown syntax** (not Jira wiki markup). The MCP server converts Markdown to Jira's ADF format automatically.

**Always use the medtasker-jira-markup skill** for reference on correct formatting.

### Quick Reference

| Element | Syntax | Example |
|---------|--------|---------|
| Heading | `## Title` | `## Code Update` |
| Bold | `**text**` | `**Branch:**` |
| Italic | `*text*` | `*branch_name*` |
| Code inline | `` `code` `` | `` `feature/MT-1234` `` |
| Link | `[text](url)` | `[PR Title](https://github.com/...)` |
| Bullet list | `- item` | `- First item` |
| Numbered list | `1. item` | `1. First step` |

### Example Comment

```markdown
## Code Update

**Branch:** `feature/MT-1234-fix`
**PR:** [PR Title](https://github.com/nimblic/medtasker-app/pull/97)

**Commits:**
1. Initial implementation
2. Added tests
3. Fixed review feedback

**Summary:** Brief description of what was done.
```

### Full Formatting Guide

For complete formatting reference, see the **medtasker-jira-markup skill**.

## Status Transitions

Common transitions to offer:
- **To Do → In Progress**: When starting work (`/medtasker-jira start`)
- **In Progress → In Review**: When PR is created (`/medtasker-jira link` or `/medtasker-jira update`)
- **In Review → Done**: When PR is merged

Always confirm with the user before transitioning status.

## Error Handling

- If MCP server is not connected, tell the user to fill in credentials in `~/.claude/mcp.json`
- If ticket not found, suggest checking the ticket ID
- If branch has no ticket ID pattern, ask the user to provide one

## Rules

Two rule files capture decisions that aren't derivable from the workflow above:

- `rules/mcp-patterns.md` — retry counts and delays per service, MCP response validation, credential extraction from `~/.claude/mcp.json`, sequential-processing invariant.
- `rules/ticket-storage.md` — backend auto-detect (beads-mcp if `bd` installed, else filesystem under `${MEDTASKER_TICKET_DIR:-./.todo}`), schemas for both, Jira→record field mapping, readiness contract for downstream consumers, re-fetch (destructive overwrite) semantics.

Confluence/GitHub link processing is handled inline in the inbox workflow (steps 7–8). Figma integration is opt-in only — never fetched automatically.
