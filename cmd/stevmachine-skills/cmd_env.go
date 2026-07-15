package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/stevmachine/skills/internal/vault"
)


func cmdEnvSet(args []string) {
	fs := flag.NewFlagSet("set", flag.ExitOnError)
	env := fs.String("e", "", "environment")
	_ = fs.Parse(args)
	remaining := fs.Args()
	if len(remaining) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: stevmachine-skills env set [-e env] KEY VALUE")
		os.Exit(1)
	}
	key, value := remaining[0], remaining[1]
	v := vault.New()
	if err := v.Set(key, value, *env); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(confirmLine("Set " + styleYellow.Render(key)))
	fmt.Println()
}

func cmdEnvList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	env := fs.String("e", "", "environment")
	_ = fs.Parse(args)
	v := vault.New()
	masked := v.ListMasked(*env)
	fmt.Println()
	fmt.Println(titleStyle.Render("  Vault Variables"))
	fmt.Println()
	if masked == nil {
		fmt.Println("  " + styleFaint.Render("No vault initialized."))
		fmt.Println()
		return
	}
	if len(masked) == 0 {
		fmt.Println("  " + styleFaint.Render("No variables set."))
		fmt.Println()
		return
	}
	// Find longest key for alignment
	keyWidth := 0
	for k := range masked {
		if len(k) > keyWidth {
			keyWidth = len(k)
		}
	}
	// Sort keys for deterministic output
	keys := make([]string, 0, len(masked))
	for k := range masked {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		pad := strings.Repeat(" ", keyWidth-len(k))
		fmt.Printf("  %s  %s%s  %s\n",
			styleGreen.Render("●"),
			k,
			pad,
			styleFaint.Render(masked[k]),
		)
	}
	fmt.Println()
	fmt.Println("  " + styleFaint.Render(fmt.Sprintf("%d variable(s) · %s", len(masked), v.EnvDir)))
	fmt.Println()
}

func cmdEnvEncrypt(args []string) {
	fs := flag.NewFlagSet("encrypt", flag.ExitOnError)
	env := fs.String("e", "", "environment")
	_ = fs.Parse(args)
	v := vault.New()
	if err := v.Encrypt(*env); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(confirmLine("Encrypted " + styleFaint.Render(v.EnvDir+"/.env")))
	fmt.Println()
}

func cmdEnvDecrypt(args []string) {
	fs := flag.NewFlagSet("decrypt", flag.ExitOnError)
	env := fs.String("e", "", "environment")
	_ = fs.Parse(args)
	v := vault.New()
	if err := v.Decrypt(*env); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(confirmLine("Decrypted " + styleFaint.Render(v.EnvDir+"/.env")))
	fmt.Println("  " + styleYellow.Render("⚠  Re-encrypt before leaving the shell — plaintext is sitting on disk."))
	fmt.Println()
}

func cmdEnvRotate(args []string) {
	fs := flag.NewFlagSet("rotate", flag.ExitOnError)
	env := fs.String("e", "", "environment")
	_ = fs.Parse(args)
	if err := vault.New().Rotate(*env); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println()
	fmt.Println(confirmLine("Rotated encryption keys"))
	fmt.Println()
}

func cmdEnvSetup() {
	var (
		jira       bool
		jiraURL    string
		jiraUser   string
		jiraToken  string
		github     bool
		githubToken string
		confluence bool
		confURL    string
		confUser   string
		confToken  string
		figma        bool
		figmaToken   string
		context7     bool
		context7Key  string
	)

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(titleStyle.Render("Stevmachine Skills Setup")).
				Description("Select integrations and enter credentials.\nThey are encrypted with dotenvx and never stored plaintext."),
		),
		huh.NewGroup(
			huh.NewConfirm().Title("Configure Jira?").Description("mcp-atlassian server").Value(&jira),
		),
		huh.NewGroup(
			huh.NewInput().Title("JIRA_URL").Placeholder("https://yourcompany.atlassian.net").Value(&jiraURL),
			huh.NewInput().Title("JIRA_USERNAME").Placeholder("you@example.com").Value(&jiraUser),
			huh.NewInput().Title("JIRA_API_TOKEN").Description("https://id.atlassian.com/manage-profile/security/api-tokens").EchoMode(huh.EchoModePassword).Value(&jiraToken),
		).WithHideFunc(func() bool { return !jira }),
		huh.NewGroup(
			huh.NewConfirm().Title("Configure GitHub?").Description("GitHub MCP server").Value(&github),
		),
		huh.NewGroup(
			huh.NewInput().Title("GITHUB_TOKEN").Description("https://github.com/settings/tokens").EchoMode(huh.EchoModePassword).Value(&githubToken),
		).WithHideFunc(func() bool { return !github }),
		huh.NewGroup(
			huh.NewConfirm().Title("Configure Confluence?").Description("Confluence MCP server").Value(&confluence),
		),
		huh.NewGroup(
			huh.NewInput().Title("CONFLUENCE_URL").Placeholder("https://yourcompany.atlassian.net/wiki").Value(&confURL),
			huh.NewInput().Title("CONFLUENCE_USERNAME").Placeholder("you@example.com").Value(&confUser),
			huh.NewInput().Title("CONFLUENCE_API_TOKEN").EchoMode(huh.EchoModePassword).Value(&confToken),
		).WithHideFunc(func() bool { return !confluence }),
		huh.NewGroup(
			huh.NewConfirm().Title("Configure Figma?").Description("Figma MCP server").Value(&figma),
		),
		huh.NewGroup(
			huh.NewInput().Title("FIGMA_API_KEY").Description("https://www.figma.com/developers/api#access-tokens").EchoMode(huh.EchoModePassword).Value(&figmaToken),
		).WithHideFunc(func() bool { return !figma }),
		huh.NewGroup(
			huh.NewConfirm().Title("Configure Context7?").Description("Library docs MCP server").Value(&context7),
		),
		huh.NewGroup(
			huh.NewInput().Title("CONTEXT7_API_KEY").Description("https://context7.com/api-access").EchoMode(huh.EchoModePassword).Value(&context7Key),
		).WithHideFunc(func() bool { return !context7 }),
	).Run()

	if err != nil {
		fmt.Fprintln(os.Stderr, "Cancelled.")
		os.Exit(1)
	}

	v := vault.New()
	stored := 0
	store := func(k, val string) {
		val = strings.TrimSpace(val)
		if val == "" {
			return
		}
		if err := v.Set(k, val, ""); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to store %s: %v\n", k, err)
		} else {
			stored++
		}
	}

	if jira {
		store("JIRA_URL", jiraURL)
		store("JIRA_USERNAME", jiraUser)
		store("JIRA_API_TOKEN", jiraToken)
	}
	if github {
		store("GITHUB_TOKEN", githubToken)
	}
	if confluence {
		store("CONFLUENCE_URL", confURL)
		store("CONFLUENCE_USERNAME", confUser)
		store("CONFLUENCE_API_TOKEN", confToken)
	}
	if figma {
		store("FIGMA_API_KEY", figmaToken)
	}
	if context7 {
		store("CONTEXT7_API_KEY", context7Key)
	}

	fmt.Println()
	fmt.Println(titleStyle.Render(fmt.Sprintf("Stored %d variable(s)", stored)))
	fmt.Println("Launch Claude Code with:")
	fmt.Println("  dotenvx run -f ~/.stevmachine-skills/.env -- claude")
}
