package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stvmachine/skills/internal/mcp"
	"github.com/stvmachine/skills/packages"
)

type installResult struct {
	name    string
	ok      bool
	servers []string
	errMsg  string
}

type skillDoneMsg installResult

type installModel struct {
	spinner    spinner.Model
	skills     []string
	current    int
	results    []installResult
	skillsDir  string
	mcpConfigs []string
}

func newInstallModel(skills []string, skillsDir string, mcpConfigs []string) installModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	return installModel{
		spinner:    s,
		skills:     skills,
		skillsDir:  skillsDir,
		mcpConfigs: mcpConfigs,
	}
}

func (m installModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.installNext())
}

func (m installModel) installNext() tea.Cmd {
	skill := m.skills[m.current]
	skillsDir := m.skillsDir
	mcpConfigs := m.mcpConfigs
	return func() tea.Msg {
		ok, servers, errMsg := doInstallOne(skill, skillsDir, mcpConfigs)
		return skillDoneMsg{name: skill, ok: ok, servers: servers, errMsg: errMsg}
	}
}

func (m installModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case skillDoneMsg:
		m.results = append(m.results, installResult(msg))
		m.current++
		if m.current >= len(m.skills) {
			return m, tea.Quit
		}
		return m, m.installNext()
	}
	return m, nil
}

func (m installModel) View() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  Stevmachine Skills") + "\n\n")

	for _, r := range m.results {
		if r.ok {
			b.WriteString(styleGreen.Render("  ✓ ") + r.name + "\n")
		} else {
			b.WriteString(styleRed.Render("  ✗ ") + r.name + styleFaint.Render("  "+r.errMsg) + "\n")
		}
	}

	if m.current < len(m.skills) {
		b.WriteString("  " + m.spinner.View() + " " + m.skills[m.current] + "\n")
	} else {
		failed := 0
		serverSet := map[string]struct{}{}
		for _, r := range m.results {
			if !r.ok {
				failed++
			}
			for _, s := range r.servers {
				serverSet[s] = struct{}{}
			}
		}
		b.WriteString("\n")
		if failed == 0 {
			b.WriteString(styleGreen.Render(fmt.Sprintf("  %d skills installed", len(m.results))) + "\n")
		} else {
			b.WriteString(styleRed.Render(fmt.Sprintf("  %d/%d failed", failed, len(m.results))) + "\n")
		}
		if len(serverSet) > 0 {
			var svrs []string
			for s := range serverSet {
				svrs = append(svrs, s)
			}
			sort.Strings(svrs)
			b.WriteString("\n")
			b.WriteString(titleStyle.Render("  MCPs Configured") + "\n\n")
			for _, s := range svrs {
				b.WriteString(styleGreen.Render("  ✓ ") + s + "\n")
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func cmdInstall(args []string) {
	home, _ := os.UserHomeDir()
	claudeDir := filepath.Join(home, ".claude")
	skillsDir := filepath.Join(claudeDir, "skills")
	mcpConfigs := []string{
		filepath.Join(claudeDir, ".mcp.json"),
		filepath.Join(home, ".claude.json"),
	}

	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "Claude Code not detected. Install with: npm install -g @anthropic-ai/claude-code")
		os.Exit(1)
	}
	_ = os.MkdirAll(skillsDir, 0o755)

	// Baseline MCP servers — intentionally empty today (see internal/mcp.BaselineServers).
	if baseline := mcp.BaselineServers(); len(baseline) > 0 {
		built := mcp.BuildMcpServers(baseline)
		for _, cfg := range mcpConfigs {
			_ = mcp.WriteMcpConfig(cfg, built)
		}
	}

	skills := args
	if len(skills) == 0 {
		skills = defaultSkills()
	}

	p := tea.NewProgram(newInstallModel(skills, skillsDir, mcpConfigs), tea.WithInput(os.Stdin))
	final, err := p.Run()
	if err != nil {
		// non-TTY fallback: run installs and print plain output
		fmt.Println()
		fmt.Println(titleStyle.Render("  Stevmachine Skills"))
		fmt.Println()
		failed := 0
		for _, skill := range skills {
			ok, _, errMsg := doInstallOne(skill, skillsDir, mcpConfigs)
			if ok {
				fmt.Printf("  %s %s\n", styleGreen.Render("✓"), skill)
			} else {
				fmt.Printf("  %s %s  %s\n", styleRed.Render("✗"), skill, styleFaint.Render(errMsg))
				failed++
			}
		}
		fmt.Println()
		suggestOptionalDeps()
		if failed > 0 {
			os.Exit(1)
		}
		return
	}
	if fm, ok := final.(installModel); ok {
		suggestOptionalDeps()
		for _, r := range fm.results {
			if !r.ok {
				os.Exit(1)
			}
		}
	}
}

// suggestOptionalDeps prints install hints if bd or rtk are missing.
// Called from cmdInstall (both TUI and non-TUI paths) after the install summary.
func suggestOptionalDeps() {
	_, bdErr := exec.LookPath("bd")
	_, rtkErr := exec.LookPath("rtk")
	printSuggestions(bdErr != nil, rtkErr != nil)
}

func defaultSkills() []string {
	return []string{"stevmachine-jira", "stevmachine-jira-markup", "stevmachine-jira-ticket-transition", "commit"}
}

func doInstallOne(skillName, skillsDir string, mcpConfigs []string) (ok bool, servers []string, errMsg string) {
	srcPath := filepath.Join("claude-plugin/skills", skillName)
	info, err := fs.Stat(packages.SkillsFS, srcPath)
	if err != nil || !info.IsDir() {
		return false, nil, "not found"
	}
	dst := filepath.Join(skillsDir, skillName)
	_ = os.RemoveAll(dst)
	if err := copyFS(dst, packages.SkillsFS, srcPath); err != nil {
		return false, nil, err.Error()
	}
	svrs, err := mcp.ParseSkillMcpConfig(dst)
	if err == nil && len(svrs) > 0 {
		built := mcp.BuildMcpServers(svrs)
		for _, cfg := range mcpConfigs {
			_ = mcp.WriteMcpConfig(cfg, built)
		}
		for _, s := range svrs {
			servers = append(servers, s.Name)
		}
	}
	return true, servers, ""
}

func copyFS(dst string, fsys fs.FS, src string) error {
	return fs.WalkDir(fsys, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		r, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()
		w, err := os.Create(target)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, r)
		_ = w.Close()
		return err
	})
}
