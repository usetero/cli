package theme

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/cli/config"
)

// Default returns the shared default theme for the current terminal.
func Default() Theme {
	return Resolve(config.ThemeModeAuto)
}

// Resolve returns a theme for the requested mode.
func Resolve(mode config.ThemeMode) Theme {
	switch mode {
	case config.ThemeModeLight:
		return New(false)
	case config.ThemeModeDark:
		return New(true)
	default:
		return New(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
	}
}

// New returns a theme for the requested terminal background mode.
func New(isDark bool) Theme {
	lightDark := lipgloss.LightDark(isDark)
	pick := func(lightHex string, darkHex string) color.Color {
		return lightDark(lipgloss.Color(lightHex), lipgloss.Color(darkHex))
	}

	background := pick(lightBackground, darkBackground)
	surface := pick(lightSurface, darkSurface)
	border := pick(lightBorder, darkBorder)
	text := pick(lightText, darkText)
	muted := pick(lightMutedText, darkMutedText)
	subtle := pick(lightSubtleText, darkSubtleText)
	accent := pick(lightAccent, darkAccent)
	accentAlt := pick(lightAccentAlt, darkAccentAlt)
	errorText := pick(lightErrorText, darkErrorText)
	warningText := pick(lightWarningText, darkWarningText)
	success := pick(lightSuccess, darkSuccess)
	return applyStyles(Theme{
		Background:    background,
		Surface:       surface,
		Border:        border,
		TextColor:     text,
		TextMuted:     muted,
		TextSubtle:    subtle,
		Accent:        accent,
		AccentAlt:     accentAlt,
		GradientStart: accent,
		GradientEnd:   accentAlt,
		Success:       success,
		Warning:       warningText,
		Error:         errorText,
		Gradients: GradientStyles{
			Brand: Gradient{Start: accent, End: accentAlt},
			Motif: Gradient{Start: accentAlt, End: accent},
		},
	})
}

func applyStyles(t Theme) Theme {
	bg := t.Background

	t.Shell = ShellStyles{
		Outer: lipgloss.NewStyle().
			Background(bg).
			Padding(0, 2, 0, 2),
		HeaderBar: lipgloss.NewStyle().
			Background(bg).
			Padding(1, 0, 0, 0),
		HeaderBrand: lipgloss.NewStyle().
			Foreground(t.Accent).
			Background(bg).
			Bold(true),
		HeaderLead: lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Background(bg),
		HeaderBadge: lipgloss.NewStyle().
			Foreground(t.Warning).
			Background(t.Surface).
			Padding(0, 1),
		HeaderRule: lipgloss.NewStyle().
			Foreground(t.TextSubtle).
			Background(bg),
		Body: lipgloss.NewStyle().
			Background(bg).
			Padding(1, 0, 0, 0),
		Footer: lipgloss.NewStyle().
			Background(bg).
			Foreground(t.TextMuted).
			Padding(1, 0, 1, 0),
		FooterLead: lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Background(bg),
		FooterRule: lipgloss.NewStyle().
			Foreground(t.TextSubtle).
			Background(bg),
	}

	t.Card = CardStyles{
		Container: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Border).
			BorderBackground(t.Surface).
			Background(t.Surface).
			Padding(1, 2),
		ErrorContainer: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Error).
			BorderBackground(t.Surface).
			Background(t.Surface).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(t.Accent).
			Background(t.Surface).
			Bold(true),
		ErrorTitle: lipgloss.NewStyle().
			Foreground(t.Error).
			Background(t.Surface).
			Bold(true),
		Body: lipgloss.NewStyle().
			Foreground(t.TextColor).
			Background(t.Surface),
	}

	t.Text = TextStyles{
		Section: lipgloss.NewStyle().
			Foreground(t.Accent).
			Background(bg).
			Bold(true),
		Body: lipgloss.NewStyle().
			Foreground(t.TextColor).
			Background(bg),
		Muted: lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Background(bg),
		Subtle: lipgloss.NewStyle().
			Foreground(t.TextSubtle).
			Background(bg),
		Error: lipgloss.NewStyle().
			Foreground(t.Error).
			Background(bg).
			Bold(true),
		Success: lipgloss.NewStyle().
			Foreground(t.Success).
			Background(bg).
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(t.Warning).
			Background(bg).
			Bold(true),
	}

	t.List = ListStyles{
		Container: lipgloss.NewStyle().
			Background(bg).
			BorderStyle(lipgloss.Border{Left: "│"}).
			BorderForeground(t.Border).
			BorderBackground(bg).
			PaddingLeft(1),
		ActiveContainer: lipgloss.NewStyle().
			Background(bg).
			BorderStyle(lipgloss.Border{Left: "┃"}).
			BorderForeground(t.AccentAlt).
			BorderBackground(bg).
			PaddingLeft(1),
		Cursor: lipgloss.NewStyle().
			Foreground(t.Accent).
			Background(bg).
			Bold(true),
		CursorInactive: lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Background(bg),
		Item: lipgloss.NewStyle().
			Foreground(t.TextColor).
			Background(bg),
		ItemActive: lipgloss.NewStyle().
			Foreground(t.Accent).
			Background(bg).
			Bold(true),
		Subtitle: lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Background(bg),
		SubtitleActive: lipgloss.NewStyle().
			Foreground(t.AccentAlt).
			Background(bg),
		Empty: lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Background(bg),
	}

	t.Input = InputStyles{
		Container: lipgloss.NewStyle().
			Background(bg).
			BorderStyle(lipgloss.Border{Left: "│"}).
			BorderForeground(t.Border).
			BorderBackground(bg).
			PaddingLeft(1),
		ActiveContainer: lipgloss.NewStyle().
			Background(bg).
			BorderStyle(lipgloss.Border{Left: "┃"}).
			BorderForeground(t.AccentAlt).
			BorderBackground(bg).
			PaddingLeft(1),
		Label: lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Background(bg).
			Bold(true),
		Value: lipgloss.NewStyle().
			Foreground(t.TextColor).
			Background(bg),
		Placeholder: lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Background(bg),
		Active: lipgloss.NewStyle().
			Foreground(t.AccentAlt).
			Background(bg).
			Bold(true),
		Inactive: lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Background(bg),
	}

	t.Progress = ProgressStyles{
		Fill: lipgloss.NewStyle().
			Foreground(t.AccentAlt).
			Background(bg).
			Bold(true),
		Empty: lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Background(bg),
	}

	return t
}
