package chart

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
)

const defaultPageTitle = "GA4 PV × Vercel Cost"

//go:embed "templates/page.tmpl.html"
var pageTmplSrc string

// pvvcChartJS はブラウザ側の描画スクリプト。生成 HTML にインライン展開する。
//
//go:embed "templates/pvvc-chart.js"
var pvvcChartJS string

//
//go:embed "templates/favicon.svg"
var faviconSVG []byte

// faviconDataURI は <link rel="icon"> に置く data URI。
// 生成 HTML を 1 ファイルで完結させたいので、外部参照にせず埋め込む。
//
//nolint:gosec // G203: 同梱の自前 SVG を base64 にしただけで外部入力は混ざらない
var faviconDataURI = template.URL(
	"data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString(faviconSVG),
)

var pageTmpl = template.Must(template.New("page").Parse(pageTmplSrc))

// PageOptions はチャート以外のページ側の設定。
type PageOptions struct {
	// Title は <title> の中身。空なら defaultPageTitle。
	Title string
	// Origin はサービスの公開 URL（config の service.url / 環境変数 BASE_URL）。
	// JS 側でページパスからリンクを組み立てるのに使う。__PVVC_DATA__ の JSON に入る。
	Origin string
}

// PageData は __PVVC_DATA__ に書き出すチャートの元データ。
// SVG を直接埋めるのをやめ、この JSON を JS 側で描画する。
type PageData struct {
	Range PageRange `json:"range"`
	Days  []PageDay `json:"days"`
	// Title / Origin は RenderPage が PageOptions から詰める。JS はここを読む。
	// Title はページ上部の見出し（<title> タグと同じ値）。
	Title  string `json:"title,omitempty"`
	Origin string `json:"origin,omitempty"`
}

// PageRange は対象期間（両端とも含む）。
type PageRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// PageDay は 1 日ぶんのデータ。
type PageDay struct {
	Date     string        `json:"date"`
	Cost     float64       `json:"cost"`
	PV       float64       `json:"pv"`
	TopPages []PageTopPath `json:"topPages"`
}

// PageTopPath はその日の上位ページ 1 件。
type PageTopPath struct {
	Path  string `json:"path"`
	Views int64  `json:"views"`
}

type pageTmplData struct {
	Title string
	// Favicon は data URI。html/template は data: を素の文字列だと
	// #ZgotmplZ に潰すので template.URL で渡す。
	Favicon template.URL
	// Data は __PVVC_DATA__ に流し込む JSON。encoding/json が <>& を
	// エスケープ済みなので、そのまま script 要素に置いてもブレイクアウトしない。
	Data template.JS
	// Script は pvvc-chart.js を <script type="module"> にそのまま展開したもの。
	Script template.JS
}

// RenderPage は元データを __PVVC_DATA__ に JSON で埋めた単体 HTML を書き出す。
// 描画は同梱の pvvc-chart.js が JSON を読んで行う。
func RenderPage(w io.Writer, data PageData, page PageOptions) error {
	data.Title = page.Title // サービス名のみ（空なら JS 側がデフォルト表示する）
	data.Origin = page.Origin
	if page.Title == "" {
		page.Title = defaultPageTitle
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("chart: %w", err)
	}

	tmplData := pageTmplData{
		Title:   page.Title,
		Favicon: faviconDataURI,
		Data:    template.JS(raw),         //nolint:gosec // G203: json.Marshal 済みで <>& はエスケープ済み
		Script:  template.JS(pvvcChartJS), //nolint:gosec // G203: 同梱の自前スクリプトを意図的にインライン展開

	}

	// 途中で失敗したときに壊れた HTML を書き残さないよう、いったん溜める。
	var out bytes.Buffer
	out.Grow(len(raw) + len(pageTmplSrc) + len(pvvcChartJS) + len(faviconDataURI))
	if err = pageTmpl.Execute(&out, tmplData); err != nil {
		return fmt.Errorf("chart: %w", err)
	}

	_, err = w.Write(out.Bytes())
	return err
}
