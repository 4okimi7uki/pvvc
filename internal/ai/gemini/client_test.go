package gemini

import (
	"fmt"
	"testing"

	"google.golang.org/api/googleapi"
)

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "503 returns true",
			err:  &googleapi.Error{Code: 503},
			want: true,
		},
		{
			name: "429 returns true",
			err:  &googleapi.Error{Code: 429},
			want: true,
		},
		{
			name: "500 returns false",
			err:  &googleapi.Error{Code: 500},
			want: false,
		},
		{
			name: "wrapped 503 returns true",
			err:  fmt.Errorf("after 3 attempts: %w", &googleapi.Error{Code: 503}),
			want: true,
		},
		{
			name: "non-api error returns false",
			err:  fmt.Errorf("some other error"),
			want: false,
		},
		{
			name: "nil returns false",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRateLimitError(tt.err)
			if got != tt.want {
				t.Errorf("isRateLimitError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
