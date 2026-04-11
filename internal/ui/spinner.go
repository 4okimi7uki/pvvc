package ui

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

func padSuffix(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

func WithSpinner(initialMsg string, fn func(update func(string), addDone func(string)) error) error {
	const spinnerSuffixWidth = 30
	const interval = 140 * time.Millisecond
	mySpinner := []string{Lime("˙"), "•", Lime("●"), Lime("•"), "˙"}
	s := spinner.New(mySpinner, interval)
	s.Writer = os.Stderr
	s.Suffix = padSuffix(" "+initialMsg, spinnerSuffixWidth)

	s.Start()
	defer func() {
		time.Sleep(600 * time.Millisecond)
		s.Stop()
	}()

	var mu sync.Mutex
	var doneItems []string

	addDone := func(label string) {
		mu.Lock()
		doneItems = append(doneItems, label)
		suffix := " " + initialMsg + "\n" + strings.Join(doneItems, "\n")
		mu.Unlock()

		s.Lock()
		s.Suffix = suffix
		s.Unlock()
	}

	update := func(msg string) {
		s.Lock()
		s.Suffix = padSuffix(" "+msg, spinnerSuffixWidth)
		s.Unlock()
	}

	return fn(update, addDone)
}
