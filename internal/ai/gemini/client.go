package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/retry"
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

	var result *genai.GenerateContentResponse
	if err := retry.Do(ctx, 3, func() error {
		var e error
		result, e = client.Models.GenerateContent(
			ctx,
			"gemini-3.1-flash-lite-preview",
			genai.Text(prompt),
			nil,
		)
		return e
	}); err != nil {
		return "", fmt.Errorf("gemini: generate content: %w", err)
	}

	return result.Text(), nil
}

func buildPrompt(reports []report.DailyReport) string {
	var sb strings.Builder

	// sb.WriteString("Slackの Block Kit (JSON形式) で作成してください。セクション、区切り線（divider）、ボタンなどを含めて、視認性の高いレイアウトにしてください。\n")
	// sb.WriteString("'*'は、見ずらいのであまり使用しないでください。視認性の高いレイアウトにしてください。\n")
	// sb.WriteString("以下はゴルフメディアサイトの直近のPV数とVercelのホスティングコストのデータです。\n")
	// sb.WriteString("昨日のデータの分析を主軸にしてください。\n\n")
	// sb.WriteString("このデータをもとに、トレンドや気になる点を簡潔に分析してください。\n\n")
	// sb.WriteString("Date, PV, Cost(USD), Cost(JPY), USD/JPY Rate\n")
	// sb.WriteString("先週と比べて今週はどういう傾向にあるか、考慮してください。\n\n")
	// sb.WriteString("毎週火曜日14時に、最新ソースを本番環境（このWebサイト）へ反映します。水曜・木曜に関して、リリースの影響はありそうか、分析してください。\n\n")
	// sb.WriteString("分析内容のみを出力してください。\n\n")
	//
	sb.WriteString("# 役割\n")
	sb.WriteString("あなたは、ゴルフメディア「ALBA Net」のシニアデータアナリストです。\n")
	sb.WriteString("GA4とVercelのデータを統合し、多忙な担当者が10秒で把握できる「超要約レポート」を作成してください。\n\n")

	sb.WriteString("# 外部背景\n")
	sb.WriteString("- 毎週火曜14:00頃リリース。水・木はその影響（コスト・PV変動）を注視。\n")
	sb.WriteString("- ゴルフ大会スケジュールや外部スポーツ（大谷選手等）のイベント性を考慮。\n\n")

	sb.WriteString("# 分析対象データ\n")
	sb.WriteString("Date, PV, Cost(USD), Cost(JPY), USD/JPY Rate\n")
	// ここに取得したデータを流し込む
	for _, r := range reports {
		fmt.Fprintf(&sb, "%s, %d, %.4f, %.2f, %.2f\n",
			r.Date.Format("2006/01/02"),
			r.PV,
			r.TotalCost,
			r.TotalCostJPY,
			r.Rate,
		)
	}
	sb.WriteString("# 依頼事項\n")
	sb.WriteString("各セクションの本文は必ず【1行】で記述してください。\n")
	sb.WriteString("'*'記号は視認性を下げるため、リストの箇条書きや強調には使用しないでください。\n")
	sb.WriteString("表形式や適切な改行、【】などの記号を活用し、一目で数値の変化がわかるレイアウトにしてください。\n\n")

	sb.WriteString("# 出力形式（このフォーマットを厳守）\n")

	sb.WriteString("### 昨日の分析結果\n")
	sb.WriteString("【最新日データ】PV 000（前日比±0%）、コスト $000（前日比±0%）。効率性は改善/悪化。\n\n")

	sb.WriteString("### 今週のトレンド\n")
	sb.WriteString("【推移】〇月〇日のピーク以降、PVは〇〇の影響で微減。コスト単価は安定/不安定。\n\n")

	sb.WriteString("### リリース影響・異常検知\n")
	sb.WriteString("【検証】火曜リリースの影響は見られず正常。もしくは、〇曜日のコスト急増につき要因調査を推奨。\n")

	return sb.String()
}
