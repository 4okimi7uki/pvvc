package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/4okimi7uki/pvvc/internal/report"
	"google.golang.org/genai"
)

type Client struct {
	apiKey string
}

func New(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

func (c *Client) Analyze(ctx context.Context, reports []report.DailyReport) (string, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: c.apiKey,
	})
	if err != nil {
		return "", fmt.Errorf("create gemini client: %w", err)
	}

	prompt := buildPrompt(reports)

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-3.1-flash-lite-preview",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("generate content: %w", err)
	}

	return result.Text(), nil
}

func buildPrompt(reports []report.DailyReport) string {
	var sb strings.Builder

	sb.WriteString("Slackの Block Kit (JSON形式) で作成してください。セクション、区切り線（divider）、ボタンなどを含めて、視認性の高いレイアウトにしてください。\n")
	// sb.WriteString("'*'は、見ずらいのであまり使用しないでください。箇条書きしたい時は'-'を使用してください。\n")
	sb.WriteString("以下はWebサービスの直近のPV数とVercelのホスティングコストのデータです。\n")
	sb.WriteString("このデータをもとに、トレンドや気になる点を簡潔に分析してください。\n\n")
	sb.WriteString("Date, PV, Cost(USD), Cost(JPY), Rate\n")
	sb.WriteString("先週と比べて今週はどういう傾向にあるか、考慮してください。\n\n")

	for _, r := range reports {
		sb.WriteString(fmt.Sprintf("%s, %d, %.4f, %.2f, %.2f\n",
			r.Date.Format("2006/01/02"),
			r.PV,
			r.TotalCost,
			r.TotalCostJPY,
			r.Rate,
		))
	}

	return sb.String()
}
