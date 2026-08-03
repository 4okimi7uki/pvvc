package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/4okimi7uki/pvvc/internal/chart"
	"github.com/4okimi7uki/pvvc/internal/datasource/ga4"
	"github.com/4okimi7uki/pvvc/internal/report"
	"github.com/4okimi7uki/pvvc/internal/slack"
)

// 取得したレポートの出力先。report / analyze の両コマンドが共通で使う。
// ターミナル・ファイル(SVG)・Slack の3系統を扱う。
//
// コマンド定義ではないので、このファイルに対応するサブコマンドは無い。
// 表示する文言そのものは internal/report/print.go にある。

// --- 出力先の振り分け ---
//
// --svg=- のときは標準出力に SVG 本体が流れるので、装飾的な出力は
// すべて stderr に逃がす必要がある。その判定をここに集約している。

const (
	// svgPathAuto は --svg に値を渡さなかったときに入る歩哨値。
	// pflag の NoOptDefVal に設定しているので、パスとしては扱わない。
	svgPathAuto = "auto"
	// svgPathStdout は標準出力に書き出す指定。
	svgPathStdout = "-"
	// svgAutoDir は --svg に値を渡さなかったときの出力ディレクトリ。
	// カレントに直接ばら撒くと散らかるのでここにまとめる。
	svgAutoDir = "pvvc_svg"
	// htmlAutoDir は --html に値を渡さなかったときの出力ディレクトリ。
	htmlAutoDir = "pvvc_html"
)

// svgRequested は --svg が指定されたか。
func svgRequested() bool { return rootOpts.svgPath != "" }

// svgToStdout は SVG の出力先が標準出力か。
func svgToStdout() bool { return rootOpts.svgPath == svgPathStdout }

// htmlRequested は --html が指定されたか。
func htmlRequested() bool { return rootOpts.htmlPath != "" }

// htmlToStdout は HTML の出力先が標準出力か。
func htmlToStdout() bool { return rootOpts.htmlPath == svgPathStdout }

// dataToStdout は標準出力が成果物（SVG / HTML）で占有されているか。
func dataToStdout() bool { return svgToStdout() || htmlToStdout() }

// chromeOut はロゴや実行時間などの装飾的な出力先。
// 成果物を標準出力に流すときだけ stderr になる。
func chromeOut() io.Writer {
	if dataToStdout() {
		return os.Stderr
	}
	return os.Stdout
}

// printResult はレポート本文をターミナルに出すべきか。
// --svg=- / --html=- のときは成果物と混ざるので出さない。
func printResult() bool { return !quiet && !dataToStdout() }

// --- SVG / HTML ---

// resolveOutPath は --svg / --html の値を実際の出力先に解決する。
// "auto"（値を省略してフラグだけ渡された場合）は dir 配下に期間入りの名前で作る。
// 明示的にパスを渡された場合は dir を無視してそのまま使う。
func resolveOutPath(p string, from, to time.Time, dir, ext string) string {
	if p != svgPathAuto {
		return p
	}
	name := fmt.Sprintf("pvvc-%s_%s%s", from.Format("20060102"), to.Format("20060102"), ext)
	return filepath.Join(dir, name)
}

// resolveSVGPath は --svg の出力先。
func resolveSVGPath(p string, from, to time.Time) string {
	return resolveOutPath(p, from, to, svgAutoDir, ".svg")
}

// resolveHTMLPath は --html の出力先。
func resolveHTMLPath(p string, from, to time.Time, isIndex bool) string {
	if isIndex {
		if p != svgPathAuto {
			return p
		}
		return filepath.Join("web", "index.html")
	}
	return resolveOutPath(p, from, to, htmlAutoDir, ".html")
}

// writeChart は --svg が指定されていればチャートを書き出す。
// 未指定なら何もしない。
func writeChart(reports []report.DailyReport) error {
	if !svgRequested() {
		return nil
	}
	if len(reports) == 0 {
		return fmt.Errorf("svg: no data to plot")
	}

	var (
		path = resolveSVGPath(rootOpts.svgPath, rootOpts.from, rootOpts.to)
		opt  = report.ChartOptions(reports, time.Now())
	)

	if err := writeOut("svg", path, func(w io.Writer) error {
		return chart.RenderSVG(w, opt)
	}); err != nil {
		return err
	}

	if path != svgPathStdout {
		report.FprintSVGBuilt(chromeOut(), path)
	}
	return nil
}

// writeChartPage は --html が指定されていれば、チャートを直接埋め込んだ
// 単体 HTML を書き出す。未指定なら何もしない。
func writeChartPage(reports []report.DailyReport) error {
	if !htmlRequested() {
		return nil
	}
	if len(reports) == 0 {
		return fmt.Errorf("html: no data to plot")
	}

	path := resolveHTMLPath(rootOpts.htmlPath, rootOpts.from, rootOpts.to, true)

	if err := writeOut("html", path, func(w io.Writer) error {
		return chart.RenderPage(w, report.PageData(reports), chart.PageOptions{
			Title:  pageTitle(),
			Origin: cfg.GetString("service.url"),
		})
	}); err != nil {
		return err
	}

	if path != svgPathStdout {
		report.FprintHTMLBuilt(chromeOut(), path)
	}
	return nil
}

// pageTitle は HTML の <title>。サービス名が設定されていれば頭に付ける。
// 空文字を返した場合は chart 側の既定値になる。
func pageTitle() string {
	if name := cfg.GetString("service.name"); name != "" {
		return name
	}
	return ""
}

// writeOut は render の出力を path に書く。"-" なら標準出力。
// label はエラーメッセージの接頭辞（"svg" / "html"）。
func writeOut(label, path string, render func(io.Writer) error) error {
	if path == svgPathStdout {
		return render(os.Stdout)
	}

	// --svg=out/cost.svg のように掘られたパスでも通るようにする。
	// 出力先はユーザーのカレント配下のグラフで、秘匿情報は含まない。
	if dir := filepath.Dir(path); dir != "." {
		//nolint:gosec // G301: public chart output, not sensitive
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("%s: %w", label, err)
		}
	}

	//nolint:gosec // G304: フラグで受け取ったパスに書くのが仕様
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	defer func() { _ = f.Close() }()

	if err := render(f); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	return nil
}

// --- Slack ---

// notifySlack は --notify が指定されていれば Slack にレポートを送る。
// 未指定なら何もしない。aiBody / llm は AI 分析を伴わない report コマンドでは空。
func notifySlack(ctx context.Context, rep []report.DailyReport, topPages []ga4.PagePathRank, aiBody, llm string) error {
	if !rootOpts.notify {
		return nil
	}

	client, err := slack.New(
		cfg.GetString("slack.webhook_url"),
		cfg.GetString("service.name"),
		cfg.GetString("vercel.project_url"),
		cfg.GetString("service.url"),
	)
	if err != nil {
		return err
	}

	if err := client.Send(ctx, aiBody, rootOpts.to, rep, llm, topPages, rootOpts.topPagesLimit); err != nil {
		return err
	}

	report.FprintSlackSent(chromeOut())
	return nil
}
