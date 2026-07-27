package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/4okimi7uki/pvvc/internal/report"
)

func TestResolveSVGPath(t *testing.T) {
	var (
		from = time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		to   = time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
	)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"--svg だけ", svgPathAuto, "pvvc_svg/pvvc-20260501_20260726.svg"},
		{"パス指定", "docs/cost.svg", "docs/cost.svg"},
		{"標準出力", "-", "-"},
		{"絶対パス", "/tmp/a.svg", "/tmp/a.svg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveSVGPath(tt.in, from, to); got != tt.want {
				t.Errorf("resolveSVGPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestResolveHTMLPath(t *testing.T) {
	var (
		from = time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		to   = time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
	)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"--html だけ", svgPathAuto, "pvvc_html/pvvc-20260501_20260726.html"},
		{"パス指定", "docs/chart.html", "docs/chart.html"},
		{"標準出力", "-", "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveHTMLPath(tt.in, from, to); got != tt.want {
				t.Errorf("resolveHTMLPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestWriteChartPageSkippedWhenFlagUnset(t *testing.T) {
	defer func(p string) { rootOpts.htmlPath = p }(rootOpts.htmlPath)
	rootOpts.htmlPath = ""

	if err := writeChartPage(nil); err != nil {
		t.Errorf("writeChartPage() = %v, want nil", err)
	}
}

func TestWriteChartPageErrorsOnEmptyData(t *testing.T) {
	defer func(p string) { rootOpts.htmlPath = p }(rootOpts.htmlPath)
	rootOpts.htmlPath = svgPathAuto

	if err := writeChartPage(nil); err == nil {
		t.Error("データが無いときはエラーを期待したが nil")
	}
}

// SVG を外部参照せず、HTML 1枚に閉じて書けていること。
func TestWriteChartPageInlinesSVG(t *testing.T) {
	defer func(p string) { rootOpts.htmlPath = p }(rootOpts.htmlPath)

	path := filepath.Join(t.TempDir(), "out", "chart.html")
	rootOpts.htmlPath = path

	if err := writeChartPage(sampleReports(7)); err != nil {
		t.Fatalf("writeChartPage() = %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("出力ファイルが読めない: %v", err)
	}
	if !bytes.HasPrefix(b, []byte("<!doctype html>")) {
		t.Errorf("HTML になっていない: %.40s", b)
	}
	if !bytes.Contains(b, []byte("<svg")) || !bytes.Contains(b, []byte("<polyline")) {
		t.Error("チャートが埋め込まれていない")
	}
}

// --html=- でも装飾的な出力が標準出力を汚さないこと。
func TestChromeOutGoesToStderrWhenHTMLOnStdout(t *testing.T) {
	defer func(s, h string) { rootOpts.svgPath, rootOpts.htmlPath = s, h }(rootOpts.svgPath, rootOpts.htmlPath)

	rootOpts.svgPath = ""
	rootOpts.htmlPath = svgPathStdout
	if got := chromeOut(); got != os.Stderr {
		t.Errorf("chromeOut() = %v, want os.Stderr", got)
	}
	if printResult() {
		t.Error("printResult() = true, want false")
	}

	rootOpts.htmlPath = "out.html"
	if got := chromeOut(); got != os.Stdout {
		t.Errorf("chromeOut() = %v, want os.Stdout", got)
	}
}

// --svg=- のときだけ、レポート本文とロゴが stdout を避けること。
func TestSVGToStdoutSuppressesTerminalOutput(t *testing.T) {
	tests := []struct {
		svgPath         string
		quiet           bool
		wantPrint       bool
		wantChromeIsErr bool
	}{
		{"", false, true, false},
		{"", true, false, false},
		{"out.svg", false, true, false},
		{"-", false, false, true},
		{"-", true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.svgPath+"/quiet="+boolStr(tt.quiet), func(t *testing.T) {
			defer func(p string, q bool) { rootOpts.svgPath, quiet = p, q }(rootOpts.svgPath, quiet)
			rootOpts.svgPath, quiet = tt.svgPath, tt.quiet

			if got := printResult(); got != tt.wantPrint {
				t.Errorf("printResult() = %v, want %v", got, tt.wantPrint)
			}
			if got := svgToStdout(); got != tt.wantChromeIsErr {
				t.Errorf("svgToStdout() = %v, want %v", got, tt.wantChromeIsErr)
			}
		})
	}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func TestWriteChartSkippedWhenFlagUnset(t *testing.T) {
	defer func(p string) { rootOpts.svgPath = p }(rootOpts.svgPath)
	rootOpts.svgPath = ""

	// データが空でも --svg 未指定なら何もせず nil を返す。
	if err := writeChart(nil); err != nil {
		t.Errorf("writeChart() = %v, want nil", err)
	}
}

func TestWriteChartErrorsOnEmptyData(t *testing.T) {
	defer func(p string) { rootOpts.svgPath = p }(rootOpts.svgPath)
	rootOpts.svgPath = svgPathAuto

	if err := writeChart(nil); err == nil {
		t.Error("データが無いときはエラーを期待したが nil")
	}
}

// --svg（値なし）はカレントではなく pvvc_svg/ 配下に書くこと。
func TestWriteChartAutoUsesOutputDir(t *testing.T) {
	defer func(p string, f, to time.Time) {
		rootOpts.svgPath, rootOpts.from, rootOpts.to = p, f, to
	}(rootOpts.svgPath, rootOpts.from, rootOpts.to)

	t.Chdir(t.TempDir())
	rootOpts.svgPath = svgPathAuto
	rootOpts.from = time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	rootOpts.to = time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	if err := writeChart(sampleReports(7)); err != nil {
		t.Fatalf("writeChart() = %v", err)
	}

	want := filepath.Join(svgAutoDir, "pvvc-20260501_20260726.svg")
	if _, err := os.Stat(want); err != nil {
		entries, _ := os.ReadDir(".")
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("%s が無い（カレントの中身: %v）", want, names)
	}
}

// 存在しないディレクトリを含むパスでも掘って書けること。
func TestWriteChartCreatesNestedDir(t *testing.T) {
	defer func(p string) { rootOpts.svgPath = p }(rootOpts.svgPath)

	path := filepath.Join(t.TempDir(), "out", "charts", "cost.svg")
	rootOpts.svgPath = path

	if err := writeChart(sampleReports(7)); err != nil {
		t.Fatalf("writeChart() = %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("出力ファイルが読めない: %v", err)
	}
	if !bytes.HasPrefix(b, []byte("<svg")) {
		t.Errorf("SVG になっていない: %.40s", b)
	}
	if !bytes.Contains(b, []byte("<polyline")) {
		t.Error("PV の折れ線が描かれていない")
	}
}

// sampleReports は n 日分のダミーレポート。APIは叩かない。
func sampleReports(n int) []report.DailyReport {
	base := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	out := make([]report.DailyReport, 0, n)
	for i := range n {
		out = append(out, report.DailyReport{
			Date:      base.AddDate(0, 0, i),
			PV:        decimal.NewFromInt(int64(700_000 + i*10_000)),
			TotalCost: decimal.NewFromFloat(120 + float64(i)*3.5),
		})
	}
	return out
}

// --svg=- のとき、装飾的な出力が標準出力を汚さないこと。
// SVG 以外が1バイトでも混ざるとパイプ先のファイルが壊れる。
func TestChromeOutGoesToStderrWhenSVGOnStdout(t *testing.T) {
	defer func(p string) { rootOpts.svgPath = p }(rootOpts.svgPath)

	rootOpts.svgPath = svgPathStdout
	if got := chromeOut(); got != os.Stderr {
		t.Errorf("chromeOut() = %v, want os.Stderr", got)
	}

	rootOpts.svgPath = "out.svg"
	if got := chromeOut(); got != os.Stdout {
		t.Errorf("chromeOut() = %v, want os.Stdout", got)
	}
}

// ステータス表示が chromeOut() を通っていること。
// 直接 os.Stdout を掴むコードが戻ってきたらここで落ちる。
func TestDecorationsDoNotTouchStdoutInPipeMode(t *testing.T) {
	defer func(p string) { rootOpts.svgPath = p }(rootOpts.svgPath)
	rootOpts.svgPath = svgPathStdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = r.Close() }()

	orig := os.Stdout
	os.Stdout = w
	report.FprintSlackSent(chromeOut())
	os.Stdout = orig
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("標準出力に %d バイト漏れている: %q", buf.Len(), buf.String())
	}
}
