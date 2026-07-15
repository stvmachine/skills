#!/usr/bin/env bash
# Regenerate all README assets (docs/assets/).
# Requires: freeze (charmbracelet/tap/freeze), vhs, ffmpeg, ttyd, chromium
set -euo pipefail

REPO="$(cd "$(dirname "$0")/.." && pwd)"
ASSETS="$REPO/docs/assets"
TAPES="$REPO/tapes"

check() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing: $1  →  $2"; exit 1; }
}
check freeze  "brew install charmbracelet/tap/freeze"
check vhs     "brew install vhs"
check ffmpeg  "brew install ffmpeg"
check ttyd    "brew install ttyd"

mkdir -p "$ASSETS"

FREEZE_FLAGS=(
  --theme "dracula"
  --background "#16213e"
  --window
  --border.radius 10
  --shadow.blur 24
  --shadow.x 0
  --shadow.y 10
  --margin 24
  --padding 20
)

echo "→ freeze: setup-flow.png"
freeze "$TAPES/setup-flow.sh" \
  "${FREEZE_FLAGS[@]}" \
  --language bash \
  --font.size 14 \
  --width 860 \
  --output "$ASSETS/setup-flow.png"

echo "→ vhs: demo.gif"
# Build into a temp dir and put it on PATH for the recording session.
# Don't overwrite the user's /opt/homebrew/bin/stevmachine-skills.
# VHS Output paths must be relative, so we cd into tapes/ and move the result.
TMPBIN="$(mktemp -d)"
trap 'rm -rf "$TMPBIN"' EXIT
go build -o "$TMPBIN/stevmachine-skills" "$REPO/cmd/stevmachine-skills"
cd "$TAPES"
PATH="$TMPBIN:$PATH" vhs demo.tape
mv demo.gif "$ASSETS/demo.gif"

echo "✓ All assets written to docs/assets/"
