package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

func padSuffix(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func WithSpinner(initialMsg string, fn func(update func(string)) error) error {
	const spinnerSuffixWidth = 30
	const interval = 140 * time.Millisecond
	mySet := []string{Lime("˙"), "•", Lime("●"), Lime("•"), "˙"}
	s := spinner.New(mySet, interval)
	s.Writer = os.Stderr
	s.Suffix = padSuffix(" "+initialMsg, spinnerSuffixWidth)

	s.Start()
	defer func() {
		fmt.Fprint(os.Stderr, "\r\033[K")
		s.Stop()
	}()

	update := func(msg string) {
		s.Suffix = padSuffix(" "+msg, spinnerSuffixWidth)
		time.Sleep(interval + 10*time.Millisecond)
	}

	return fn(update)
}
