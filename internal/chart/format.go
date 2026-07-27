package chart

import (
	"math"
	"strconv"
	"strings"
)

// 軸の目盛りラベル用のフォーマッタ。Series.FormatTick に渡して使う。
// いずれも func(float64) string なので、足したいものがあれば同じ形で書けばよい。

// Plain は小数点以下を落とした素の数値。FormatTick 未指定時の既定。
func Plain(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// USD は "$130" 形式。
func USD(v float64) string { return "$" + strconv.FormatFloat(v, 'f', -1, 64) }

// Comma は "859,241" 形式。
// TODO: internal/decimalfmt に同等の実装があるなら寄せる。
func Comma(v float64) string {
	s := strconv.FormatFloat(math.Round(v), 'f', 0, 64)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")

	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	if neg {
		return "-" + string(out)
	}
	return string(out)
}

// Compact は "200k" / "1M" 形式。
func Compact(v float64) string {
	switch {
	case v >= 1_000_000:
		return Plain(v/1_000_000) + "M"
	case v >= 1_000:
		return Plain(v/1_000) + "k"
	default:
		return Plain(v)
	}
}
