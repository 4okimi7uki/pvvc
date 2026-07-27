package chart

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io"
)

const defaultPageTitle = "GA4 PV × Vercel Cost chart | PVVC"

//go:embed "templates/page.tmpl.html"
var pageTmplSrc string

var pageTmpl = template.Must(template.New("page").Parse(pageTmplSrc))

// PageOptions はチャート以外のページ側の設定。
type PageOptions struct {
	// Title は <title> の中身。空なら defaultPageTitle。
	Title string
}

// pageData はテンプレートに渡す値。
type pageData struct {
	Title string
	Chart template.HTML
}

// RenderPage は SVG を直接埋め込んだ単体 HTML を書き出す。
// 外部ファイルを一切参照しないので file:// でそのまま開ける。
func RenderPage(w io.Writer, opt Options, page PageOptions) error {
	var svg bytes.Buffer
	if err := RenderSVG(&svg, opt); err != nil {
		return err
	}

	if page.Title == "" {
		page.Title = defaultPageTitle
	}

	// RenderSVG の出力は esc / escAttr を通した自前生成の SVG なので、
	// html/template のエスケープを外して生のマークアップとして埋める。
	// ここに外部由来の文字列を素通しで足さないこと。
	data := pageData{Title: page.Title, Chart: template.HTML(svg.String())} //nolint:gosec // G203: 自前生成の SVG を意図的にインライン展開

	// 途中で失敗したときに壊れた HTML を書き残さないよう、いったん溜める。
	var out bytes.Buffer
	out.Grow(svg.Len() + len(pageTmplSrc))
	if err := pageTmpl.Execute(&out, data); err != nil {
		return fmt.Errorf("chart: %w", err)
	}

	_, err := w.Write(out.Bytes())
	return err
}
