package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/4okimi7uki/pvvc/internal/config"
	"github.com/4okimi7uki/pvvc/internal/ui"
	"github.com/charmbracelet/huh"
	"github.com/pelletier/go-toml/v2"
)

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
	override := true
	if _, err := os.Stat(filePath); err == nil {
		err = huh.NewConfirm().Title("Config file already exists. Override?").Affirmative("Yes").Negative("No").Value(&override).WithTheme(pvvcTheme()).Run()
		if err != nil {
			return fmt.Errorf("failed to check config file: %w", err)
		}
	}

	if override {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("Vercel").
					Description("Enter your Vercel credentials."),
				huh.NewInput().Title("Token").EchoMode(huh.EchoModePassword).Value(&configs.Vercel.Token),
				huh.NewInput().Title("Team ID").EchoMode(huh.EchoModeNormal).Value(&configs.Vercel.TeamId),
				huh.NewInput().Title("Project ID").EchoMode(huh.EchoModeNormal).Value(&configs.Vercel.ProjectId),
			),
			huh.NewGroup(
				huh.NewNote().
					Title("Google Analytics 4").
					Description("Enter your GA4 credentials."),
				huh.NewInput().Title("Property ID").Validate(isInt).EchoMode(huh.EchoModeNormal).Value(&configs.Ga4.PropertyId),
				huh.NewInput().Title("Google Analytics Credential").EchoMode(huh.EchoModeNormal).Value(&configs.Ga4.Credential),
			),
			huh.NewGroup(
				huh.NewNote().
					Title("Additional Settings").
					Description("Enter your AI and notification settings."),
				huh.NewInput().Title("Gemini API Key").EchoMode(huh.EchoModePassword).Value(&configs.Ai.GeminiKey),
				huh.NewInput().Title("Slack Webhook").EchoMode(huh.EchoModePassword).Value(&configs.Slack.WebhookUrl),
				huh.NewInput().Title("Service Name").EchoMode(huh.EchoModeNormal).Value(&configs.Service.Name),
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

		fmt.Printf("%s Successfully saved config:\n%s\n", ui.Green("✔"), path)
	}
	return nil
}

func createConfigFile(dir, filePath string, cfg config.Config) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create dir: %w", err)
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to Marshal: %w", err)
	}

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
