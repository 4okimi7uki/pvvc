package config

type VercelConfig struct {
	Token      string `toml:"token"`
	TeamId     string `toml:"team_id"`
	ProjectId  string `toml:"project_id"`
	ProjectIds string `toml:"project_ids"`
}

type Ga4Config struct {
	PropertyId string `toml:"property_id"`
	Credential string `toml:"credential"`
}

type AiConfig struct {
	GeminiKey string `toml:"gemini_key"`
}

type SlackConfig struct {
	WebhookUrl string `toml:"webhook_url"`
}

type ServiceConfig struct {
	Name string `toml:"name"`
}

type Config struct {
	Vercel  VercelConfig  `toml:"vercel"`
	Ga4     Ga4Config     `toml:"ga4"`
	Ai      AiConfig      `toml:"ai"`
	Slack   SlackConfig   `toml:"slack"`
	Service ServiceConfig `toml:"service"`
}
