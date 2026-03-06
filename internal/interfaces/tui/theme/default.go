package theme

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
)

// Default returns the shared default theme for the current terminal.
func Default() Theme {
	return New(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
}

// New returns a theme for the requested terminal background mode.
func New(isDark bool) Theme {
	lightDark := lipgloss.LightDark(isDark)
	pick := func(lightHex string, darkHex string) color.Color {
		return lightDark(lipgloss.Color(lightHex), lipgloss.Color(darkHex))
	}

	surface := pick(lightSurface, darkSurface)
	border := pick(lightBorder, darkBorder)
	text := pick(lightText, darkText)
	muted := pick(lightMutedText, darkMutedText)
	accent := pick(lightAccent, darkAccent)
	errorText := pick(lightErrorText, darkErrorText)
	warningText := pick(lightWarningText, darkWarningText)
	success := pick(lightSuccess, darkSuccess)
	progressEmpty := pick(lightProgressEmpty, darkProgressEmpty)

	return Theme{
		Shell: ShellStyles{
			Outer: lipgloss.NewStyle().Padding(0, 2, 0, 2),
			HeaderBar: lipgloss.NewStyle().
				Padding(1, 0, 0, 0),
			HeaderBrand: lipgloss.NewStyle().
				Foreground(accent).
				Bold(true),
			HeaderLead: lipgloss.NewStyle().
				Foreground(muted),
			Body: lipgloss.NewStyle().
				Padding(1, 0, 0, 0),
			Footer: lipgloss.NewStyle().
				Foreground(muted).
				Padding(0, 0, 1, 0),
		},
		Card: CardStyles{
			Container: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(border).
				Background(surface).
				Padding(1, 2),
			ErrorContainer: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(errorText).
				Background(surface).
				Padding(1, 2),
			Title: lipgloss.NewStyle().
				Foreground(accent).
				Bold(true),
			ErrorTitle: lipgloss.NewStyle().
				Foreground(errorText).
				Bold(true),
			Body: lipgloss.NewStyle().Foreground(text),
		},
		Text: TextStyles{
			Section: lipgloss.NewStyle().
				Foreground(accent).
				Bold(true),
			Body: lipgloss.NewStyle().Foreground(text),
			Muted: lipgloss.NewStyle().
				Foreground(muted),
			Error: lipgloss.NewStyle().
				Foreground(errorText).
				Bold(true),
			Success: lipgloss.NewStyle().
				Foreground(success).
				Bold(true),
			Warning: lipgloss.NewStyle().
				Foreground(warningText).
				Bold(true),
		},
		List: ListStyles{
			Cursor: lipgloss.NewStyle().
				Foreground(accent).
				Bold(true),
			CursorInactive: lipgloss.NewStyle().
				Foreground(muted),
			Item: lipgloss.NewStyle().Foreground(text),
			ItemActive: lipgloss.NewStyle().
				Foreground(accent).
				Bold(true),
			Subtitle: lipgloss.NewStyle().Foreground(muted),
			SubtitleActive: lipgloss.NewStyle().
				Foreground(accent),
			Empty: lipgloss.NewStyle().Foreground(muted),
		},
		Input: InputStyles{
			Label: lipgloss.NewStyle().
				Foreground(muted).
				Bold(true),
			Value:       lipgloss.NewStyle().Foreground(text),
			Placeholder: lipgloss.NewStyle().Foreground(muted),
			Active: lipgloss.NewStyle().
				Foreground(accent).
				Bold(true),
			Inactive: lipgloss.NewStyle().Foreground(muted),
		},
		Progress: ProgressStyles{
			Fill:  lipgloss.NewStyle().Foreground(success).Bold(true),
			Empty: lipgloss.NewStyle().Foreground(progressEmpty),
		},
	}
}
