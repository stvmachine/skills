# Development

## Prerequisites

- [Go](https://go.dev/dl/)
- [Node.js](https://nodejs.org/) (for `dotenvx` and Claude Code)
- [dotenvx](https://github.com/dotenvx/dotenvx): `npm install -g @dotenvx/dotenvx`
- [Claude Code](https://docs.claude.com/en/docs/claude-code): `npm install -g @anthropic-ai/claude-code`

## Build and test

```bash
go build ./cmd/stevmachine-skills
go test ./...
```

## Generating README assets

The demo GIF and setup screenshot are generated from `tapes/` using `tapes/generate.sh`.

Install the required tools via Homebrew:

```bash
brew install charmbracelet/tap/freeze
brew install vhs
brew install ffmpeg
brew install ttyd
brew install --cask chromium
```

Then run:

```bash
./tapes/generate.sh
```

Assets are written to `docs/assets/`.
