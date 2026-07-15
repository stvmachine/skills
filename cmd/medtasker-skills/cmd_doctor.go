package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stevmachine/skills/internal/vault"
)

func cmdDoctor() {
	home, _ := os.UserHomeDir()
	claudeDir := filepath.Join(home, ".claude")
	mcpConfig := filepath.Join(claudeDir, ".mcp.json")

	_, dotenvxErr := exec.LookPath("dotenvx")
	_, bdErr := exec.LookPath("bd")
	_, rtkErr := exec.LookPath("rtk")
	v := vault.New()

	const labelWidth = 22

	fmt.Println()
	fmt.Println(titleStyle.Render("  Doctor"))
	fmt.Println()
	fmt.Println(statusLine("Claude Code dir", boolStatus(claudeDir != "" && fileExists(claudeDir)), claudeDir, labelWidth))
	fmt.Println(statusLine("dotenvx", boolStatus(dotenvxErr == nil), "", labelWidth))
	fmt.Println(statusLine("Vault", vaultStatus(v), v.EnvDir, labelWidth))
	fmt.Println(statusLine("MCP config", boolStatus(fileExists(mcpConfig)), mcpConfig, labelWidth))
	fmt.Println(statusLine("beads (bd)", optionalBoolStatus(bdErr == nil), "", labelWidth))
	fmt.Println(statusLine("rtk", optionalBoolStatus(rtkErr == nil), "", labelWidth))

	ticketDir := os.Getenv("MEDTASKER_TICKET_DIR")
	if ticketDir == "" {
		ticketDir = "./.todo (default)"
	}
	fmt.Println(statusLine("MEDTASKER_TICKET_DIR", "info", ticketDir, labelWidth))

	fmt.Println()
	printSuggestions(bdErr != nil, rtkErr != nil)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func boolStatus(ok bool) string {
	if ok {
		return "OK"
	}
	return "MISSING"
}

// optionalBoolStatus is for tools that enhance the workflow but aren't required.
func optionalBoolStatus(ok bool) string {
	if ok {
		return "OK"
	}
	return "not installed (optional)"
}

func vaultStatus(v *vault.Manager) string {
	if v.IsInitialized("") {
		return "initialized"
	}
	return "empty"
}

// printSuggestions shows install hints for optional tools. Called from both
// cmdDoctor and cmdInstall when any optional dep is missing.
func printSuggestions(bdMissing, rtkMissing bool) {
	if !bdMissing && !rtkMissing {
		return
	}
	fmt.Println(titleStyle.Render("  Suggestions"))
	fmt.Println()
	if bdMissing {
		fmt.Println("  " + styleFaint.Render("beads (bd) not installed — Jira tickets will use filesystem (./.todo/) instead of beads."))
		fmt.Println("    " + styleGreen.Render("→") + " Install: " + styleYellow.Render("curl -fsSL https://raw.githubusercontent.com/gastownhall/beads/main/integrations/beads-mcp/install.sh | bash"))
		fmt.Println()
	}
	if rtkMissing {
		fmt.Println("  " + styleFaint.Render("rtk not installed — Claude Code commands won't be token-optimized."))
		fmt.Println("    " + styleGreen.Render("→") + " Install: " + styleYellow.Render("brew install rtk"))
		fmt.Println()
	}
}
