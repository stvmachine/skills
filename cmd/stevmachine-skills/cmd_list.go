package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/stvmachine/skills/internal/mcp"
	"github.com/stvmachine/skills/packages"
)

func cmdList() {
	home, _ := os.UserHomeDir()
	skillsDir := filepath.Join(home, ".claude", "skills")

	installed := map[string]bool{}
	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				installed[e.Name()] = true
			}
		}
	}

	type entry struct {
		name        string
		description string
		installed   bool
	}

	var skills []entry
	dirEntries, _ := fs.ReadDir(packages.SkillsFS, "claude-plugin/skills")
	for _, d := range dirEntries {
		if !d.IsDir() {
			continue
		}
		name := d.Name()
		desc := ""
		if fm, err := mcp.ParseSkillFrontmatterFS(packages.SkillsFS, "claude-plugin/skills/"+name); err == nil {
			desc = fm.Description
		}
		skills = append(skills, entry{name: name, description: desc, installed: installed[name]})
	}

	maxName := 0
	for _, s := range skills {
		if len(s.name) > maxName {
			maxName = len(s.name)
		}
	}

	fmt.Println()
	fmt.Println(titleStyle.Render("  Skills"))
	fmt.Println()
	for _, s := range skills {
		indicator := styleFaint.Render("–")
		if s.installed {
			indicator = styleGreen.Render("✓")
		}
		pad := strings.Repeat(" ", maxName-len(s.name))
		fmt.Printf("  %s  %s%s  %s\n", indicator, s.name, pad, styleFaint.Render(s.description))
	}
	fmt.Println()
}
