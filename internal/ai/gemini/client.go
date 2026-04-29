package gemini

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/4okimi7uki/pvvc/internal/ai"
	"github.com/4okimi7uki/pvvc/internal/httpclient"
	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/retry"
	"google.golang.org/api/googleapi"
	"google.golang.org/genai"
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
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:     c.apiKey,
		HTTPClient: httpclient.New(),
	})
	if err != nil {
		return "", fmt.Errorf("create gemini client: %w", err)
	}

	data := ai.BuildPromptData(reports, c.serviceName)
	prompt, err := ai.BuildPrompt(c.promptPath, data)
	if err != nil {
		return "", fmt.Errorf("build prompt: %w", err)
	}

	// Note: 新しいモデルが出た際はここを更新
	// https://ai.google.dev/gemini-api/docs/models
	geminiModels := []string{
		"gemini-3-flash-preview",
		"gemini-3.1-flash-lite-preview",
		"gemini-2.5-flash",
	}

	update("Gemini Thinking...")
	var lastErr error
	var result *genai.GenerateContentResponse
	for i, model := range geminiModels {
		if i > 0 {
			update(fmt.Sprintf("Taking longer than usual... Switching models to %s", model))
		}
		retryErr := retry.Do(ctx, 3, func() error {
			var e error
			result, e = client.Models.GenerateContent(
				ctx,
				model,
				genai.Text(prompt),
				nil,
			)
			return e
		})
		if retryErr == nil {
			return result.Text(), nil
		}
		if !isRateLimitError(retryErr) || i == len(geminiModels)-1 {
			return "", fmt.Errorf("gemini: generate content: %w", retryErr)
		}
		lastErr = retryErr
	}

	return "", fmt.Errorf("gemini: all models exhausted: %w", lastErr)
}

func isRateLimitError(err error) bool {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.Code == 429 || apiErr.Code == 503
	}
	return strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "RESOURCE_EXHAUSTED")
}
