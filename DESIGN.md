# Medtasker Skills Distribution Repository — Design Document

## 1. Overview

This document specifies the architecture for a **skill distribution repository** that packages Medtasker skills for distribution across multiple AI agent platforms (Claude Code, OpenCode, and others). It covers repository structure, installation mechanisms, MCP encapsulation, subagent orchestration, dependency resolution, multi-platform support, and security.

### Goals

- Provide a **single source of truth** for Medtasker skills
- Enable **one-command installation** of skills and their dependencies
- **Auto-install** infrastructure tools (`beads`, `rtk`) when missing
- **Hide MCP complexity** from end users via skill-embedded configuration
- Support **multiple agent platforms** with platform-specific manifests
- Ensure **secure credential handling** and dependency resolution

---

## 2. Repository Structure

```
medtasker-skills/
├── README.md                          # Project overview and quick start
├── DESIGN.md                          # This document
├── LICENSE
├──
├── packages/                          # Platform-specific packages
│   ├── claude-plugin/                 # Claude Code plugin package
│   │   ├── .claude-plugin/
│   │   │   ├── plugin.json            # Plugin manifest
│   │   │   └── marketplace.json       # Marketplace catalog
│   │   ├── skills/                    # Skills for Claude Code
│   │   │   ├── medtasker-jira/
│   │   │   │   └── SKILL.md
│   │   │   ├── medtasker-jira-markup/
│   │   │   │   └── SKILL.md
│   │   │   ├── medtasker-jira-ticket-transition/
│   │   │   │   └── SKILL.md
│   │   ├── commands/                  # Flat command skills (optional)
│   │   ├── agents/                    # Subagent definitions
│   │   ├── hooks/                     # Hook configurations
│   │   └── .mcp.json                  # MCP server configurations
│   │
│   ├── opencode-package/              # OpenCode-compatible package
│   │   ├── skills/                    # Skills with YAML frontmatter
│   │   │   ├── medtasker-jira/
│   │   │   │   ├── SKILL.md           # With embedded MCP config
│   │   │   │   └── rules/
│   │   │   │       ├── fetch.md
│   │   │   │       ├── update.md
│   │   │   │       ├── mcp-utils.md
│   │   │   │       └── ...
│   │   │   ├── medtasker-jira-markup/
│   │   │   │   └── SKILL.md
│   │   │   ├── medtasker-jira-ticket-transition/
│   │   │   │   └── SKILL.md
│   │   └── manifest.json              # OpenCode manifest
│   │
│   └── skillfish-package/             # skill.fish compatible package
│       ├── skillfish.json             # Manifest for skillfish CLI
│       └── skills/
│           ├── medtasker-jira/
│           │   └── SKILL.md
│           ├── medtasker-jira-markup/
│           │   └── SKILL.md
│           ├── medtasker-jira-ticket-transition/
│           │   └── SKILL.md
│
├── src/                               # Source code for installer CLI
│   ├── installer/
│   │   ├── __init__.py
│   │   ├── main.py                    # Entry point
│   │   ├── platforms/
│   │   │   ├── __init__.py
│   │   │   ├── base.py                # Abstract platform interface
│   │   │   ├── claude_code.py         # Claude Code platform
│   │   │   ├── opencode.py            # OpenCode platform
│   │   │   └── skillfish.py           # skill.fish platform
│   │   ├── dependency_resolver.py     # Dependency graph resolver
│   │   ├── mcp_manager.py             # MCP configuration manager
│   │   ├── infrastructure.py          # beads/rtk installer
│   │   └── credentials.py             # Secure credential handling
│   │
│   └── tests/                         # Test suite (TDD)
│       ├── test_dependency_resolver.py
│       ├── test_mcp_manager.py
│       ├── test_platforms.py
│       └── test_credentials.py
│
├── scripts/
│   ├── install.sh                     # One-line shell installer
│   ├── install.ps1                    # PowerShell installer (Windows)
│   └── bootstrap.py                   # Python bootstrap script
│
├── configs/
│   ├── mcp-templates/                 # MCP configuration templates
│   │   ├── mcp-atlassian.json         # Jira/Confluence MCP template
│   │   ├── github.json                # GitHub MCP template
│   │   └── figma.json                 # Figma MCP template
│   └──
│
└── docs/
    ├── INSTALL.md                     # Installation guide
    ├── CONTRIBUTING.md                # Contribution guidelines
    ├── SECURITY.md                    # Security policy
    └── PLATFORM-GUIDE.md              # Platform-specific guides
```

### Key Design Decisions

1. **Platform-specific packages**: Each target platform (Claude Code, OpenCode, skill.fish) gets its own package with the appropriate manifest format and directory structure. This avoids translation layers and ensures native compatibility.

2. **Shared skill source**: Skills are authored once in a canonical format and packaged into platform-specific formats during build. The canonical format uses YAML frontmatter with embedded MCP config (OpenCode style) as the source of truth.

3. **Build pipeline**: A build script (`scripts/build.py`) transforms canonical skills into platform packages:
   - Claude Code: Converts YAML frontmatter to `plugin.json` + `.mcp.json`
   - OpenCode: Passes through as-is
   - skill.fish: Generates `skillfish.json` manifest

---

## 3. Installation Mechanism Specification

### 3.1 Installer (clone + run)

The repo is private; anonymous HTTPS to `github.com/nimblic/medtasker-skills` returns 404, so the historical `curl | bash` and `go install …@latest` flows do not work. The installer is invoked from a fresh clone:

```bash
# macOS / Linux
git clone git@github.com:nimblic/medtasker-skills.git
cd medtasker-skills && ./scripts/install.sh
```

Windows isn't supported by `install.sh`; build the CLI manually from a cloned repo (`go build -o <PATH-dir>/medtasker-skills ./cmd/medtasker-skills`).

### 3.2 Installer Behavior

The installer performs the following steps:

1. **Detect Platform**: Identify which agent platforms are installed:
   - Claude Code: Check for `~/.claude/` directory
   - OpenCode: Check for `~/.opencode/` or `opencode` command
   - Other: Check for `~/.agents/skills/` (medtasker)

2. **Auto-Install Infrastructure**:
   ```bash
   # Check for beads
   if ! command -v bd &> /dev/null; then
       echo "Installing beads..."
       curl -fsSL https://raw.githubusercontent.com/gastownhall/beads/main/integrations/beads-mcp/install.sh | bash
   fi

   # Check for rtk
   if ! command -v rtk &> /dev/null; then
       echo "Installing rtk..."
       brew install rtk
   fi
   ```

3. **Install Skills**:
   - Clone the repository to `~/.medtasker-skills/`
   - Install platform-appropriate packages to detected directories
   - Resolve and install dependencies

4. **Configure MCP**:
   - Detect existing MCP configurations
   - Prompt for missing credentials
   - Write secure MCP config files

### 3.3 CLI Tool: `medtasker-skills`

After installation, users interact with the system via the `medtasker-skills` CLI:

```bash
# Install all Medtasker skills
medtasker-skills install

# Install specific skill
medtasker-skills install medtasker-jira

# Update all skills
medtasker-skills update

# List installed skills
medtasker-skills list

# Check dependencies
medtasker-skills deps --tree

# Verify installation
medtasker-skills doctor
```

### 3.4 Installation Flow Diagram

```
User runs install.sh
    │
    ▼
┌─────────────────┐
│ Detect Platform │──► Claude Code? ──► Install to ~/.claude/plugins/
│                 │──► OpenCode? ─────► Install to ~/.opencode/skills/
│                 │──► medtasker? ────► Install to ~/.agents/skills/
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ Check Infra     │──► beads installed? ──► curl -fsSL https://raw.githubusercontent.com/gastownhall/beads/main/integrations/beads-mcp/install.sh | bash
│                 │──► rtk installed? ────► brew install rtk
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ Resolve Deps    │──► Build dependency graph
│                 │──► Topological sort
│                 │──► Install in order
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ Configure MCP   │──► Detect existing configs
│                 │──► Prompt for missing credentials
│                 │──► Write secure configs
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ Verify          │──► Run tests
│                 │──► Check skill loading
└─────────────────┘
```

---

## 4. MCP Encapsulation Strategy

### 4.1 Embedded MCP Configuration

Skills bundle MCP server configurations directly in their YAML frontmatter. This hides MCP complexity from end users — they install a skill and get the MCP capabilities automatically.

#### Canonical Format (OpenCode Style)

```yaml
---
name: medtasker-jira
description: Connect code work to Jira tickets
mcp_servers:
  - name: mcp-atlassian
    type: stdio
    command: npx
    args:
      - -y
      - @modelcontextprotocol/server-atlassian
    env:
      JIRA_USERNAME: ${JIRA_USERNAME}
      JIRA_API_TOKEN: ${JIRA_API_TOKEN}
      CONFLUENCE_URL: ${CONFLUENCE_URL}
permissions:
  mcp: allow
---
```

#### Claude Code Format

The build pipeline transforms the canonical format into Claude Code's `plugin.json` + `.mcp.json`:

```json
// .claude-plugin/plugin.json
{
  "name": "medtasker-jira",
  "version": "1.0.0",
  "description": "Connect code work to Jira tickets",
  "mcpServers": "./.mcp.json"
}
```

```json
// .mcp.json
{
  "mcpServers": {
    "mcp-atlassian": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-atlassian"],
      "env": {
        "JIRA_USERNAME": "${JIRA_USERNAME}",
        "JIRA_API_TOKEN": "${JIRA_API_TOKEN}"
      }
    }
  }
}
```

### 4.2 Lazy Loading and Session Isolation

Following OhMyOpenCode's pattern:

1. **Lazy Loading**: MCP servers are only started when a skill's MCP tool is first invoked
2. **Session Isolation**: Each agent session gets its own MCP server instance
3. **Cleanup**: MCP servers are stopped when the session ends

```python
# Conceptual implementation
class SkillMcpManager:
    def __init__(self):
        self._servers = {}  # session_id -> {mcp_name -> process}
    
    def get_or_start_server(self, session_id: str, mcp_config: dict):
        """Lazy load MCP server for session"""
        if session_id not in self._servers:
            self._servers[session_id] = {}
        
        mcp_name = mcp_config['name']
        if mcp_name not in self._servers[session_id]:
            # Start new process
            proc = subprocess.Popen(
                [mcp_config['command']] + mcp_config['args'],
                env={**os.environ, **mcp_config.get('env', {})}
            )
            self._servers[session_id][mcp_name] = proc
        
        return self._servers[session_id][mcp_name]
    
    def cleanup_session(self, session_id: str):
        """Stop all MCP servers for session"""
        if session_id in self._servers:
            for proc in self._servers[session_id].values():
                proc.terminate()
            del self._servers[session_id]
```

### 4.3 Unified Gateway: `skill_mcp` Tool

Skills use a unified `skill_mcp` tool to invoke MCP operations, hiding the underlying MCP complexity:

```python
# Tool definition
skill_mcp_tool = {
    "name": "skill_mcp",
    "description": "Invoke MCP server operations from skill-embedded MCPs",
    "parameters": {
        "mcp_name": "Name of the MCP server from skill config",
        "tool_name": "MCP tool to call (optional)",
        "resource_name": "MCP resource URI to read (optional)",
        "prompt_name": "MCP prompt to get (optional)",
        "arguments": "JSON string or object of arguments"
    }
}
```

Usage in skills:
```markdown
To fetch a Jira ticket, use:
```
skill_mcp(mcp_name="mcp-atlassian", tool_name="jira_get_issue", 
          arguments='{"issue_key": "MT-1234"}')
```
```

### 4.4 Environment Variable Expansion

MCP configurations use `${VAR}` syntax for environment variables:

```yaml
mcp_servers:
  - name: mcp-atlassian
    env:
      JIRA_USERNAME: ${JIRA_USERNAME}
      JIRA_API_TOKEN: ${JIRA_API_TOKEN}
```

The installer resolves these at install time or runtime:
- **Install time**: Prompt user for values, write to `~/.medtasker-skills/.env`
- **Runtime**: Load from environment or `.env` file before starting MCP server

---

## 5. Subagent Orchestration Patterns

### 5.1 Skill-Initiated Subagents

Skills spawn subagents for external service interactions. The orchestrator skill (`medtasker-jira-ticket-transition`) demonstrates this pattern:

```markdown
# medtasker-jira-ticket-transition dependencies
This skill orchestrates three other skills:
- `/commit` — for creating properly formatted commits
- `/medtasker-jira` — for Jira transitions and linking
- `/medtasker-jira-markup` — for Jira comment formatting
```

### 5.2 Subagent Spawning Pattern

```python
# Conceptual subagent spawn
def spawn_subagent(skill_name: str, task: str, context: dict):
    """Spawn a subagent with the specified skill loaded"""
    return task(
        category="quick",
        load_skills=[skill_name],
        prompt=f"""
        You are a subagent with the {skill_name} skill loaded.
        
        Task: {task}
        
        Context:
        {json.dumps(context, indent=2)}
        """
    )
```

### 5.3 Pipeline Mode

For agent-to-agent invocation (no human available):

```markdown
## Pipeline Mode

When invoked from the work executor's `rules/integration.md`:

**Pipeline mode is active when:** The invoking context provides `TICKET_ID`, 
`BRANCH_NAME`, `BASE_BRANCH`, and `PR_BODY_FILE` variables directly.

**Overrides:**
- Skip user confirmations
- Use provided variables directly
- Fail hard on ambiguous states (do not prompt)
```

### 5.4 Background Agent Delegation

For long-running operations (e.g., waiting for Cloudflare deployment):

```python
# Spawn background agent for polling
def spawn_background_agent(instructions: str, timeout: int = 600):
    """Delegate long-running task to background agent"""
    return task(
        run_in_background=True,
        prompt=instructions
    )

# Usage in skill
background_agent = spawn_background_agent("""
Poll `gh pr checks <PR_NUMBER> --json name,state,targetUrl` 
every 15 seconds for up to 10 minutes.
Return the first targetUrl from a check whose name contains "cloudflare".
""")

# Main agent continues with other work...
# Later, retrieve result before Step 6
cf_url = background_agent.get_result()
```

---

## 6. Dependency Resolution

### 6.1 Dependency Declaration

Skills declare dependencies in their YAML frontmatter:

```yaml
---
name: medtasker-jira-ticket-transition
description: Advance a Jira ticket — ship to QA or send for peer review
dependencies:
  - name: commit
    version: ">=1.0.0"
    required: true
  - name: medtasker-jira
    version: ">=1.0.0"
    required: true
  - name: medtasker-jira-markup
    version: ">=1.0.0"
    required: true
---
```

### 6.2 Dependency Graph Resolution

```python
class DependencyResolver:
    def __init__(self):
        self.graph = {}  # skill_name -> [dependencies]
    
    def add_skill(self, name: str, dependencies: list):
        """Add skill and its dependencies to graph"""
        self.graph[name] = dependencies
    
    def resolve(self, target_skill: str) -> list:
        """Return topologically sorted installation order"""
        visited = set()
        order = []
        
        def visit(skill):
            if skill in visited:
                return
            visited.add(skill)
            for dep in self.graph.get(skill, []):
                if dep['required']:
                    visit(dep['name'])
            order.append(skill)
        
        visit(target_skill)
        return order
    
    def detect_cycles(self) -> list:
        """Detect circular dependencies"""
        # Implementation using DFS
        pass
```

### 6.3 Installation Order Example

For `medtasker-jira-ticket-transition`:

```
medtasker-jira-ticket-transition
├── commit (required)
├── medtasker-jira (required)
│   └── medtasker-jira-markup (required)

Installation order:
1. commit
2. medtasker-jira-markup
3. medtasker-jira
4. medtasker-jira-ticket-transition
```

### 6.4 Version Resolution

- **Exact version**: `version: "1.0.0"`
- **Semver range**: `version: ">=1.0.0 <2.0.0"`
- **Latest**: `version: "*"` or omit version field
- **Git ref**: `version: "main"` or `version: "v1.0.0"`

---

## 7. Multi-Agent Platform Support

### 7.1 Platform Detection Matrix

| Platform | Directory | Manifest File | Discovery |
|----------|-----------|---------------|-----------|
| Claude Code | `~/.claude/skills/` or `~/.claude/plugins/` | `plugin.json` | Auto-scan on startup |
| OpenCode | `~/.opencode/skills/` or `.opencode/skills/` | None (auto-discover) | Auto-scan on startup |
| Medtasker | `~/.agents/skills/` | None | Auto-scan on startup |
| skill.fish | `~/skillfish/skills/` or `./skills/` | `skillfish.json` | CLI command |

### 7.2 Platform-Specific Installation

```python
class PlatformInstaller(ABC):
    @abstractmethod
    def install_skill(self, skill_path: str, target_dir: str):
        pass
    
    @abstractmethod
    def configure_mcp(self, mcp_config: dict):
        pass

class ClaudeCodeInstaller(PlatformInstaller):
    def install_skill(self, skill_path: str, target_dir: str):
        # Copy skill to ~/.claude/skills/ or plugin directory
        shutil.copytree(skill_path, target_dir)
    
    def configure_mcp(self, mcp_config: dict):
        # Write to ~/.claude/.mcp.json or plugin's .mcp.json
        pass

class OpenCodeInstaller(PlatformInstaller):
    def install_skill(self, skill_path: str, target_dir: str):
        # Copy skill to ~/.opencode/skills/
        shutil.copytree(skill_path, target_dir)
    
    def configure_mcp(self, mcp_config: dict):
        # MCP is embedded in skill YAML frontmatter
        pass

class SkillfishInstaller(PlatformInstaller):
    def install_skill(self, skill_path: str, target_dir: str):
        # skill.fish handles installation
        subprocess.run(['skillfish', 'add', skill_path])
    
    def configure_mcp(self, mcp_config: dict):
        # skill.fish doesn't support MCP yet
        pass
```

### 7.3 Cross-Platform Skill Format

The canonical skill format uses YAML frontmatter that all platforms can parse:

```yaml
---
name: medtasker-jira
description: Connect code work to Jira tickets
platforms:
  - claude-code
  - opencode
  - skillfish
mcp_servers:
  - name: mcp-atlassian
    type: stdio
    command: npx
    args: [-y, @modelcontextprotocol/server-atlassian]
dependencies:
  - name: medtasker-jira-markup
    required: true
---
```

Build pipeline transforms this into platform-specific formats.

---

## 8. Security Considerations

### 8.1 Credential Handling

1. **Never hardcode credentials** in skill files
2. **Use environment variables** with `${VAR}` syntax
3. **Store credentials securely**:
   - macOS: Keychain
   - Linux: Secret Service API / keyring
   - Windows: Credential Manager

```python
class CredentialManager:
    def __init__(self):
        self.backend = self._detect_backend()
    
    def _detect_backend(self):
        if sys.platform == 'darwin':
            return 'keychain'
        elif sys.platform == 'linux':
            return 'secret_service'
        elif sys.platform == 'win32':
            return 'windows_credential'
    
    def store(self, service: str, username: str, password: str):
        """Store credential in platform-native store"""
        if self.backend == 'keychain':
            subprocess.run([
                'security', 'add-generic-password',
                '-s', service,
                '-a', username,
                '-w', password
            ])
    
    def retrieve(self, service: str, username: str) -> str:
        """Retrieve credential from platform-native store"""
        if self.backend == 'keychain':
            result = subprocess.run([
                'security', 'find-generic-password',
                '-s', service,
                '-a', username,
                '-w'
            ], capture_output=True, text=True)
            return result.stdout.strip()
```

### 8.2 MCP Server Security

1. **Validate MCP server commands** before execution
2. **Sandbox MCP processes** where possible
3. **Limit MCP tool permissions** via `allowed-tools` in skill frontmatter
4. **Audit MCP calls** in logs (without exposing credentials)

### 8.3 Dependency Security

1. **Pin dependencies** to specific versions or git SHAs
2. **Verify dependency integrity** via checksums
3. **Scan for known vulnerabilities** in dependencies
4. **Isolate optional dependencies** — don't fail if optional deps are unavailable

### 8.4 Installation Security

1. **Verify installer integrity** via checksum
2. **Use HTTPS for all downloads**
3. **Validate skill signatures** if available
4. **Principle of least privilege**: Only install to user directories, never system-wide

### 8.5 Audit Logging

```python
class AuditLogger:
    def log_install(self, skill: str, version: str, platform: str):
        """Log skill installation"""
        self._write_log({
            'event': 'install',
            'skill': skill,
            'version': version,
            'platform': platform,
            'timestamp': datetime.utcnow().isoformat(),
            'user': getpass.getuser()
        })
    
    def log_mcp_access(self, skill: str, mcp_name: str, tool: str):
        """Log MCP access (without credentials)"""
        self._write_log({
            'event': 'mcp_access',
            'skill': skill,
            'mcp': mcp_name,
            'tool': tool,
            'timestamp': datetime.utcnow().isoformat()
        })
```

---

## 9. Testing Strategy (TDD)

### 9.1 Test Pyramid

```
         /\
        /  \
       / E2E \      # End-to-end: Install full stack, verify skills work
      /--------\
     / Integration \ # Integration: Test platform installers, MCP manager
    /--------------\
   /    Unit        \ # Unit: Test dependency resolver, credential manager
  /------------------\
```

### 9.2 Unit Tests

```python
# test_dependency_resolver.py
class TestDependencyResolver:
    def test_simple_resolution(self):
        resolver = DependencyResolver()
        resolver.add_skill('A', [{'name': 'B', 'required': True}])
        resolver.add_skill('B', [])
        
        order = resolver.resolve('A')
        assert order == ['B', 'A']
    
    def test_circular_dependency_detection(self):
        resolver = DependencyResolver()
        resolver.add_skill('A', [{'name': 'B', 'required': True}])
        resolver.add_skill('B', [{'name': 'A', 'required': True}])
        
        with pytest.raises(CircularDependencyError):
            resolver.resolve('A')
    
    def test_optional_dependency(self):
        resolver = DependencyResolver()
        resolver.add_skill('A', [
            {'name': 'B', 'required': True},
            {'name': 'C', 'required': False}
        ])
        resolver.add_skill('B', [])
        
        order = resolver.resolve('A')
        assert order == ['B', 'A']  # C is optional, not included
```

### 9.3 Integration Tests

```python
# test_platforms.py
class TestClaudeCodeInstaller:
    def test_install_skill(self, tmp_path):
        installer = ClaudeCodeInstaller()
        skill_path = tmp_path / 'test-skill'
        skill_path.mkdir()
        (skill_path / 'SKILL.md').write_text('---\nname: test\n---\n')
        
        target = tmp_path / 'claude-skills' / 'test'
        installer.install_skill(str(skill_path), str(target))
        
        assert target.exists()
        assert (target / 'SKILL.md').exists()
```

### 9.4 E2E Tests

```bash
#!/bin/bash
# test_e2e.sh

set -e

# Setup temp environment
export HOME=$(mktemp -d)

# Run installer
./scripts/install.sh

# Verify skills installed
medtasker-skills list | grep medtasker-jira
medtasker-skills list | grep medtasker-jira-ticket-transition

# Verify dependencies resolved
medtasker-skills deps --tree | grep "commit"
medtasker-skills deps --tree | grep "medtasker-jira-markup"

# Verify infrastructure installed
which bd
which rtk

# Cleanup
rm -rf $HOME
```

---

## 10. Implementation Roadmap

### Phase 1: Foundation (Week 1)
- [ ] Set up repository structure
- [ ] Implement dependency resolver with tests
- [ ] Create build pipeline for platform packages
- [ ] Write shell installer script

### Phase 2: Platform Support (Week 2)
- [ ] Implement Claude Code plugin package
- [ ] Implement OpenCode skill package
- [ ] Implement skill.fish compatibility
- [ ] Add platform detection logic

### Phase 3: MCP Integration (Week 3)
- [ ] Implement MCP configuration manager
- [ ] Add credential storage/retrieval
- [ ] Create MCP template library
- [ ] Test MCP lazy loading

### Phase 4: Infrastructure (Week 4)
- [ ] Auto-install beads-mcp and rtk
- [ ] Add `medtasker-skills` CLI
- [ ] Implement update mechanism
- [ ] Add verification (`doctor`) command

### Phase 5: Polish (Week 5)
- [ ] Write comprehensive documentation
- [ ] Add E2E tests
- [ ] Security audit
- [ ] Release v1.0.0

---

## 11. Appendix

### A.1 Skill Manifest Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["name", "description"],
  "properties": {
    "name": {
      "type": "string",
      "pattern": "^[a-z0-9-]+$"
    },
    "description": {
      "type": "string",
      "maxLength": 200
    },
    "version": {
      "type": "string"
    },
    "platforms": {
      "type": "array",
      "items": {
        "enum": ["claude-code", "opencode", "skillfish"]
      }
    },
    "mcp_servers": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name", "type"],
        "properties": {
          "name": {"type": "string"},
          "type": {"enum": ["stdio", "http"]},
          "command": {"type": "string"},
          "args": {"type": "array", "items": {"type": "string"}},
          "env": {"type": "object"},
          "url": {"type": "string"}
        }
      }
    },
    "dependencies": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name"],
        "properties": {
          "name": {"type": "string"},
          "version": {"type": "string"},
          "required": {"type": "boolean", "default": true}
        }
      }
    },
    "permissions": {
      "type": "object",
      "properties": {
        "mcp": {"enum": ["allow", "deny"]},
        "write": {"enum": ["allow", "deny"]},
        "edit": {"enum": ["allow", "deny"]}
      }
    }
  }
}
```

### A.2 Directory Permissions

| Directory | Permissions | Owner |
|-----------|-------------|-------|
| `~/.medtasker-skills/` | 700 (rwx------) | User |
| `~/.medtasker-skills/.env` | 600 (rw-------) | User |
| `~/.claude/skills/` | 755 (rwxr-xr-x) | User |
| `~/.opencode/skills/` | 755 (rwxr-xr-x) | User |

### A.3 Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `JIRA_USERNAME` | Jira API username | For Jira skills |
| `JIRA_API_TOKEN` | Jira API token | For Jira skills |
| `CONFLUENCE_URL` | Confluence base URL | For Confluence skills |
| `GITHUB_TOKEN` | GitHub personal access token | For GitHub skills |
| `FIGMA_API_TOKEN` | Figma API token | For Figma skills |
| `MEDTASKER_SKILLS_DEBUG` | Enable debug logging | Optional |

---

## 12. References

1. [Claude Code Plugin Documentation](https://code.claude.com/docs/en/plugins)
2. [OhMyOpenCode MCP Configuration](https://github.com/code-yeongyu/oh-my-opencode)
3. [skill.fish NPM Package](https://www.npmjs.com/package/skillfish)
4. [DockYard/skill CLI](https://github.com/DockYard/skill)
5. [Model Context Protocol Specification](https://modelcontextprotocol.io/)
6. [Medtasker Jira Skill](~/.agents/skills/medtasker-jira/SKILL.md)
7. [Medtasker Move Skill](~/.agents/skills/medtasker-jira-ticket-transition/SKILL.md)

---

*Document Version: 1.0.0*
*Last Updated: 2026-05-21*
*Authors: Medtasker Engineering Team*
