package report

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// fixedNow は生成時刻の固定値。time.Now() を使うとテストが日替わりで壊れる。
var fixedNow = time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)

func days(n int) []DailyReport {
	base := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	out := make([]DailyReport, 0, n)
	for i := range n {
		out = append(out, DailyReport{
			Date:      base.AddDate(0, 0, i),
			PV:        decimal.NewFromInt(int64(700_000 + i*1_000)),
			TotalCost: decimal.NewFromFloat(120 + float64(i)),
		})
	}
	return out
}

func TestChartOptions(t *testing.T) {
	reports := days(92)
	opt := ChartOptions(reports, fixedNow)

	if len(opt.Bars.Values) != 92 {
		t.Errorf("棒が %d 本、期待は 92 本", len(opt.Bars.Values))
	}
	if len(opt.Line.Values) != 92 {
		t.Errorf("折れ線が %d 点、期待は 92 点", len(opt.Line.Values))
	}
	if opt.Bars.Values[0] != 120 {
		t.Errorf("Bars.Values[0] = %v, want 120", opt.Bars.Values[0])
	}
	if opt.Line.Values[0] != 700_000 {
		t.Errorf("Line.Values[0] = %v, want 700000", opt.Line.Values[0])
	}
	// 期間が aria-label に入っていること
	if want := "2026-05-01 to 2026-07-31"; !strings.Contains(opt.AriaLabel, want) {
		t.Errorf("AriaLabel = %q, %q を含んでほしい", opt.AriaLabel, want)
	}
}

func TestChartOptionsEmpty(t *testing.T) {
	opt := ChartOptions(nil, fixedNow)
	if len(opt.Bars.Values) != 0 || len(opt.XLabels) != 0 {
		t.Error("空入力なのに値が入っている")
	}
	if opt.AriaLabel == "" {
		t.Error("空入力でも aria-label は欲しい")
	}
}

func TestXLabels(t *testing.T) {
	tests := []struct {
		days      int
		limit     int
		wantCount int
		wantFirst int
		wantLast  int
	}{
		{92, 8, 8, 0, 91},
		{7, 8, 7, 0, 6},   // 少ないときは全部
		{8, 8, 8, 0, 7},   // ちょうど
		{9, 8, 8, 0, 8},   // 1つ間引く
		{1, 8, 1, 0, 0},   // 1日だけ
		{30, 2, 2, 0, 29}, // 両端だけ
	}

	for _, tt := range tests {
		got := xLabels(days(tt.days), tt.limit)

		if len(got) != tt.wantCount {
			t.Errorf("days=%d limit=%d: ラベル %d 個、期待は %d 個", tt.days, tt.limit, len(got), tt.wantCount)
			continue
		}
		if got[0].At != tt.wantFirst {
			t.Errorf("days=%d: 先頭が %d、期待は %d", tt.days, got[0].At, tt.wantFirst)
		}
		if got[len(got)-1].At != tt.wantLast {
			t.Errorf("days=%d: 末尾が %d、期待は %d", tt.days, got[len(got)-1].At, tt.wantLast)
		}

		// 単調増加、かつ範囲内であること
		for i, l := range got {
			if l.At < 0 || l.At >= tt.days {
				t.Errorf("days=%d: At=%d が範囲外", tt.days, l.At)
			}
			if i > 0 && l.At <= got[i-1].At {
				t.Errorf("days=%d: At が単調増加していない (%d -> %d)", tt.days, got[i-1].At, l.At)
			}
			if l.Text == "" {
				t.Errorf("days=%d: ラベルが空", tt.days)
			}
		}
	}
}

func TestXLabelsNoData(t *testing.T) {
	if got := xLabels(nil, 8); got != nil {
		t.Errorf("xLabels(nil) = %v, want nil", got)
	}
}

func TestChartCaption(t *testing.T) {
	tests := []struct {
		name    string
		reports []DailyReport
		want    string
	}{
		{
			"92日",
			days(92),
			"2026-05-01 → 2026-07-31 (92 days)  ·  generated 2026-07-27 10:00 UTC",
		},
		{
			"1日だけ",
			days(1),
			"2026-05-01  ·  generated 2026-07-27 10:00 UTC",
		},
		{"データなし", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chartCaption(tt.reports, fixedNow); got != tt.want {
				t.Errorf("chartCaption() = %q,\n                want %q", got, tt.want)
			}
		})
	}
}

// Caption が Options に載っていること。
func TestChartOptionsIncludesCaption(t *testing.T) {
	opt := ChartOptions(days(92), fixedNow)
	if !strings.Contains(opt.Caption, "92 days") {
		t.Errorf("Caption = %q, 期間が入っていない", opt.Caption)
	}
	if !strings.Contains(opt.Caption, "generated") {
		t.Errorf("Caption = %q, 生成時刻が入っていない", opt.Caption)
	}
}
