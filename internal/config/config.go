package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func New() *viper.Viper {
	_ = godotenv.Load()

	v := viper.New()
	v.AutomaticEnv()

	v.BindEnv("vercel.token", "VERCEL_TOKEN")
	v.BindEnv("vercel.team_id", "TEAM_ID")
	v.BindEnv("ga4.property_id", "PROPERTY_ID")

	return v
}
