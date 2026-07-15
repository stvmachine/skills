# Troubleshooting

## "Command not found: stevmachine-skills"

The CLI is not in your PATH.

```bash
# Find where it was installed
which stevmachine-skills || find ~/.local -name "stevmachine-skills" 2>/dev/null

# Add to PATH (bash/zsh)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

## "dotenvx not found"

```bash
npm install -g @dotenvx/dotenvx
# or
curl -sfS https://dotenvx.sh/install.sh | sh
```

## "Vault not initialized"

```bash
stevmachine-skills env set KEY VALUE
```

This creates the vault automatically on first use.

## "MCP credentials not resolved"

Claude Code must be launched via `dotenvx run` so the vault variables are in its env:

```bash
dotenvx run -f ~/.stevmachine-skills/.env -- claude
```

Check that variables are set:

```bash
stevmachine-skills env list
```

## Permission denied on vault directory

```bash
chmod 700 ~/.stevmachine-skills
chmod 600 ~/.stevmachine-skills/.env*
```

## Skills not showing up in Claude Code

Verify they are installed:

```bash
stevmachine-skills list
ls ~/.claude/skills/
```

Restart Claude Code if you just installed them.
