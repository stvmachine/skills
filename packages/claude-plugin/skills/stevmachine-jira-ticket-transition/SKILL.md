---
name: stevmachine-jira-ticket-transition
description: Advance a Jira ticket through its lifecycle — ship to QA or send for peer review. Creates the PR if missing, transitions the Jira ticket to the chosen target, links the PR on the ticket, and updates the bead status. Use when the user says "ship", "ship to QA", "send to QA", "peer review", "send for review", "move to review", "ready for review", "create PR and update Jira", or wants to advance a branch through its lifecycle.
allowed-tools: Bash(git *), Bash(gh *), Bash(curl *), Bash(python3 *), Bash(bd *), mcp__mcp-atlassian__*, mcp__github__*
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

# Move Ticket

Single command to advance a feature branch through its lifecycle: commit (if needed), create PR, transition Jira to the chosen target, link the PR on the ticket, and update the bead status. The workflow is identical across targets — only Step 5 (transition match), Step 6 (comment header) and Step 7 (bead status) vary.

## Targets

| Target | Common phrasings | Transition keywords (case-insensitive substring match) | Local status after | Comment header |
|---|---|---|---|---|
| `review` | "peer review", "send for review", "move to review", "ready for review" | `Peer Review`, `In Review`, `Code Review` | `in_progress` (work isn't done — review may bounce back) | "Ready for peer review" |
| `qa` | "ship", "ship to QA", "send to QA" | `QA`, `Ready for QA` | `closed` (handed off to QA) | "Shipped to QA" |

**To add a new target:** add a row to this table. The rest of the workflow doesn't change. The dispatch happens at Step 5/6/7 by reading this table.

**Local status** applies to whichever ticket storage backend stevmachine-jira used (see stevmachine-jira's `rules/ticket-storage.md`):
- **Beads** (when `bd` — beads-mcp — is installed): `bd update <id> --status in_progress` or `bd close <id>`
- **Filesystem** (default fallback): edit the `**Status**` field in `${STEVMACHINE_TICKET_DIR:-./.todo}/<TICKET-ID>/TICKET_DESCRIPTION.md`

## Dependencies

This skill orchestrates three other skills — invoke them, don't duplicate their logic:
- **`/commit`** — used in Step 3 for properly formatted commits
- **`/stevmachine-jira`** — used in Step 5 for Jira transitions (and patterns for MCP usage, credential extraction)
- **`/stevmachine-jira-markup`** — used in Step 6 only, for formatting the comment that posts the PR link

## Pipeline Mode

Pipeline mode is an agent-to-agent invocation contract for cases where no human is available to answer prompts. The following overrides apply when activated.

**Pipeline mode is active when:** the invoking context provides `TICKET_ID`, `BRANCH_NAME`, `BASE_BRANCH`, `PR_BODY_FILE`, and `MOVE_TARGET` variables. The presence of `PR_BODY_FILE` is the signal.

**Pipeline mode overrides by step:**
- **Step 0 (Resolve Target):** Use `MOVE_TARGET` directly. If missing, FAIL with "MOVE_TARGET required in pipeline mode (qa | review)".
- **Step 1 (Gather Context):** Skip user confirmation. Use `TICKET_ID` directly. Skip `git status`/`git diff --stat` display. Use `BASE_BRANCH` directly.
- **Step 2 (Ensure Feature Branch):** Skip entirely — branch already exists and is checked out.
- **Step 3 (Commit):** Skip entirely — code is already committed.
- **Step 4 (Create PR):** Push branch with `git push -u origin HEAD`. Check for existing PR with `gh pr view --json url,title 2>/dev/null`. If PR exists, reuse without asking. If no PR, create with `gh pr create --base "$BASE_BRANCH" --title "<generated title>" --body-file "$PR_BODY_FILE"`. PR title follows gitmoji conventional commit format.
- **Step 4.5 (Wait for Cloudflare):** Apply the project detection logic. Spawn the background agent as normal. Still collect `CF_URL` before Step 6.
- **Step 5 (Transition Jira):** Get transitions via `mcp__mcp-atlassian__jira_get_transitions`. Match against the target's keywords (Targets table above). If zero matches, FAIL with "No {target} transition found for {TICKET_ID}". If 2+ matches, FAIL with "Ambiguous {target} transitions: {names}. Cannot auto-select in pipeline mode." Do NOT prompt the user.
- **Step 6 (Link PR):** Execute normally — include `CF_URL` in the comment if available. Use the target's comment header.
- **Step 7 (Update Bead):** Execute normally — set bead status per the Targets table.
- **Step 8 (Summary):** Execute normally.

## Workflow

### Step 0: Resolve Target

> In pipeline mode: read `MOVE_TARGET` from env and skip this step.

Determine which target the user wants from the invocation arguments:

- `/stevmachine-jira-ticket-transition qa [TICKET-ID]` → `target = qa`
- `/stevmachine-jira-ticket-transition review [TICKET-ID]` → `target = review`
- `/stevmachine-jira-ticket-transition [TICKET-ID]` (no target) → prompt the user: "Move to which state? [qa | review]"

Reject any other target with: "Unknown target '<value>'. Valid targets: qa, review."

Look up the target row in the Targets table — that gives the transition keywords, bead status, and comment header used in later steps.

### Step 1: Gather Context & Identify Ticket

> In pipeline mode: use the provided `TICKET_ID` directly and skip substeps 1–3 + 5–6.

1. Run `git branch --show-current` to get current branch
2. Extract Jira ticket ID from branch name using pattern `([A-Z][A-Z0-9]+-\d+)`
3. Check for an existing ticket record (acceptance criteria for the PR body live here):
   - **Beads:** `bd query "label=MT-XXXX" --sort updated --reverse`
   - **Filesystem:** test for `${STEVMACHINE_TICKET_DIR:-./.todo}/MT-XXXX/TICKET_DESCRIPTION.md`
4. **Resolve the ticket ID:**
   - If a ticket ID was provided as argument, use it
   - If the branch has a ticket ID, use it
   - Otherwise, list open ticket records — beads: `bd query "label=jira" --status open`; filesystem: `ls ${STEVMACHINE_TICKET_DIR:-./.todo}/` and read each `**Status**` field. If exactly one is open, confirm; if multiple, ask; if zero, ask for the ticket ID directly.
5. Run `git status` and `git diff --stat` to see what's pending
6. Run `git log --oneline master..HEAD` to see all commits on this branch
7. Determine the base branch — use `master` unless the branch name starts with `release/` (then find the parent release branch)

### Step 2: Ensure Feature Branch

> In pipeline mode: skip this step entirely.

**NEVER push directly to protected branches:** `master`, `main`, `dev`, or `release/*`.

1. Check if the current branch matches `master`, `main`, `dev`, or `release/*`
2. If on a protected branch, create a feature branch and move commits:
   - Determine branch prefix from commit type: `feat/`, `fix/`, `refactor/`, `chore/`, etc.
   - Build branch name: `<prefix>/<TICKET-ID>_<short_description>` (e.g. `feat/MT-9561_filter_coordinator_roles`)
   - If there are unpushed commits on the protected branch:
     ```bash
     git checkout -b <feature-branch>
     git branch -f <protected-branch> origin/<protected-branch>
     ```
   - If there are no unpushed commits, just create and switch: `git checkout -b <feature-branch>`
3. If already on a feature branch, continue as-is

### Step 3: Commit (if needed)

> In pipeline mode: skip this step entirely.

If there are uncommitted changes, invoke the `/commit` skill with these constraints:
- Do NOT commit files that are local config (e.g. `values.js` with local URLs, `.env`, credentials)
- Ask the user what to include if unclear

If working tree is clean, skip to Step 4.

### Step 4: Create GitHub PR

> In pipeline mode: use `PR_BODY_FILE` with `--body-file` flag instead of inline `--body`. Reuse existing PR template if present without asking.

Try `gh` CLI first. If `gh` is not available or fails, fall back to GitHub MCP tools.

**Check for PR Template:**
1. Look for `.github/pull_request_template.md` in the repo
2. If it exists, use its structure for the PR body
3. If no template exists, use the default format below

**Option A — gh CLI (preferred):**
1. Check if a PR already exists: `gh pr view --json url,title 2>/dev/null`
2. If PR exists, show it and skip to Step 4.5
3. If no PR, push the branch: `git push -u origin HEAD`
4. Create the PR with appropriate body.

**Default PR body (when no template exists):**
```markdown
## Summary
<!-- Brief description of what this PR does and why -->

## Description
<!-- Optional: Detailed explanation of changes -->

## JIRA Ticket
[TICKET-ID](https://example.atlassian.net/browse/TICKET-ID)

## ACs
<!-- Acceptance Criteria from the ticket -->
- [ ] <AC 1>
- [ ] <AC 2>
```

**Example gh command:**
```bash
gh pr create --base <base-branch> --title "<gitmoji> <type>(<scope>): <short description>" --body "$(cat <<'EOF'
## Summary
<brief description>

## Description
<detailed description>

## JIRA Ticket
[TICKET-ID](https://example.atlassian.net/browse/TICKET-ID)

## ACs
- [ ] <acceptance criteria from ticket, if available>
EOF
)"
```

**Option B — GitHub MCP (fallback if gh fails):**
1. Push the branch first: `git push -u origin HEAD`
2. Get repo owner/name from: `git remote get-url origin`
3. Check for existing PR: `mcp__github__list_pull_requests` with head=`<branch>` and base=`<base-branch>`
4. If no PR, create one: `mcp__github__create_pull_request` with title, body, head, base

**PR content rules:**
- PR title follows the same gitmoji conventional commit format as commits
- If a local ticket record exists (bead or `${STEVMACHINE_TICKET_DIR:-./.todo}/<TICKET-ID>/TICKET_DESCRIPTION.md`), read it for acceptance criteria to populate the PR body
- Base branch: use `master` by default, or the release branch if on a release/* sub-branch
- Respect the repository's PR template structure if one exists

### Step 4.5: Wait for Cloudflare Deployment (your-project only)

After the PR is created, check whether this is the your-project project:

```bash
git remote get-url origin
```

If the remote URL does NOT contain `your-project`, skip this step and set `CF_URL=""`. Continue to Step 5.

If the remote URL contains `your-project`, **delegate the wait to a background agent** — do NOT block the main agent. Spawn a `general-purpose` subagent with the following instructions:

> Poll `gh pr checks <PR_NUMBER> --json name,state,targetUrl` every 15 seconds for up to 10 minutes (40 attempts). Return the first `targetUrl` from a check whose name contains "cloudflare" (case-insensitive). If no URL is found after 40 attempts, return an empty string.

The main agent proceeds to Step 5 while the background agent waits. Before executing Step 6, retrieve the result and store it as `CF_URL`. Useful for both review (reviewer clicks the preview) and QA (tester clicks the preview).

### Step 5: Transition Jira Ticket

> In pipeline mode: fail hard on zero or ambiguous matches instead of asking. See Pipeline Mode section above.

1. Get available transitions for the ticket via `mcp__mcp-atlassian__jira_get_transitions`
2. Look up the target's keywords from the Targets table at the top of this file
3. Find a transition whose name contains any of those keywords (case-insensitive substring match)
4. If multiple match, show options and ask the user
5. If none match, list all available transitions and ask the user which to use
6. Execute the transition via `mcp__mcp-atlassian__jira_transition_issue`

### Step 6: Link PR on Jira Ticket

Use the `/stevmachine-jira-markup` skill's formatting rules and post the comment via `mcp__mcp-atlassian__jira_add_comment` — **Markdown syntax, not Jira wiki markup**. The mcp-atlassian server converts Markdown to ADF automatically.

The comment should include:
- The target's **Comment header** from the Targets table (e.g. "Ready for peer review" for `review`, "Shipped to QA" for `qa`)
- PR title and URL
- Branch name
- Base branch
- Cloudflare preview URL (only if `CF_URL` is non-empty from Step 4.5)

### Step 7: Update Local Ticket Status

Look up the target's **Local status after** from the Targets table. Apply it via whichever backend stored this ticket (see stevmachine-jira's `rules/ticket-storage.md`):

**Beads** (when `bd` is installed):
```bash
bd query "label=<TICKET-ID>" --sort updated --reverse
# for review:
bd update <bead-id> --status in_progress
# for qa:
bd close <bead-id> --reason "Shipped to QA — PR: <PR_URL>"
```

- **Filesystem** (when no beads-mcp, or `STEVMACHINE_FORCE_FILESYSTEM=1`):
```bash
FILE="${STEVMACHINE_TICKET_DIR:-./.todo}/<TICKET-ID>/TICKET_DESCRIPTION.md"
# Update the **Status** line in the metadata table to in_progress or closed.
# Example with sed (BSD/macOS):
sed -i '' -E 's/(\*\*Status\*\*[[:space:]]*\|).*/\1 in_progress |/' "$FILE"
# For qa (closed), also append a "Shipped" section to the file:
cat >> "$FILE" <<EOF

## Shipped
- **PR**: <PR_URL>
- **Date**: $(date -u +%Y-%m-%d)
- **Branch**: $(git branch --show-current)
EOF
```

If no record is found in either backend, skip silently — do NOT fail. If the record is already at the target status, this is a no-op.

### Step 8: Summary

Display a summary:
```
Moved <TICKET-ID> to <target>:
  PR: <PR URL>
  Jira: transitioned to <new status>
  Comment: linked PR on ticket
  Local: <new local status> (or "no record found — skipped")  [backend: beads | filesystem]
```

## Invocation Patterns

- `/stevmachine-jira-ticket-transition qa` — ship to QA, auto-detect ticket from branch
- `/stevmachine-jira-ticket-transition qa MT-1234` — ship to QA for specific ticket
- `/stevmachine-jira-ticket-transition review` — send for peer review, auto-detect ticket
- `/stevmachine-jira-ticket-transition review MT-1234` — send for peer review for specific ticket
- `/stevmachine-jira-ticket-transition` — prompt the user for target, then auto-detect ticket

## Error Handling

- If `gh` CLI not authenticated: tell user to run `gh auth login`
- If `gh auth status` shows invalid/expired credentials: tell the user to either re-run `gh auth login` or export a fresh token with `export GH_TOKEN=<fresh_token>`
- If Jira MCP not connected: tell user to check `~/.claude/mcp.json` credentials
- If no matching transition available: show available transitions and ask user which to use
- If PR creation fails (e.g. no upstream): push first, then retry
- If in pipeline mode and transition is ambiguous: FAIL with descriptive error (do not prompt)
- If in pipeline mode and PR creation fails: FAIL with error (do not retry interactively)
- If target is unknown: FAIL with "Unknown target '<value>'. Valid targets: qa, review."
