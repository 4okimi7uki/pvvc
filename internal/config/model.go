package config

type VercelConfig struct {
	Token      string `toml:"token"`
	TeamID     string `toml:"team_id"`
	ProjectID  string `toml:"project_id"`
	ProjectIDs string `toml:"project_ids"`
	ProjectURL string `toml:"project_url"`
}

type Ga4Config struct {
	PropertyID string `toml:"property_id"`
	Credential string `toml:"credential"`
}

type AiConfig struct {
	GeminiKey string `toml:"gemini_key"`
	ClaudeKey string `toml:"claude_key"`
}

type SlackConfig struct {
	WebhookURL string `toml:"webhook_url"`
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
