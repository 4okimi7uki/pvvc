package slack

import (
	"fmt"
	"strings"

	rep "github.com/4okimi7uki/pvvc/internal/report"
)

func toLinkedRows(rows []rep.Row, baseURL string) []rep.Row {
	var result []rep.Row
	for _, row := range rows {
		linked := fmt.Sprintf("<%s%s|%s>", baseURL, row.Label, row.Label)
		result = append(result, rep.Row{Label: linked, Value: row.Value})
	}
	return result
}

// <url|text> 形式のSlackリンクは表示テキスト(|以降)の長さを返す
func slackDisplayLen(s string) int {
	if len(s) > 0 && s[0] == '<' {
		if i := strings.Index(s, "|"); i != -1 {
			return len(s) - i - 2 // '|text>' の text 部分
		}
	}
	return len(s)
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}

func chunkRowBySize(rows []rep.Row, maxLen int) [][]rep.Row {
	var chunks [][]rep.Row
	var current []rep.Row

	for _, row := range rows {
		candidate := append(current, row)
		var sb strings.Builder
		rep.WriteTableFn(&sb, rep.RowsToCells(candidate), rep.StrLen)
		if len(sb.String()) > maxLen && len(candidate) > 0 {
			chunks = append(chunks, current)
			current = []rep.Row{row}
		} else {
			current = candidate
		}
	}

	if len(current) > 0 {
		chunks = append(chunks, current)
	}

	return chunks
}
