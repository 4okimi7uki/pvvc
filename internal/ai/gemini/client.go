package gemini

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/retry"
	"google.golang.org/api/googleapi"
	"google.golang.org/genai"
)

type Client struct {
	apiKey string
}

func New(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

func (c *Client) Analyze(ctx context.Context, reports []report.DailyReport, update func(string)) (string, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: c.apiKey,
	})
	if err != nil {
		return "", fmt.Errorf("create gemini client: %w", err)
	}

	prompt := buildPrompt(reports)

	// Note: 新しいモデルが出た際はここを更新
	geminiModels := []string{
		"gemini-3-flash-preview",
		"gemini-3.1-flash-lite-preview",
	}

	update("Gemini Thinking...")
	var lastErr error
	var result *genai.GenerateContentResponse
	for i, model := range geminiModels {
		if i > 0 {
			update(fmt.Sprintf("Taking longer than usual... So, Change model: %s", model))
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
		if !isRateLimitError(retryErr) {
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
	return false
}

func buildPrompt(reports []report.DailyReport) string {
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
	sb.WriteString("あなたは、ゴルフメディア「ALBA Net」のシニアデータアナリストです。\n")
	sb.WriteString("GA4とVercelのデータおよびゴルフ大会情報・ニュースを統合し、多忙な担当者が10秒で把握できる「超要約レポート」を作成してください。\n\n")

	sb.WriteString("# 前提・制約\n")
	sb.WriteString("- 現時点で提供できるデータはサイト全体のPV合計とVercelの総コストのみです。\n")
	sb.WriteString("- ページ別・サービス別の詳細データはないため、技術的な原因の断定は行わないでください。\n")
	sb.WriteString("- 分析は「トレンドの把握」と「外部要因との相関」に絞ってください。\n\n")

	sb.WriteString("# サイト・インフラ特性\n")
	sb.WriteString("- ALBA Netはゴルフメディアサイトで、大会開催中は速報・スコアページへのアクセスが集中します。\n")
	sb.WriteString("- Vercelエッジキャッシュが有効なため、同一URLへの繰り返しアクセスはコストにほぼ影響しません。\n")
	sb.WriteString("- 【重要】集中アクセス時はキャッシュが効きコスト効率が良く、大会終了後の分散アクセス時はコスト効率が悪化します。\n")
	sb.WriteString("- 「PVが減ってもコストが上がる」場合は、アクセスが集中から分散に変化した可能性を優先的に考えてください。\n\n")

	sb.WriteString("# 外部背景\n")
	sb.WriteString("- 毎週火曜14:00頃にサイトリリースがあり、水・木はその影響でPV・コストが変動しやすい。\n")
	sb.WriteString("- ゴルフ大会のスケジュールや注目選手の動向がPVに直結します。\n\n")

	sb.WriteString("# 分析対象データ\n")
	sb.WriteString("## Vercel & GA4 集計データ\n")
	sb.WriteString("Date, PV, Cost(USD), Cost(JPY), USD/JPY Rate\n")
	for _, r := range reports {
		fmt.Fprintf(&sb, "%s, %d, %.4f, %.2f, %.2f\n",
			r.Date.Format("2006/01/02"),
			r.PV,
			r.TotalCost,
			r.TotalCostJPY,
			r.Rate,
		)
	}
	sb.WriteString("\n")

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
	sb.WriteString("```\n")
	sb.WriteString("日付        PV          コスト     \n")
	sb.WriteString("04/11(土)   978,567     $197.50    \n")
	sb.WriteString("```\n\n")

	sb.WriteString("### 今週のトレンドとニュース相関\n")
	sb.WriteString("（ゴルフ大会・外部イベントとPV増減の相関を中心に分析してください）\n\n")

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
