package slack

import (
	"reflect"
	"testing"

	rep "github.com/4okimi7uki/pvvc/internal/report"
)

func TestTurncate(t *testing.T) {
	got := truncate("hello world", 5)
	want := "hello"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestToLinkedRowst(t *testing.T) {
	got := toLinkedRows(formattedSmampleTopPages, "https://example.com")
	want := []rep.Row{
		{Label: "<https://example.com/|/>", Value: "5,000"},
		{Label: "<https://example.com/blog|/blog>", Value: "2,500"},
		{Label: "<https://example.com/about|/about>", Value: "1,200"},
		{Label: "<https://example.com/contact|/contact>", Value: "800"},
		{Label: "<https://example.com/pricing|/pricing>", Value: "600"},
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestChunkRowBySize(t *testing.T) {
	got := chunkRowBySize(formattedSmampleTopPages, 100)
	want := [][]rep.Row{
		{
			{Label: "/", Value: "5,000"},
			{Label: "/blog", Value: "2,500"},
			{Label: "/about", Value: "1,200"},
			{Label: "/contact", Value: "800"},
		},
		{
			{Label: "/pricing", Value: "600"},
		},
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("got %q, want %q", got, want)
	}
}
