package decimalfmt

import (
	"strings"

	"github.com/shopspring/decimal"
)

func DecimalCommaf(d decimal.Decimal, degits int32) string {
	rounded := d.Round(degits)
	s := rounded.String()

	sign := ""
	if strings.HasPrefix(s, "-") {
		sign = "-"
		s = s[1:]
	}

	parts := strings.Split(s, ".")
	intPart := parts[0]

	intPart = addCommas(intPart)

	if len(parts) == 2 {
		return sign + intPart + "." + parts[1]
	}
	return sign + intPart
}

func addCommas(s string) string {
	const degit = 3
	n := len(s)
	if n < degit {
		return s
	}
	var b strings.Builder
	pre := n % degit
	if pre > 0 {
		b.WriteString(s[:pre])
		if n > pre {
			b.WriteString(",")
		}
	}
	for i := pre; i < n; i += degit {
		b.WriteString(s[i : i+degit])
		if i+degit < n {
			b.WriteString(",")
		}
	}
	return b.String()
}
