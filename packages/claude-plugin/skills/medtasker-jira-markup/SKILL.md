---
name: medtasker-jira-markup
description: Correct Jira comment formatting for use with mcp-atlassian MCP server. Use when writing or editing Jira comments to ensure proper formatting renders correctly in Jira Cloud. Essential companion to the jira skill for properly formatted comments.
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
---

# Jira Comment Formatting Guide

## Important: MCP Server Behavior

The mcp-atlassian MCP server handles the conversion from Markdown to Jira's ADF (Atlassian Document Format). Based on observed behavior:

- The MCP accepts **Markdown** formatting in the `body` parameter
- It converts this to ADF internally for Jira Cloud
- **All standard Markdown syntax works correctly**

## Working Format Patterns

### Headings
Use markdown heading syntax:
```markdown
# Main Heading
## Section Heading  
### Subsection
```

### Bold and Italic
```markdown
**bold text**
*italic text*
***bold and italic***
```

### Lists

**Bullet lists** - use dashes:
```markdown
- Item one
- Item two
  - Nested item
- Item three
```

**Numbered lists** - use numbers:
```markdown
1. First step
2. Second step
3. Third step
```

### Code

**Inline code** - use backticks:
```markdown
Run `npm install` to install dependencies.
```

**Code blocks** - use triple backticks:
```markdown
```
function example() {
  return "hello";
}
```
```

### Links

**External links** - use markdown syntax:
```markdown
[Link text](https://example.com)
```

**Jira ticket links** - just use the ticket ID (auto-linked):
```markdown
See MT-1234 for related work.
```

### Blockquotes
```markdown
> This is a quoted block of text.
```

## Recommended Comment Template

```markdown
## Discovery Phase Complete - ADR Created

We have completed a quick discovery phase on the build strategy. 
An actionable plan has been documented in PR [ADR Legacy RN Build Strategy](https://github.com/...).

*Branch:* `chore/adr_legacy_rn_build_strategy`

### Summary of ADR-001:

The core challenge is that React Native 0.67 was designed for **Xcode 13.x**.

### Recommended Approach

- **Android builds**: Run on `ubuntu-latest` (fully supported)
- **iOS builds**: Run on a *self-hosted macOS runner* with Xcode 13.2.1
- **Tests and linting**: Run on `ubuntu-latest`

### Key Decisions

1. Self-hosted runner is required for iOS
2. Version pinning: Node 14.21.3, JDK 11, CocoaPods 1.12.1
3. Recommended hardware: Mac mini M1 (~$500-600 refurbished)

### Risks Identified

- Self-hosted runner is a single point of failure
- Node 14, Xcode 13 are EOL (*security debt*)
- Apple certificate expiration tracking needed
```

## Formatting Reference

| Element | Syntax | Result |
|---------|--------|--------|
| Heading 2 | `## Title` | Large heading |
| Heading 3 | `### Title` | Medium heading |
| Bold | `**text**` | **bold** |
| Italic | `*text*` | *italic* |
| Code inline | `` `code` `` | `code` |
| Code block | ` ```code``` ` | Code block |
| Link | `[text](url)` | [Link](url) |
| Bullet | `- item` | • item |
| Numbered | `1. item` | 1. item |

## Common Patterns for Jira Comments

### Code Update Comment

```markdown
## Code Update

**Branch:** `feature/MT-1234-fix`
**PR:** [PR Title](https://github.com/...)
**Commits:**
1. Initial implementation
2. Added tests
3. Fixed review feedback

**Summary:** Brief description of changes made.
```

### Discovery/Research Comment

```markdown
## Discovery Phase Complete

Completed investigation of MT-5678.

**Findings:**
1. Root cause identified in authentication module
2. Fix requires updating dependency to v2.1
3. Estimated effort: 2 days

**Next Steps:**
- Create implementation ticket
- Schedule for next sprint
```

### Status Update Comment

```markdown
## Status Update

**Progress:**
- Backend API changes - Done
- Frontend integration - In Progress
- Testing - Pending

**Blockers:**
1. Waiting for MT-9999 design approval
```

## Validation Checklist

Before submitting a Jira comment:

- [ ] Use markdown heading syntax (`##`, `###`)
- [ ] Use dashes (`-`) for bullet lists
- [ ] Use numbers (`1.`, `2.`) for ordered lists
- [ ] Use backticks (`` ` ``) for inline code
- [ ] Use triple backticks (`` ``` ``) for code blocks
- [ ] Use markdown link syntax (`[text](url)`) for PR links
- [ ] Use `**text**` for bold emphasis
- [ ] Use `*text*` for italic emphasis
