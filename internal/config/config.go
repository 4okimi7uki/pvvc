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
	_ = v.BindEnv("ga4.property_id", "PROPERTY_ID")
	_ = v.BindEnv("ga4.credential", "GOOGLE_ANALYTICS_CREDENTIAL")
	_ = v.BindEnv("ai.gemini_key", "GEMINI_API_KEY")
	_ = v.BindEnv("slack.webhook_url", "SLACK_WEBHOOK_URL")
	_ = v.BindEnv("service.name", "TARGET_WEBSITE_NAME")

	return v
}
