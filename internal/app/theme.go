package app

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func pvvcTheme() *huh.Theme {
	t := huh.ThemeBase()

	var (
		yellow      = lipgloss.AdaptiveColor{Light: "#D4A017", Dark: "#F5C542"}
		yellowLight = lipgloss.AdaptiveColor{Light: "#F5C542", Dark: "#FFE08A"}
		amber       = lipgloss.AdaptiveColor{Light: "#B8860B", Dark: "#FFA500"}
		normalFg    = lipgloss.AdaptiveColor{Light: "235", Dark: "252"}
		subtleFg    = lipgloss.AdaptiveColor{Light: "243", Dark: "243"}
		cream       = lipgloss.AdaptiveColor{Light: "#FFFDF5", Dark: "#FFFDF5"}
		red         = lipgloss.AdaptiveColor{Light: "#FF4672", Dark: "#ED567A"}
	)

	t.Focused.Base = t.Focused.Base.BorderForeground(yellow)
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(yellow).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(yellow).Bold(true).MarginBottom(1)
	t.Focused.Description = t.Focused.Description.Foreground(subtleFg)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(red)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(red)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(amber)
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(amber)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(amber)
	t.Focused.Option = t.Focused.Option.Foreground(normalFg)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(amber)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(yellowLight)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(yellowLight).SetString("✓ ")
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(subtleFg).SetString("• ")
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(normalFg)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(cream).Background(amber)
	t.Focused.Next = t.Focused.FocusedButton
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(normalFg).Background(lipgloss.AdaptiveColor{Light: "252", Dark: "237"})
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(yellow)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(subtleFg)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(amber)

	t.Blurred = t.Focused
	t.Blurred.Base = t.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description

	return t
}
