package chart

import (
	"bytes"
	"strings"
	"testing"
)

// SVG が外部参照ではなく本文に埋まっていること。
func TestRenderPageInlinesSVG(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderPage(&buf, fixtureOptions(), PageOptions{}); err != nil {
		t.Fatalf("RenderPage() = %v", err)
	}

	got := buf.String()
	if !strings.HasPrefix(got, "<!doctype html>") {
		t.Errorf("HTML になっていない: %.40s", got)
	}
	if !strings.Contains(got, `<svg xmlns="http://www.w3.org/2000/svg"`) {
		t.Error("SVG がインライン展開されていない")
	}
	if strings.Contains(got, "&lt;svg") {
		t.Error("SVG がエスケープされて文字列として出ている")
	}
	if !strings.Contains(got, "</svg>") || !strings.Contains(got, "</html>") {
		t.Error("SVG または HTML が閉じていない")
	}
}

// 外部ファイルを一切参照しないこと（file:// で開けることの担保）。
func TestRenderPageIsSelfContained(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderPage(&buf, fixtureOptions(), PageOptions{}); err != nil {
		t.Fatalf("RenderPage() = %v", err)
	}

	for _, ng := range []string{"fetch(", "<img", "<link", "<script", "src="} {
		if strings.Contains(buf.String(), ng) {
			t.Errorf("外部参照が含まれている: %q", ng)
		}
	}
}

func TestRenderPageTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"既定", "", defaultPageTitle},
		{"指定", "example.com | PVVC chart", "example.com | PVVC chart"},
		{"エスケープ", "a<b>c", "a&lt;b&gt;c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := RenderPage(&buf, fixtureOptions(), PageOptions{Title: tt.title}); err != nil {
				t.Fatalf("RenderPage() = %v", err)
			}
			if want := "<title>" + tt.want + "</title>"; !strings.Contains(buf.String(), want) {
				t.Errorf("%q が無い", want)
			}
		})
	}
}

// 描画できないときは HTML の断片を書き残さないこと。
func TestRenderPageErrorsOnNoData(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderPage(&buf, Options{}, PageOptions{}); err == nil {
		t.Fatal("データが無いときはエラーを期待したが nil")
	}
	if buf.Len() != 0 {
		t.Errorf("失敗時に %d バイト書かれている: %q", buf.Len(), buf.String())
	}
}
