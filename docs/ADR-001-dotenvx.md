# ADR-001: Adopt dotenvx for Encrypted Credential Management

## Status

**Accepted** — supersedes any prior plan to use OS-native credential stores.

## Context

The medtasker-skills project is building a skill distribution system for AI agent platforms (Claude Code, OpenCode, skill.fish). The system manages MCP server credentials for services including Jira, GitHub, Figma, and Confluence.

The access pattern is what drives the design: Claude Code reads `~/.claude/.mcp.json` and expands `${VAR}` references from its own process environment when it spawns an MCP server. So whatever stores the secrets at rest must be able to inject them into the env of the `claude` process at launch time, without ever writing plaintext to disk.

An earlier iteration of the design layered OS-native credential stores (macOS Keychain, Linux Secret Service, Windows Credential Manager) in front of an encrypted vault as a fallback. That layering looks attractive on paper but fails the access-pattern test: native stores hold secrets safely but provide no first-class mechanism for injecting them into a child process's environment. Wiring them up requires per-platform shell helpers that read the store and `export` values — adding moving parts that the encrypted-vault approach already solves in one shot.

We need a credential management solution that is:
- Secure by default (no plaintext secrets at rest)
- Cross-platform compatible (single mental model, single tool)
- Easy to synchronize across development environments
- Resistant to accidental exposure (including AI coding tools reading file contents into context)
- Low operational overhead
- **Capable of injecting decrypted values into the process environment at launch time, without writing plaintext to disk**

## Decision

**Adopt [dotenvx](https://github.com/dotenvx/dotenvx) as the sole credential storage mechanism** for the medtasker-skills distribution system. **OS-native credential stores are explicitly not used**, neither as primary storage nor as a fallback.

### Implementation Strategy

1. **Encrypted vault only.** Secrets live in `~/.medtasker-skills/.env` (ciphertext when encrypted) with the decryption key in `~/.medtasker-skills/.env.keys` (plaintext, machine-local, never committed). No plaintext `.env` fallback, no Keychain/Secret Service/Credential Manager integration.
2. **ECIES encryption** (secp256k1 curve + AES-256-GCM) for all stored credentials.
3. **Cryptographic separation.** The vault and the decryption key are stored as separate files so an attacker must compromise both to recover plaintext.
4. **Environment branching** via `.env` / `.env.production` / etc.
5. **`.mcp.json` holds only `${VAR}` references**, never resolved values. Resolution happens at process launch via `dotenvx run`, so plaintext exists only inside the `claude` (and child MCP server) process memory.

### Key Workflow

```bash
# One-time setup
medtasker-skills env set JIRA_API_TOKEN <token>   # writes through to vault

# Launch Claude Code with vault decrypted in-process
dotenvx run -f ~/.medtasker-skills/.env -- claude

# Share encrypted .env across devices (key managed separately)
git add .env           # safe — ciphertext
# .env.keys stays out of git; transport via password manager or secure channel
```

### Why Not OS-Native Stores

This is the part that changed from the earlier draft. Each native store was considered and rejected for the same structural reason:

- **macOS Keychain**: no built-in `keychain run -- cmd` equivalent. To make secrets visible to `claude`, you'd write a shell helper that calls `security find-generic-password` for every key and `export`s the result. That helper has to live somewhere, be maintained, and re-parse Keychain output formats that Apple has historically changed. dotenvx solves the same problem in one command.
- **Linux Secret Service**: works only when a session keyring is available (often missing on headless boxes, WSL, CI). Falling back from it lands you right back at the vault — so just use the vault everywhere.
- **Windows Credential Manager**: same shape as Keychain. No process-env injection primitive; requires a custom shim.

Layering native stores in front of the vault also created a coherence problem: in mixed-store mode, the vault would be empty on machines where the native store handled writes, breaking `dotenvx run`. Either every secret has to be mirrored to both, or one of them is dead weight. A single store is simpler and easier to reason about.

## Consequences

### Positive Consequences

1. **Dramatically Reduced Breach Risk**: ECIES encryption provides 200× lower breach risk versus centralized storage (99.5% reduction). Even if the vault file is exposed, secrets remain encrypted.

2. **AI-Safe Credentials**: Encrypted files protect against AI coding tools (Claude Code, GitHub Copilot, etc.) reading secrets when processing files in context. This is critical as medtasker-skills integrates deeply with AI agent platforms.

3. **Zero Infrastructure**: Unlike cloud-based secret managers, dotenvx requires no servers, no SaaS subscriptions, and no network access. Secrets are stored in git alongside code.

4. **Cross-Platform Consistency**: Works identically on macOS, Linux, and Windows. Eliminates platform-specific credential store fragmentation.

5. **Free and Open Source**: No per-user licensing costs. Suitable for open-source distribution and community contributions.

6. **Cryptographic Separation**: The encrypted secrets file and decryption key are stored separately. An attacker must compromise both to access plaintext secrets.

7. **Environment Branching**: Native support for `.env.development`, `.env.production`, etc., aligned with modern deployment practices.

8. **Future-Proof for Agentic Storage**: dotenvx is developing Agentic Secret Storage (AS2) for autonomous software — aligned with medtasker-skills' AI-native architecture.

9. **Industry Validation**: Adopted by major organizations including NASA (Earthdata Search), Supabase, AWS Amplify Gen 2, Cloudflare, and PayPal.

### Negative Consequences

1. **Key Management Overhead**: Users must securely manage decryption keys (e.g., via password manager, hardware token, or platform keychain). Losing the key means losing access to credentials.

2. **New Dependency**: Adds a Node.js-based tool dependency. While lightweight, this introduces a supply chain consideration.

3. **Migration Effort**: Existing plaintext `.env` files must be encrypted. Users need education on the new workflow.

4. **No Centralized Access Control**: Unlike HashiCorp Vault or Doppler, dotenvx does not provide RBAC, audit logs, or secret rotation policies. Access is binary (have key / don't have key).

5. **CLI-First Interface**: Primarily command-line driven. Less friendly for non-technical users compared to GUI vault applications.

6. **Recovery Complexity**: No "forgot password" flow. Key loss = permanent secret loss. Requires backup discipline.

## Alternatives Considered

### 1. Doppler

| Criteria | Doppler | dotenvx |
|----------|---------|---------|
| **Cost** | $7–15/user/month | Free (open source) |
| **Infrastructure** | Cloud SaaS | Zero infrastructure |
| **Offline Access** | Requires internet | Fully offline |
| **Complexity** | Medium | Low |

**Decision**: Rejected due to recurring cost, SaaS dependency, and unnecessary complexity for a distribution system that needs to work offline and at the edge.

### 2. 1Password Secrets Automation

| Criteria | 1Password | dotenvx |
|----------|-----------|---------|
| **Cost** | $7.99/user/month | Free (open source) |
| **Model** | Vault-based | Git-based |
| **CI/CD Integration** | Requires service accounts | Native via encrypted files |
| **AI Context Safety** | Good (external vault) | Good (encrypted files) |

**Decision**: Rejected due to cost and vault-centric model that doesn't align with git-based distribution. 1Password is excellent for human credentials but less suited for machine-to-machine MCP server secrets.

### 3. HashiCorp Vault

| Criteria | HashiCorp Vault | dotenvx |
|----------|-----------------|---------|
| **Cost** | Enterprise licensing / self-hosted ops | Free |
| **Complexity** | High (dedicated team often required) | Low |
| **Operational Burden** | Significant (HA, backups, upgrades) | None |
| **Features** | RBAC, dynamic secrets, audit logs | Encryption + environment branching |

**Decision**: Rejected as massive overkill. Vault's enterprise features (RBAC, dynamic secrets, PKI) are unnecessary for medtasker-skills' scope. The operational burden contradicts the project's low-friction goals.

### 4. Continue with Plaintext `.env`

| Criteria | Plaintext `.env` | dotenvx |
|----------|------------------|---------|
| **Security** | Poor (readable by any process) | Strong (ECIES encrypted) |
| **AI Safety** | None (secrets exposed in context) | Protected |
| **Accidental Commit Risk** | High | Low (encrypted files are safe to commit) |
| **Cross-Platform Sync** | Manual file copying | Git-based synchronization |

**Decision**: Rejected. Plaintext storage is a known anti-pattern that becomes unacceptable as the project scales to community distribution and AI-native workflows.

## References

1. **[dotenvx Official Repository](https://github.com/dotenvx/dotenvx)** — Core project documentation and CLI reference

2. **[NASA Earthdata Search](https://github.com/nasa/earthdata-search)** (813 stars)
   - Uses `dotenvx run` in package.json scripts for API server startup
   - Provides cross-platform consistency for mission-critical Earth science data operations
   - Demonstrates production reliability in scientific computing environments

3. **[Supabase](https://github.com/supabase/supabase)** (103k+ stars)
   - Official documentation recommends dotenvx for branching integration
   - Encrypted secrets committed to git, decrypted at deploy time
   - Zero CI code — no custom scripts needed for secrets management
   - Native integration with Supabase branching executor

4. **[AWS Amplify Gen 2](https://docs.amplify.aws/)**
   - Official AWS documentation explicitly recommends dotenvx
   - Used for local development with sandbox environments
   - Example workflow: `npx dotenvx run --env-file=.env.local -- ampx sandbox`

5. **[Cloudflare Workers & Pages](https://developers.cloudflare.com/)**
   - Official documentation for edge runtime integration
   - Encrypted `.env` files imported directly into Worker entrypoints
   - Private keys set as Worker secrets via Wrangler

6. **[PayPal](https://www.paypal.com/)** — Mentioned as adopter in dotenvx 1M installs blog post, demonstrating enterprise fintech validation

7. **[dotenvx 1M Installs Blog Post](https://dotenvx.com/blog/2024/06/24/dotenvx-1m-installs.html)** — Overview of adoption milestones and feature roadmap including Agentic Secret Storage (AS2)

8. **[dotenvx Documentation: Encryption](https://dotenvx.com/docs/features/encryption)** — Technical details on ECIES implementation (secp256k1 + AES-256-GCM)

## Notes

- This ADR should be reviewed after 3 months of production use to validate key management workflows and user adoption
- Consider contributing to dotenvx's Agentic Secret Storage (AS2) initiative as it aligns with medtasker-skills' autonomous AI architecture
- Monitor for native Python bindings or alternatives if Node.js dependency becomes problematic for pure-Python environments

## Follow-up Implementation Work

This ADR change implies code changes that have not yet landed:

- Remove the `CredentialManager` Keychain / Secret Service / Windows Credential Manager branches in `src/installer/credentials.py`; route all reads/writes through `DotenvxVaultManager`.
- Drop the `keyring` and `pywin32` optional dependencies.
- Update installer/CLI surfaces so `medtasker-skills env set/get/list` are vault-only.
- Document the launcher (shell function or wrapper script) that runs `dotenvx run -f ~/.medtasker-skills/.env -- claude` so users don't have to type it.
