package decimalfmt

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestDecimalCommaf(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		digits int32
		want   string
	}{
		// 整数
		{name: "zero", input: "0", digits: 0, want: "0"},
		{name: "small int", input: "123", digits: 0, want: "123"},
		{name: "1000", input: "1000", digits: 0, want: "1,000"},
		{name: "1234567", input: "1234567", digits: 0, want: "1,234,567"},

		// 小数
		{name: "decimal no rounding", input: "1234.56", digits: 2, want: "1,234.56"},
		{name: "decimal rounded up", input: "1234.567", digits: 2, want: "1,234.57"},
		{name: "decimal rounded down", input: "1234.561", digits: 2, want: "1,234.56"},
		{name: "digits=0 truncates decimal", input: "1234.9", digits: 0, want: "1,235"},

		// 負数
		{name: "negative int", input: "-1000", digits: 0, want: "-1,000"},
		{name: "negative decimal", input: "-1234.56", digits: 2, want: "-1,234.56"},
		{name: "negative decimal2", input: "-1234.567891011121314", digits: 2, want: "-1,234.57"},
		{name: "negative small", input: "-999", digits: 0, want: "-999"},

		// 桁境界
		{name: "999", input: "999", digits: 0, want: "999"},
		{name: "1000 boundary", input: "1000", digits: 0, want: "1,000"},
		{name: "999999", input: "999999", digits: 0, want: "999,999"},
		{name: "1000000", input: "1000000", digits: 0, want: "1,000,000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := decimal.NewFromString(tt.input)
			if err != nil {
				t.Fatalf("invalid input %q: %v", tt.input, err)
			}
			got := DecimalCommaf(d, tt.digits)
			if got != tt.want {
				t.Errorf("DecimalCommaf(%q, %d) = %q, want %q", tt.input, tt.digits, got, tt.want)
			}
		})
	}
}

func TestAddCommas(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0", "0"},
		{"12", "12"},
		{"123", "123"},
		{"1234", "1,234"},
		{"12345", "12,345"},
		{"123456", "123,456"},
		{"1234567", "1,234,567"},
		{"12345678", "12,345,678"},
		{"123456789", "123,456,789"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := addCommas(tt.input)
			if got != tt.want {
				t.Errorf("addCommas(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
