package mcp

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Name    string            `yaml:"name"`
	Type    string            `yaml:"type,omitempty"`
	Command string            `yaml:"command,omitempty"`
	URL     string            `yaml:"url,omitempty"`
	Args    []string          `yaml:"args,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

type SkillFrontmatter struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	McpServers  []ServerConfig `yaml:"mcp_servers"`
}

func parseFrontmatterBytes(data []byte) (*SkillFrontmatter, error) {
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return &SkillFrontmatter{}, nil
	}
	rest := content[3:]
	end := strings.Index(rest, "---")
	if end == -1 {
		return &SkillFrontmatter{}, nil
	}
	var fm SkillFrontmatter
	if err := yaml.Unmarshal([]byte(rest[:end]), &fm); err != nil {
		return nil, err
	}
	return &fm, nil
}

func ParseSkillFrontmatter(skillPath string) (*SkillFrontmatter, error) {
	data, err := os.ReadFile(filepath.Join(skillPath, "SKILL.md"))
	if err != nil {
		return nil, err
	}
	return parseFrontmatterBytes(data)
}

func ParseSkillFrontmatterFS(fsys fs.FS, skillPath string) (*SkillFrontmatter, error) {
	data, err := fs.ReadFile(fsys, skillPath+"/SKILL.md")
	if err != nil {
		return nil, err
	}
	return parseFrontmatterBytes(data)
}

func ParseSkillMcpConfig(skillPath string) ([]ServerConfig, error) {
	fm, err := ParseSkillFrontmatter(skillPath)
	if err != nil {
		return nil, err
	}
	var result []ServerConfig
	for _, s := range fm.McpServers {
		if s.Name != "" {
			result = append(result, s)
		}
	}
	return result, nil
}

// BaselineServers returns MCP servers that should always be present in
// ~/.claude/.mcp.json regardless of which skills are installed. Intentionally
// empty — no MCP currently earns a place in every user's global config by
// default. Kept as a named extension point.
func BaselineServers() []ServerConfig {
	return []ServerConfig{}
}

func BuildMcpServers(servers []ServerConfig) map[string]map[string]any {
	result := make(map[string]map[string]any)
	for _, s := range servers {
		entry := map[string]any{}
		if s.Type != "" {
			entry["type"] = s.Type
		}
		if s.Type == "http" {
			entry["url"] = s.URL
			if s.Headers != nil {
				entry["headers"] = s.Headers
			} else {
				entry["headers"] = map[string]string{}
			}
		} else {
			entry["command"] = s.Command
			if s.Args != nil {
				entry["args"] = s.Args
			} else {
				entry["args"] = []string{}
			}
			if s.Env != nil {
				entry["env"] = s.Env
			} else {
				entry["env"] = map[string]string{}
			}
		}
		result[s.Name] = entry
	}
	return result
}

func WriteMcpConfig(path string, newServers map[string]map[string]any) error {
	var existing map[string]any
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &existing)
	}
	if existing == nil {
		existing = map[string]any{"mcpServers": map[string]any{}}
	}
	mcpServers, _ := existing["mcpServers"].(map[string]any)
	if mcpServers == nil {
		mcpServers = map[string]any{}
		existing["mcpServers"] = mcpServers
	}
	for k, v := range newServers {
		mcpServers[k] = v
	}
	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
