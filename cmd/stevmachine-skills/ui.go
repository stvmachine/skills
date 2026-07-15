package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")) // purple
	styleGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	styleRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4672"))
	styleYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#F5C25E"))
	styleFaint  = lipgloss.NewStyle().Faint(true)
)

// statusIcon returns the colored icon for a status word.
// OK / present → green ✓; MISSING / required → red ✗;
// "not installed (optional)" or warning → yellow ⚠; anything else → faint dot.
func statusIcon(status string) string {
	switch {
	case status == "OK" || status == "initialized":
		return styleGreen.Render("✓")
	case strings.HasPrefix(status, "MISSING"):
		return styleRed.Render("✗")
	case strings.Contains(status, "not installed") || strings.Contains(status, "optional"):
		return styleYellow.Render("⚠")
	default:
		return styleFaint.Render("·")
	}
}

// statusLine renders one row of the doctor table.
//
//   ✓  Claude Code dir:    OK  /Users/esteban/.claude
//
// `label` gets a trailing colon and is left-padded to labelWidth.
// `status` is colored. `extra` is faint. When status == "info", the status
// word itself is hidden (icon + label + extra is enough for informational rows).
func statusLine(label, status, extra string, labelWidth int) string {
	labelWithColon := label + ":"
	pad := strings.Repeat(" ", max(0, labelWidth-len(labelWithColon)))
	out := fmt.Sprintf("  %s  %s%s", statusIcon(status), labelWithColon, pad)
	if status != "info" {
		out += "  " + statusValueStyled(status)
	}
	if extra != "" {
		out += "  " + styleFaint.Render(extra)
	}
	return out
}

// statusValueStyled colors the status word itself (not the icon).
func statusValueStyled(status string) string {
	switch {
	case status == "OK" || status == "initialized":
		return styleGreen.Render(status)
	case strings.HasPrefix(status, "MISSING"):
		return styleRed.Render(status)
	case strings.Contains(status, "not installed") || strings.Contains(status, "optional"):
		return styleYellow.Render(status)
	default:
		return styleFaint.Render(status)
	}
}

// confirmLine renders a "✓ <message>" line for command-write confirmations
// (env set, env encrypt, env decrypt, env rotate). Matches the install
// command's row style.
func confirmLine(message string) string {
	return fmt.Sprintf("  %s %s", styleGreen.Render("✓"), message)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
