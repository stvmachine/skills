# ADR-002: MCP Server Choices

Notes on which MCP servers we install and why. Less formal than a typical ADR — just enough to explain the choices to someone new.

## What we install

Two categories: **skill-bound** servers (declared in a specific skill's `mcp_servers:` frontmatter, installed only when that skill is) and **baseline** servers (always written to global `.mcp.json` regardless of installed skills).

### Skill-bound

| Server | How it runs | Auth | Used by |
|--------|-------------|------|---------|
| `mcp-atlassian` | `uvx mcp-atlassian --transport stdio` | env vars (`JIRA_*`, `CONFLUENCE_*`) | `medtasker-jira`, `medtasker-jira-ticket-transition` |
| `github` | `npx -y @modelcontextprotocol/server-github` | `GITHUB_TOKEN` env var | `medtasker-jira`, `medtasker-jira-ticket-transition` |
| `figma` | `npx -y @tmegit/figma-developer-mcp --stdio --json` | `FIGMA_API_KEY` env var | (opt-in; not in any default skill) |
| `context7` | HTTP to `https://mcp.context7.com/mcp` | `CONTEXT7_API_KEY` header | `medtasker-jira`, `medtasker-jira-ticket-transition` |

### Baseline

Currently empty. `internal/mcp.BaselineServers()` returns `[]ServerConfig{}` — no MCP currently earns a place in every user's global `.mcp.json` by default. The function is kept as a named extension point for any future general-purpose server that does.

dotenvx (ADR-001) puts the env vars into Claude Code's process; `.mcp.json` references them as `${VAR}`. The `install` command writes to both `~/.claude/.mcp.json` and `~/.claude.json` (the latter takes precedence).

## Why these specific ones

**Atlassian: `mcp-atlassian` (Python, via `uvx`).** The `@modelcontextprotocol/server-atlassian` npm package is an early shim with thin API coverage. `mcp-atlassian` is the real one. It's Python, so `uvx` instead of `npx`. The `--transport stdio` flag matters — it defaults to SSE otherwise. One process serves both Jira and Confluence, which is why we pass both sets of env vars.

**GitHub: stdio, not the Copilot HTTP endpoint.** GitHub publishes an HTTP MCP at `api.githubcopilot.com/mcp/` aimed at VS Code. Claude Code's HTTP MCP client does OAuth2 dynamic client registration (RFC 7591), which that endpoint doesn't support — you get `Incompatible auth server: does not support dynamic client registration` and it never falls back. The stdio server takes a plain `GITHUB_TOKEN` and works fine.

**Figma: `@tmegit/figma-developer-mcp`.** Picked for the `--stdio --json` flags that give clean structured output. Env var is `FIGMA_API_KEY` (not `FIGMA_ACCESS_TOKEN` — some third-party docs use that older name).

**Context7: HTTP with a custom header.** Unlike GitHub's HTTP endpoint, Context7 doesn't require OAuth dynamic client registration — the API key is just a custom HTTP header that Claude Code expands `${VAR}` into. The stdio alternative (`npx @upstash/context7-mcp --api-key <key>`) would force us to embed the key in the `args` array, where `${VAR}` expansion doesn't apply — meaning the plaintext key would land in `.mcp.json`. HTTP avoids that.

## The context7 concern

context7 is a third-party SaaS (Upstash). If it goes down, gets sunset, or starts charging in a way that doesn't work for us, every skill that relies on it loses library-doc lookups with no built-in fallback. Worth keeping an eye on, and worth knowing we have no offline path.

## Prerequisites this adds

- `uv` must be installed (for `uvx mcp-atlassian`). The `doctor` command should check for it.
- `npx` / Node.js for the GitHub and Figma servers.
- dotenvx (ADR-001) for env injection.
