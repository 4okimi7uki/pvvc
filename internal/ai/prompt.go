package ai

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/dustin/go-humanize"
)

var weekdaysJa = [...]string{"日", "月", "火", "水", "木", "金", "土"}

// vercelBillingCutoffUTCHour はVercelの課金データが確定するUTC時刻です。
// ref: https://vercel.com/docs/billing
const vercelBillingCutoffUTCHour = 7

var newsUrlList = []string{
	"https://www.pgatour.com/news",
	"https://www.lpga.com/news",
	"https://www.livgolf.com/news",
	"https://www.lpga.or.jp/news/news_and_topics/",
	"https://www.alba.co.jp/tour/category/next/schedule/",
	"https://golf.com/",
	"https://news.golfdigest.co.jp/search/",
	"https://www.alba.co.jp/",
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

const rowFmt = "%-7s  %-12s  %-12s  %-14s  %-12s  %s"
const serviceRowFmt = "%-40s %s"

func BuildPromptData(reports []report.DailyReport, serviceName string, end time.Time) PromptData {
	rows := make([]ReportRow, len(reports))
	var serviceTableRows []ReportRow
	latestDate := end.AddDate(0, 0, -1).Format("20060102")
	latestCostByService := reports[0].CostByServices[latestDate]

	for i, r := range reports {
		date := r.Date.Format("01/02") + fmt.Sprintf("(%s)", weekdaysJa[r.Date.Weekday()])
		pv := humanize.Comma(r.PV)
		cost := "$" + humanize.CommafWithDigits(r.TotalCost, 2)
		costJPY := "¥" + humanize.CommafWithDigits(r.TotalCostJPY, 0)
		costPerPV := "¥" + humanize.CommafWithDigits(r.TotalCostJPY/float64(r.PV), 4)
		rate := humanize.CommafWithDigits(r.Rate, 2)

		rows[i] = ReportRow{
			Line: fmt.Sprintf(rowFmt, date, pv, cost, costJPY, costPerPV, rate),
		}
	}
	for _, l := range latestCostByService {
		serviceTableRows = append(serviceTableRows, ReportRow{
			Line: fmt.Sprintf(serviceRowFmt, l.ServiceName, "$"+humanize.FtoaWithDigits(l.BilledCost, 4)),
		})
	}

	return PromptData{
		ServiceName:        serviceName,
		Today:              time.Now().Format("2006年01月02日"),
		TableHeader:        fmt.Sprintf(rowFmt, "日付", "PV", "Cost(USD)", "Cost(JPY)", "Cost/PV", "USD/JPY"),
		Rows:               rows,
		ServiceTableHeader: fmt.Sprintf(serviceRowFmt, "SERVICE NAME", "BILLED COST"),
		ServiceTableRows:   serviceTableRows,
		NewsURLs:           newsUrlList,
		IsBeforeCutoff:     time.Now().UTC().Hour() < vercelBillingCutoffUTCHour,
		HasAnomaly:         detectAnomaly(reports),
	}
}

//go:embed "templates/analyze.tmpl"
var defaultPrompt embed.FS

func BuildPrompt(tmplPath string, data PromptData) (string, error) {
	var tmplBytes []byte
	var err error

	switch {
	case strings.HasPrefix(tmplPath, "http://") || strings.HasPrefix(tmplPath, "https://"):
		resp, fetchErr := http.Get(tmplPath)
		if fetchErr != nil {
			return "", fmt.Errorf("failed to fetch template: %w", fetchErr)
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		tmplBytes, err = io.ReadAll(resp.Body)
	case tmplPath != "":
		tmplBytes, err = os.ReadFile(tmplPath)
	default:
		tmplBytes, err = defaultPrompt.ReadFile("templates/analyze.tmpl") // fallback
	}
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	tmpl, err := template.New("prompt").Parse(string(tmplBytes))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	fmt.Println(buf.String())

	return buf.String(), nil
}
