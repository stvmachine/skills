# Troubleshooting

## "Command not found: medtasker-skills"

The CLI is not in your PATH.

```bash
# Find where it was installed
which medtasker-skills || find ~/.local -name "medtasker-skills" 2>/dev/null

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
medtasker-skills env set KEY VALUE
```

This creates the vault automatically on first use.

## "MCP credentials not resolved"

Claude Code must be launched via `dotenvx run` so the vault variables are in its env:

```bash
dotenvx run -f ~/.medtasker-skills/.env -- claude
```

Check that variables are set:

```bash
medtasker-skills env list
```

## Permission denied on vault directory

```bash
chmod 700 ~/.medtasker-skills
chmod 600 ~/.medtasker-skills/.env*
```

## Skills not showing up in Claude Code

Verify they are installed:

```bash
medtasker-skills list
ls ~/.claude/skills/
```

Restart Claude Code if you just installed them.
