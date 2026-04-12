package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/4okimi7uki/pvvc/internal/report"
)

type Client struct {
	webhookURL string
	httpClient *http.Client
}

func New(webhookURL string) (*Client, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("slack: webhook url is required")
	}
	return &Client{
		webhookURL: webhookURL,
		httpClient: &http.Client{},
	}, nil
}

type Text struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji"`
}

type RichTextElement struct {
	Type     string            `json:"type"`
	Text     string            `json:"text,omitempty"`
	Elements []RichTextElement `json:"elements,omitempty"`
}

type Block struct {
	Type     string            `json:"type"`
	Text     *Text             `json:"text,omitempty"`
	Elements []RichTextElement `json:"elements,omitempty"`
}

type blockPayload struct {
	Blocks []Block `json:"blocks"`
}

func (c *Client) Send(ctx context.Context, text string) error {
	body, err := json.Marshal(blockPayload{
		Blocks: []Block{
			{
				Type: "header",
				Text: &Text{
					Type:  "plain_text",
					Text:  "📊 P.V.V.C. daily report",
					Emoji: true,
				},
			},
			{
				Type: "divider",
			},
			{
				Type: "rich_text",
				Elements: []RichTextElement{
					{
						Type: "rich_text_section",
						Elements: []RichTextElement{
							{
								Type: "text",
								Text: text,
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("slack: faild to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("slack: request failed %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	report.PrintSection("Notification")
	fmt.Println()
	fmt.Println(" Sent the analysis result to Slack 🕊️")
	fmt.Println()

	return nil
}
