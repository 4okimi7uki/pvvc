package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/pelletier/go-toml/v2"

	"github.com/4okimi7uki/pvvc/internal/config"
	"github.com/4okimi7uki/pvvc/internal/ui"
)

func loadExistingConfig(filePath string) (config.Config, error) {
	var cfg config.Config
	//nolint:gosec // G304: Potential file inclusion via variable
	data, err := os.ReadFile(filePath)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file: %w", err)
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse config file: %w", err)
	}
	return cfg, nil
}

func RunSetup() error {
	var configs config.Config
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home dir: %w", err)
	}

	var (
		dir      = filepath.Join(home, ".config", "pvvc")
		filePath = filepath.Join(dir, "config.toml")
	)

	// check exists
	if _, err := os.Stat(filePath); err == nil {
		existing, err := loadExistingConfig(filePath)
		if err != nil {
			return err
		}
		configs = existing
	}

	{
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Vercel").
					Description("Billing API credentials.\nFind your token at: vercel.com/account/tokens"),
				huh.NewInput().
					Title("Token").
					EchoMode(huh.EchoModePassword).
					Value(&configs.Vercel.Token),
				huh.NewInput().
					Title("Team ID").
					Description("Settings → General → Team ID  (leave blank for personal accounts)").
					EchoMode(huh.EchoModeNormal).
					Value(&configs.Vercel.TeamID),
				huh.NewInput().
					Title("Project IDs").
					Description("Settings → General → Project ID\nComma-separated for multiple projects: prj_aaa,prj_bbb").
					EchoMode(huh.EchoModeNormal).
					Value(&configs.Vercel.ProjectIDs),
				huh.NewInput().
					Title("Project URL").
					Description("Optional. Adds Usage & Logs links to Slack notifications.\ne.g. https://vercel.com/your-team/your-project").
					EchoMode(huh.EchoModeNormal).
					Value(&configs.Vercel.ProjectURL),
			),
			huh.NewGroup(
				huh.NewNote().
					Title("Google Analytics 4").
					Description("Pageview data source.\nSet up at: console.cloud.google.com"),
				huh.NewInput().
					Title("Property ID").
					Description("Admin → Property Settings → Property ID (numeric)").
					Validate(isInt).
					EchoMode(huh.EchoModeNormal).
					Value(&configs.Ga4.PropertyID),
				huh.NewInput().
					Title("Credential").
					Description("Service account JSON compressed to one line:\ncat key.json | tr -d '\\n'").
					EchoMode(huh.EchoModeNormal).
					Value(&configs.Ga4.Credential),
			),
			huh.NewGroup(
				huh.NewNote().
					Title("Additional Settings").
					Description("All optional — leave blank to skip each feature."),
				huh.NewInput().
					Title("Gemini API Key").
					Description("AI trend analysis (default). Skip to disable.  → aistudio.google.com/app/apikey").
					EchoMode(huh.EchoModePassword).
					Value(&configs.Ai.GeminiKey),
				huh.NewInput().
					Title("Claude API Key").
					Description("Alternative AI provider (--llm claude). Skip to disable.  → console.anthropic.com").
					EchoMode(huh.EchoModePassword).
					Value(&configs.Ai.ClaudeKey),
				huh.NewInput().
					Title("Slack Webhook URL").
					Description("Required for --notify flag.  → api.slack.com/messaging/webhooks").
					EchoMode(huh.EchoModePassword).
					Value(&configs.Slack.WebhookURL),
				huh.NewInput().
					Title("Service Name").
					Description("Display name shown in reports and Slack messages.").
					EchoMode(huh.EchoModeNormal).
					Value(&configs.Service.Name),
				huh.NewInput().
					Title("Service URL").
					Description("Base URL used to build page links in Slack notifications.\ne.g. https://www.example.com").
					EchoMode(huh.EchoModeNormal).
					Value(&configs.Service.URL),
			),
		).WithTheme(pvvcTheme())
		if err := form.Run(); err != nil {
			return err
		}

		// create config file
		path, err := createConfigFile(dir, filePath, configs)
		if err != nil {
			return err
		}

		fmt.Printf("\n %s Config saved!\n", ui.Green("✔"))
		fmt.Printf("   %s\n\n", ui.LimeYellow(path))
		fmt.Printf("   %s\n\n", ui.MossGray("Run `pvvc report` to get started."))
	}
	return nil
}

func createConfigFile(dir, filePath string, cfg config.Config) (string, error) {
	// templates/ holds prompt template files fetched from a public GitHub repo.
	// No secrets, no PII — safe with default permissions.
	//nolint:gosec // G301: public template cache, not sensitive
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create dir: %w", err)
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to Marshal: %w", err)
	}

	//nolint:gosec // G306: public template content, not sensitive
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", filePath, err)
	}

	return filePath, nil
}

func isInt(s string) error {
	_, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return err
	}
	return nil
}
