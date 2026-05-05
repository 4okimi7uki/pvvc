package claude

import (
	"context"
	"fmt"

	"github.com/4okimi7uki/pvvc/internal/ai"
	"github.com/4okimi7uki/pvvc/internal/httpclient"
	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type Client struct {
	apiKey      string
	serviceName string
	promptPath  string
}

func New(apiKey, serviceName string, promptPath string) *Client {
	return &Client{apiKey: apiKey, serviceName: serviceName, promptPath: promptPath}
}

func (c *Client) Analyze(ctx context.Context, reports []report.DailyReport, update func(string)) (string, error) {
	client := anthropic.NewClient(
		option.WithHTTPClient(httpclient.New()),
		option.WithAPIKey(c.apiKey),
	)

	data := ai.BuildPromptData(reports, c.serviceName)
	prompt, err := ai.BuildPrompt(c.promptPath, data)
	if err != nil {
		return "", fmt.Errorf("build prompt: %w", err)
	}

	// Note: 新しいモデルが出た際はここを更新
	// https://ai.google.dev/claude-api/docs/models
	// claudeModels := []string{
	// 	anthropic.ModelClaudeSonnet4_6,
	// 	anthropic.ModelClaudeOpus4_7,
	// }

	update("Claude Thinking...")
	// var lastErr error
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Model: anthropic.ModelClaudeSonnet4_6,
	})
	if err != nil {
		return "", fmt.Errorf("claude: generate content: %w", err)
	}

	var result string
	for _, block := range message.Content {
		switch variant := block.AsAny().(type) {
		case anthropic.TextBlock:
			result += variant.Text
		}
	}
	return result, nil
	// var result *genai.GenerateContentResponse
	// for i, model := range claudeModels {
	// if i > 0 {
	// 	update(fmt.Sprintf("Taking longer than usual... Switching models to %s", model))
	// }
	// retryErr := retry.Do(ctx, 3, func() error {
	// 	var e error
	// 	result, e = client.Models.GenerateContent(
	// 		ctx,
	// 		model,
	// 		genai.Text(prompt),
	// 		nil,
	// 	)
	// 	return e
	// })
	// if retryErr == nil {
	// 	return result.Text(), nil
	// }
	// if !isRateLimitError(retryErr) || i == len(claudeModels)-1 {
	// 	return "", fmt.Errorf("claude: generate content: %w", retryErr)
	// }
	// lastErr = retryErr
	// }

	// return "", fmt.Errorf("claude: all models exhausted: %w", lastErr)
}

// func isRateLimitError(err error) bool {
// 	var apiErr *googleapi.Error
// 	if errors.As(err, &apiErr) {
// 		return apiErr.Code == 429 || apiErr.Code == 503
// 	}
// 	return strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "RESOURCE_EXHAUSTED")
// }
