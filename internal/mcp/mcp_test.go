package mcp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSkillMcpConfig(t *testing.T) {
	dir := t.TempDir()

	// No SKILL.md
	_, err := ParseSkillMcpConfig(dir)
	if err == nil {
		t.Error("expected error for missing SKILL.md")
	}

	// No frontmatter
	_ = os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# plain\n"), 0o644)
	servers, err := ParseSkillMcpConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(servers) != 0 {
		t.Error("expected no servers")
	}

	// Valid frontmatter
	content := `---
mcp_servers:
  - name: mcp-atlassian
    command: npx
    env:
      JIRA_API_TOKEN: ${JIRA_API_TOKEN}
---
# Jira Skill
`
	_ = os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644)
	servers, err = ParseSkillMcpConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	if servers[0].Name != "mcp-atlassian" {
		t.Errorf("unexpected name: %s", servers[0].Name)
	}
	if servers[0].Env["JIRA_API_TOKEN"] != "${JIRA_API_TOKEN}" {
		t.Errorf("env var not literal: %s", servers[0].Env["JIRA_API_TOKEN"])
	}
}

func TestParseIgnoresEntriesWithoutName(t *testing.T) {
	dir := t.TempDir()
	content := `---
mcp_servers:
  - command: npx
---
# Skill
`
	_ = os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644)
	servers, _ := ParseSkillMcpConfig(dir)
	if len(servers) != 0 {
		t.Error("expected empty servers for missing names")
	}
}

func TestBuildMcpServers(t *testing.T) {
	servers := []ServerConfig{
		{
			Name:    "srv1",
			Command: "npx",
			Args:    []string{"-y", "pkg"},
			Env:     map[string]string{"KEY": "${VAL}"},
		},
	}
	built := BuildMcpServers(servers)
	if built["srv1"]["command"] != "npx" {
		t.Error("command wrong")
	}
	if built["srv1"]["env"].(map[string]string)["KEY"] != "${VAL}" {
		t.Error("env not literal")
	}

	// Missing args/env should be normalized to empty collections
	servers2 := []ServerConfig{{Name: "srv2", Command: "cmd"}}
	built2 := BuildMcpServers(servers2)
	if built2["srv2"]["args"] == nil {
		t.Error("nil args should become empty slice")
	}
	if built2["srv2"]["env"] == nil {
		t.Error("nil env should become empty map")
	}
}

func TestWriteMcpConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mcp.json")

	newServers := map[string]map[string]any{
		"srv1": {"command": "npx"},
	}
	if err := WriteMcpConfig(path, newServers); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	if !contains(string(data), "srv1") {
		t.Error("expected srv1 in config")
	}

	// Merge into existing
	newServers2 := map[string]map[string]any{
		"srv2": {"command": "node"},
	}
	if err := WriteMcpConfig(path, newServers2); err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	data, _ = os.ReadFile(path)
	if !contains(string(data), "srv1") || !contains(string(data), "srv2") {
		t.Error("expected both servers in merged config")
	}

	// Overwrite same name
	newServers3 := map[string]map[string]any{
		"srv1": {"command": "updated"},
	}
	if err := WriteMcpConfig(path, newServers3); err != nil {
		t.Fatalf("overwrite failed: %v", err)
	}
	data, _ = os.ReadFile(path)
	if !contains(string(data), `"command": "updated"`) {
		t.Error("expected updated command")
	}
}

func contains(s, substr string) bool {
	return len(substr) <= len(s) && (s == substr || len(s) > 0 && containsSub(s, substr))
}

func containsSub(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
