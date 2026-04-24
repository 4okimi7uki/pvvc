package gemini

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/retry"
	"github.com/dustin/go-humanize"
	"google.golang.org/api/googleapi"
	"google.golang.org/genai"
)

var weekdaysJa = [...]string{"日", "月", "火", "水", "木", "金", "土"}

type Client struct {
	apiKey      string
	serviceName string
}

func New(apiKey, serviceName string) *Client {
	return &Client{apiKey: apiKey, serviceName: serviceName}
}

func (c *Client) Analyze(ctx context.Context, reports []report.DailyReport, update func(string)) (string, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: c.apiKey,
	})
	if err != nil {
		return "", fmt.Errorf("create gemini client: %w", err)
	}

	prompt := buildPrompt(reports, c.serviceName)

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

// vercelBillingCutoffUTCHour はVercelの課金データが確定するUTC時刻です。
// ref: https://vercel.com/docs/billing
const vercelBillingCutoffUTCHour = 7

func buildPrompt(reports []report.DailyReport, serviceName string) string {
	var sb strings.Builder

	newsUrlList := []string{
		"https://www.pgatour.com/news",
		"https://www.lpga.com/news",
		"https://www.livgolf.com/news",
		"https://www.lpga.or.jp/news/news_and_topics/",
		"https://www.alba.co.jp/tour/category/next/schedule/",
		"https://golf.com/",
		"https://news.golfdigest.co.jp/search/",
		"https://www.alba.co.jp/",
	}

	sb.WriteString("# 役割\n")
	fmt.Fprintf(&sb, "あなたは、ゴルフメディア「%s」のシニアデータアナリストです。\n", serviceName)
	sb.WriteString("GA4とVercelのデータおよびゴルフ大会情報・ニュースを統合し、多忙な担当者が10秒で把握できる「超要約レポート」を作成してください。\n\n")

	sb.WriteString("# 前提・制約\n")
	sb.WriteString("- 提供データ（PV・コスト）と、**ゴルフニュースソースで得た「今年」の事実**のみを結合してください。\n")
	sb.WriteString("- 【ハルシネーション厳禁】最新のニュースソースで今週開催されていることが確認できない大会名は、1文字も出力してはいけません。\n")
	sb.WriteString("- 大会名が100%確実でない場合は、代わりに「ツアー」「大会」「国内・海外ツアー」など、一般名詞を必ず使用してください。\n")
	sb.WriteString("- 記述の根拠となるニュースが見つからない場合は、推測せず「現在、特定の大会要因は確認できていません」と明記してください。\n")
	fmt.Fprintf(&sb, "- 本日の日付は %s です。これより過去や未来の大会スケジュールと混同しないでください。\n", time.Now().Format("2006年01月02日"))

	sb.WriteString("# サイト・インフラ特性\n")
	fmt.Fprintf(&sb, "- %sはゴルフメディアサイトで、大会開催中は速報・スコアページへのアクセスが集中します。\n", serviceName)
	sb.WriteString("- Vercelエッジキャッシュが有効なため、同一URLへの繰り返しアクセスはコストにほぼ影響しません。\n")
	sb.WriteString("- 【重要】集中アクセス時はキャッシュが効きコスト効率が良く、大会終了後の分散アクセス時はコスト効率が悪化します。\n")
	sb.WriteString("- 「PVが減ってもコストが上がる」場合は、アクセスが集中から分散に変化した可能性を優先的に考えてください。\n\n")

	sb.WriteString("# 外部背景\n")
	sb.WriteString("- 毎週火曜14:00頃にサイトリリースがあり、水・木はその影響でPV・コストが変動しやすい。\n")
	sb.WriteString("- ゴルフ大会のスケジュールや注目選手の動向がPVに直結します。\n\n")

	sb.WriteString("# 分析対象データ\n")
	sb.WriteString("## Vercel & GA4 集計データ\n")
	const rowFmt = "%-11s  %-12s  %-12s  %-14s  %-12s  %s\n"
	sb.WriteString("```\n")
	fmt.Fprintf(&sb, rowFmt, "日付", "PV", "Cost(USD)", "Cost(JPY)", "Cost/PV", "USD/JPY")
	for _, r := range reports {
		dateStr := r.Date.Format("01/02") + fmt.Sprintf("(%s)", weekdaysJa[r.Date.Weekday()])
		fmt.Fprintf(&sb, rowFmt,
			dateStr,
			humanize.Comma(r.PV),
			"$"+humanize.CommafWithDigits(r.TotalCost, 2),
			"¥"+humanize.CommafWithDigits(r.TotalCostJPY, 0),
			"¥"+humanize.CommafWithDigits(r.TotalCostJPY/float64(r.PV), 4),
			humanize.CommafWithDigits(r.Rate, 2),
		)
	}
	sb.WriteString("```\n\n")

	sb.WriteString("## ゴルフニュースソース\n")
	sb.WriteString("以下のURLから最新情報を確認し、PV増減の背景（大会の有無・注目選手の結果等）を把握してください。\n")
	for _, news := range newsUrlList {
		fmt.Fprintf(&sb, " - %s\n", news)
	}
	sb.WriteString("\n")

	sb.WriteString("# 出力ルール\n")
	sb.WriteString("1. 各セクションの本文は【1〜2行】で簡潔に記述してください。\n")
	sb.WriteString("2. '*' 記号は使用禁止。代わりに【】や絵文字（📈📉⚠️⛳️）を活用してください。\n")
	sb.WriteString("3. Markdownテーブル（|---|）は絶対に使用しないでください。\n")
	sb.WriteString("4. データ表はコードブロック（```）内にスペースで列を揃えた等幅テキストで出力してください。\n\n")

	sb.WriteString("# 出力形式\n\n")

	sb.WriteString("### 昨日（昨日の日付）の分析結果\n")
	sb.WriteString("（最新日のPV・コストを前日比で簡潔に評価してください）\n\n")

	sb.WriteString("### 直近の推移データ\n")
	sb.WriteString("（上記の集計データをそのままコードブロック内に転記してください）\n\n")

	sb.WriteString("### 今週のトレンドとニュース相関\n")
	sb.WriteString("（ゴルフ大会・外部イベントとPV増減の相関を中心に分析してください）\n\n")

	if time.Now().UTC().Hour() < vercelBillingCutoffUTCHour {
		sb.WriteString("🕖 Vercelの課金データは7:00 UTC（16:00 JST）以降に確定します。このレポートはその前に実行されているため、最新日のコストは暫定値です。\n")
	}
	sb.WriteString("\n")

	if detectAnomaly(reports) {
		sb.WriteString("### ⚠️ 異常検知\n")
		sb.WriteString("直近のコストに通常比1.3倍以上の急増が検出されています。\n")
		sb.WriteString("現時点のデータでは技術的原因の特定は困難ですが、気になる点があれば簡潔に記述してください。\n\n")
	}

	return sb.String()
}

func detectAnomaly(reports []report.DailyReport) bool {
	if len(reports) < 2 {
		return false
	}
	// 直近2日のコスト比較で1.3倍以上なら異常
	const anomalyCriteria = 1.3
	latest := reports[len(reports)-1]
	prev := reports[len(reports)-2]
	if prev.TotalCost == 0 {
		return false
	}
	return latest.TotalCost/prev.TotalCost >= anomalyCriteria
}
