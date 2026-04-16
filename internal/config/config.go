package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func New() *viper.Viper {
	_ = godotenv.Load()

	v := viper.New()
	v.AutomaticEnv()

	_ = v.BindEnv("vercel.token", "VERCEL_TOKEN")
	_ = v.BindEnv("vercel.team_id", "TEAM_ID")
	_ = v.BindEnv("vercel.project_id", "PROJECT_ID")
	_ = v.BindEnv("ga4.property_id", "PROPERTY_ID")
	_ = v.BindEnv("ga4.credential", "GOOGLE_ANALYTICS_CREDENTIAL")
	_ = v.BindEnv("ai.gemini_key", "GEMINI_API_KEY")
	_ = v.BindEnv("slack.webhook_url", "SLACK_WEBHOOK_URL")
	_ = v.BindEnv("service.name", "TARGET_WEBSITE_NAME")

	return v
}

// Warnings returns alert messages for env vars that are unset but would cause
// silent wrong output (e.g. costs showing as $0).
func Warnings(v *viper.Viper) []string {
	var warns []string
	if v.GetString("vercel.project_id") == "" {
		warns = append(warns, "PROJECT_ID is not set — Vercel costs will show as $0 for all days")
	}
	if v.GetString("vercel.team_id") == "" {
		warns = append(warns, "TEAM_ID is not set — fetching personal account billing (not team)")
	}
	return warns
}
