package chart

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func fixturePageData() PageData {
	return PageData{
		Range: PageRange{From: "2026-07-22", To: "2026-07-23"},
		Days: []PageDay{
			{Date: "2026-07-22", Cost: 172.29, PV: 759707, TopPages: []PageTopPath{
				{Path: "/ranking", Views: 1200},
			}},
			{Date: "2026-07-23", Cost: 168.4, PV: 743210},
		},
	}
}

// 元データが __PVVC_DATA__ に JSON として埋まっていること。
func TestRenderPageEmbedsJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := RenderPage(&buf, fixturePageData(), PageOptions{}); err != nil {
		t.Fatalf("RenderPage() = %v", err)
	}

	got := buf.String()
	if !strings.HasPrefix(got, "<!doctype html>") {
		t.Errorf("HTML になっていない: %.40s", got)
	}
	if !strings.Contains(got, `<script id="__PVVC_DATA__" type="application/json">`) {
		t.Error("__PVVC_DATA__ の script が無い")
	}

	// script の中身が JSON として読め、往復できること。
	const open = `<script id="__PVVC_DATA__" type="application/json">`
	_, rest, _ := strings.Cut(got, open)
	body, _, _ := strings.Cut(rest, "</script>")

	var back PageData
	if err := json.Unmarshal([]byte(body), &back); err != nil {
		t.Fatalf("埋め込み JSON が読めない: %v\n%s", err, body)
	}
	if back.Range.From != "2026-07-22" || len(back.Days) != 2 {
		t.Errorf("JSON の内容が違う: %+v", back)
	}
	if len(back.Days[0].TopPages) != 1 || back.Days[0].TopPages[0].Path != "/ranking" {
		t.Errorf("topPages が復元できない: %+v", back.Days[0].TopPages)
	}
}

// JSON 内の </script> や <>& が script をブレイクアウトしないこと。
func TestRenderPageEscapesJSON(t *testing.T) {
	data := PageData{
		Range: PageRange{From: "2026-07-22", To: "2026-07-22"},
		Days: []PageDay{
			{Date: "2026-07-22", Cost: 1, PV: 1, TopPages: []PageTopPath{
				{Path: "</script><b>&", Views: 1},
			}},
		},
	}

	var buf bytes.Buffer
	if err := RenderPage(&buf, data, PageOptions{}); err != nil {
		t.Fatalf("RenderPage() = %v", err)
	}

	// JSON 内の生の </script> が素通しされていないこと（< にエスケープされる）。
	if strings.Contains(buf.String(), "</script><b>&") {
		t.Error("JSON 内の </script> がエスケープされず素通しされている")
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
			if err := RenderPage(&buf, fixturePageData(), PageOptions{Title: tt.title}); err != nil {
				t.Fatalf("RenderPage() = %v", err)
			}
			if want := "<title>" + tt.want + "</title>"; !strings.Contains(buf.String(), want) {
				t.Errorf("%q が無い", want)
			}
		})
	}
}
