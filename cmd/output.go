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
)

// svgRequested は --svg が指定されたか。
func svgRequested() bool { return rootOpts.svgPath != "" }

// svgToStdout は SVG の出力先が標準出力か。
func svgToStdout() bool { return rootOpts.svgPath == svgPathStdout }

// chromeOut はロゴや実行時間などの装飾的な出力先。
// SVG を標準出力に流すときだけ stderr になる。
func chromeOut() io.Writer {
	if svgToStdout() {
		return os.Stderr
	}
	return os.Stdout
}

// printResult はレポート本文をターミナルに出すべきか。
// --svg=- のときは SVG と混ざるので出さない。
func printResult() bool { return !quiet && !svgToStdout() }

// --- SVG ---

// resolveSVGPath は --svg の値を実際の出力先に解決する。
// "auto"（値を省略して --svg だけ渡された場合）は pvvc_svg/ に期間入りの名前で作る。
// 明示的にパスを渡された場合は svgAutoDir を無視してそのまま使う。
func resolveSVGPath(p string, from, to time.Time) string {
	if p != svgPathAuto {
		return p
	}
	name := fmt.Sprintf("pvvc-%s_%s.svg", from.Format("20060102"), to.Format("20060102"))
	return filepath.Join(svgAutoDir, name)
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

	if path == svgPathStdout {
		return chart.RenderSVG(os.Stdout, opt)
	}

	// --svg=out/cost.svg のように掘られたパスでも通るようにする。
	// 出力先はユーザーのカレント配下のグラフ画像で、秘匿情報は含まない。
	if dir := filepath.Dir(path); dir != "." {
		//nolint:gosec // G301: public chart output, not sensitive
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("svg: %w", err)
		}
	}

	//nolint:gosec // G304: --svg で受け取ったパスに書くのが仕様
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("svg: %w", err)
	}
	defer func() { _ = f.Close() }()

	if err := chart.RenderSVG(f, opt); err != nil {
		return fmt.Errorf("svg: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("svg: %w", err)
	}

	report.FprintSVGBuilt(chromeOut(), path)
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
