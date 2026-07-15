# MCP Interaction Patterns

How to talk to `mcp-atlassian` (and other MCP servers) from this skill. Covers retry policy, response validation, credential extraction, and the sequential-processing rule.

## Retry policy (per service)

| Service | Attempts | Backoff |
|---|---:|---|
| Jira (`mcp-atlassian`) | 3 | 30s, 60s, 120s |
| Confluence (`mcp-atlassian`) | 3 | 30s, 60s, 120s |
| Figma | 1 | 30s |

**Partial-context resilience:** if one MCP source fails after retries, **continue with the others**. Jira is the only required source — Confluence/Figma failures are logged and noted in the bead, not fatal.

```bash
retry_with_backoff() {
    local max_attempts="$1"; shift
    local delay=30 cmd="$*"
    for attempt in $(seq 1 "$max_attempts"); do
        if eval "$cmd"; then return 0; fi
        if (( attempt < max_attempts )); then
            sleep "$delay"; delay=$((delay * 2))
        fi
    done
    return 1
}
```

## Response validation

Every MCP response carries an `isError` flag. Check it before extracting any field.

```bash
# Returns 0 if response is OK, 1 if MCP errored
validate_mcp_response() {
    python3 -c "
import json, sys
data = json.load(sys.stdin)
if data.get('isError', False):
    print('MCP_ERROR: ' + str(data.get('error', 'unknown')), file=sys.stderr)
    sys.exit(1)
" <<< "$1"
}

# For Jira specifically, also require these critical fields
validate_jira_response() {
    validate_mcp_response "$1" || return 1
    python3 -c "
import json, sys
data = json.load(sys.stdin)
if 'key' not in data or not data.get('fields', {}).get('summary'):
    print('ERROR: Missing key or fields.summary', file=sys.stderr)
    sys.exit(1)
" <<< "$1"
}
```

For Confluence, require `id` and `body` instead.

## Credential extraction

Never hardcode credentials. Read them from `~/.claude/mcp.json` at call time:

```bash
get_jira_credentials() {
    python3 -c "
import json
with open('$HOME/.claude/mcp.json') as f:
    env = json.load(f)['mcpServers']['mcp-atlassian']['env']
    print(env['JIRA_USERNAME'])
    print(env['JIRA_API_TOKEN'])
"
}

IFS=$'\n' read -d '' -r JIRA_USERNAME JIRA_API_TOKEN < <(get_jira_credentials && printf '\0')
```

**Shell history hygiene:** when calling `curl` with credentials inline, **prefix the command with a leading space**. That keeps it out of bash/zsh history (assuming `HISTCONTROL=ignorespace` or equivalent):

```bash
 curl -s -L -u "$JIRA_USERNAME:$JIRA_API_TOKEN" "$JIRA_URL/rest/api/3/issue/$TICKET_ID"
```

Never include credentials in error messages or debug output. If validation fails, log the field name, not the value.

## Sequential processing

Process tickets **one at a time**. No parallel fetches. This is by design:

- One ticket at a time → no concurrent writes to the same bead.
- One source at a time per ticket → no concurrent MCP calls competing for the same retry budget.
- `mkdir -p` (used for attachment dirs, when applicable) is kernel-atomic.

No file locking is needed because no two operations ever overlap. If you find yourself reaching for `flock` or worrying about races, you've broken the sequential invariant — back up.
