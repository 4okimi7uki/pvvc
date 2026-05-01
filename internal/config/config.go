package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/4okimi7uki/pvvc/internal/ui"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// NOTE: priority
// 1. 環境変数 / `.env`
// 2. ~/.config/pvvc/config.toml

func New() *viper.Viper {
	_ = godotenv.Load()

	v := viper.New()

	if home, err := os.UserHomeDir(); err == nil {
		v.SetConfigFile(filepath.Join(home, ".config", "pvvc", "config.toml"))
		v.SetConfigType("toml")
		_ = v.ReadInConfig()
	}

	v.AutomaticEnv()

	_ = v.BindEnv("vercel.token", "VERCEL_TOKEN")
	_ = v.BindEnv("vercel.team_id", "TEAM_ID")
	_ = v.BindEnv("vercel.project_id", "PROJECT_ID")
	_ = v.BindEnv("vercel.project_ids", "PROJECT_IDS")
	_ = v.BindEnv("ga4.property_id", "PROPERTY_ID")
	_ = v.BindEnv("ga4.credential", "GOOGLE_ANALYTICS_CREDENTIAL")
	_ = v.BindEnv("ai.gemini_key", "GEMINI_API_KEY")
	_ = v.BindEnv("slack.webhook_url", "SLACK_WEBHOOK_URL")
	_ = v.BindEnv("service.name", "TARGET_WEBSITE_NAME")

	return v
}

// Validate checks that required credentials are set.
// Returns an error with a suggestion to run `pvvc init` if any are missing.
func Validate(v *viper.Viper) error {
	required := []struct {
		key  string
		name string
	}{
		{"vercel.token", "Vercel token"},
		{"ga4.property_id", "GA4 property ID"},
		{"ga4.credential", "GA4 credential"},
	}

	var missing []string
	for _, r := range required {
		if v.GetString(r.key) == "" {
			missing = append(missing, r.name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required config: %s\n%s", strings.Join(missing, ", "), ui.LimeYellow("Hint: run `pvvc init` to set up your credentials"))
	}
	return nil
}

// Warnings returns alert messages for env vars that are unset but would cause
// silent wrong output (e.g. costs showing as $0).
func Warnings(v *viper.Viper) []string {
	var warns []string
	if GetProjectIDs(v) == nil {
		warns = append(warns, "PROJECT_IDS is not set — Vercel costs will show as $0 for all days")
	}
	if v.GetString("vercel.team_id") == "" {
		warns = append(warns, "TEAM_ID is not set — fetching personal account billing (not team)")
	}
	if v.GetString("vercel.project_id") != "" && v.GetString("vercel.project_ids") == "" {
		warns = append(warns, "PROJECT_ID is deprecated — migrate to PROJECT_IDS (e.g. PROJECT_IDS=prj_xxxxxxxx)")
	}
	return warns
}

func GetProjectIDs(v *viper.Viper) []string {
	if raw := v.GetString("vercel.project_ids"); raw != "" {
		var ids []string
		for id := range strings.SplitSeq(raw, ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				ids = append(ids, trimmed)
			}
		}
		if len(ids) > 0 {
			return ids
		}
	}
	if id := v.GetString("vercel.project_id"); id != "" {
		return []string{id}
	}
	return nil
}
