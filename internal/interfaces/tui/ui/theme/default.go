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
	brand := pick(lightBrand, darkBrand)
	accent := pick(lightAccent, darkAccent)
	errorText := pick(lightErrorText, darkErrorText)
	warningText := pick(lightWarningText, darkWarningText)
	success := pick(lightSuccess, darkSuccess)
	return applyStyles(Theme{
		Background: background,
		Palette: Palette{
			Surface:       surface,
			Border:        border,
			Text:          text,
			TextMuted:     muted,
			TextSubtle:    subtle,
			Brand:         brand,
			Accent:        accent,
			GradientStart: brand,
			GradientEnd:   accent,
			Success:       success,
			Warning:       warningText,
			Error:         errorText,
		},
		Gradients: GradientStyles{
			Brand: Gradient{Start: brand, End: accent},
			Motif: Gradient{Start: accent, End: brand},
		},
	})
}

func applyStyles(t Theme) Theme {
	bg := t.Background

	t.Shell = ShellStyles{
		Outer: lipgloss.NewStyle().
			Background(bg).
			Padding(0, 1, 1, 1),
		HeaderBar: lipgloss.NewStyle().
			Background(bg).
			Padding(1, 0, 0, 0),
		HeaderBrand: lipgloss.NewStyle().
			Foreground(t.Palette.Brand).
			Background(bg).
			Bold(true),
		HeaderLead: lipgloss.NewStyle().
			Foreground(t.Palette.TextMuted).
			Background(bg),
		HeaderBadge: lipgloss.NewStyle().
			Foreground(t.Palette.Warning).
			Background(t.Palette.Surface).
			Padding(0, 1),
		HeaderRule: lipgloss.NewStyle().
			Foreground(t.Palette.TextSubtle).
			Background(bg),
		Body: lipgloss.NewStyle().
			Background(bg).
			Padding(1, 0, 0, 0),
		Footer: lipgloss.NewStyle().
			Background(bg).
			Foreground(t.Palette.TextMuted).
			Padding(1, 0, 1, 0),
		FooterLead: lipgloss.NewStyle().
			Foreground(t.Palette.TextSubtle).
			Background(bg),
		FooterRule: lipgloss.NewStyle().
			Foreground(t.Palette.Border).
			Background(bg),
	}

	t.Card = CardStyles{
		Container: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Palette.Border).
			BorderBackground(t.Palette.Surface).
			Background(t.Palette.Surface).
			Padding(1, 2),
		ErrorContainer: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Palette.Error).
			BorderBackground(t.Palette.Surface).
			Background(t.Palette.Surface).
			Padding(1, 2),
		Title: lipgloss.NewStyle().
			Foreground(t.Palette.Brand).
			Background(t.Palette.Surface).
			Bold(true),
		ErrorTitle: lipgloss.NewStyle().
			Foreground(t.Palette.Error).
			Background(t.Palette.Surface).
			Bold(true),
		Body: lipgloss.NewStyle().
			Foreground(t.Palette.Text).
			Background(t.Palette.Surface),
	}

	t.Text = TextStyles{
		Section: lipgloss.NewStyle().
			Foreground(t.Palette.Brand).
			Background(bg).
			Bold(true),
		Body: lipgloss.NewStyle().
			Foreground(t.Palette.Text).
			Background(bg),
		Muted: lipgloss.NewStyle().
			Foreground(t.Palette.TextMuted).
			Background(bg),
		Subtle: lipgloss.NewStyle().
			Foreground(t.Palette.TextSubtle).
			Background(bg),
		Error: lipgloss.NewStyle().
			Foreground(t.Palette.Error).
			Background(bg).
			Bold(true),
		Success: lipgloss.NewStyle().
			Foreground(t.Palette.Success).
			Background(bg).
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(t.Palette.Warning).
			Background(bg).
			Bold(true),
	}

	t.List = ListStyles{
		Container: lipgloss.NewStyle().
			Background(bg).
			BorderStyle(lipgloss.Border{Left: "│"}).
			BorderForeground(t.Palette.Border).
			BorderBackground(bg).
			PaddingLeft(1),
		ActiveContainer: lipgloss.NewStyle().
			Background(bg).
			BorderStyle(lipgloss.Border{Left: "┃"}).
			BorderForeground(t.Palette.Brand).
			BorderBackground(bg).
			PaddingLeft(1),
		Cursor: lipgloss.NewStyle().
			Foreground(t.Palette.Brand).
			Background(bg).
			Bold(true),
		CursorInactive: lipgloss.NewStyle().
			Foreground(t.Palette.TextMuted).
			Background(bg),
		Item: lipgloss.NewStyle().
			Foreground(t.Palette.Text).
			Background(bg),
		ItemActive: lipgloss.NewStyle().
			Foreground(t.Palette.Brand).
			Background(bg).
			Bold(true),
		Subtitle: lipgloss.NewStyle().
			Foreground(t.Palette.TextMuted).
			Background(bg),
		SubtitleActive: lipgloss.NewStyle().
			Foreground(t.Palette.Accent).
			Background(bg),
		Empty: lipgloss.NewStyle().
			Foreground(t.Palette.TextMuted).
			Background(bg),
	}

	t.Input = InputStyles{
		Container: lipgloss.NewStyle().
			Background(bg).
			BorderStyle(lipgloss.Border{Left: "│"}).
			BorderForeground(t.Palette.Border).
			BorderBackground(bg).
			PaddingLeft(1),
		ActiveContainer: lipgloss.NewStyle().
			Background(bg).
			BorderStyle(lipgloss.Border{Left: "┃"}).
			BorderForeground(t.Palette.Brand).
			BorderBackground(bg).
			PaddingLeft(1),
		Label: lipgloss.NewStyle().
			Foreground(t.Palette.TextMuted).
			Background(bg).
			Bold(true),
		Value: lipgloss.NewStyle().
			Foreground(t.Palette.Text).
			Background(bg),
		Placeholder: lipgloss.NewStyle().
			Foreground(t.Palette.TextMuted).
			Background(bg),
		Active: lipgloss.NewStyle().
			Foreground(t.Palette.Brand).
			Background(bg).
			Bold(true),
		Inactive: lipgloss.NewStyle().
			Foreground(t.Palette.TextMuted).
			Background(bg),
	}

	t.Progress = ProgressStyles{
		Fill: lipgloss.NewStyle().
			Foreground(t.Palette.Accent).
			Background(bg).
			Bold(true),
		Empty: lipgloss.NewStyle().
			Foreground(t.Palette.TextMuted).
			Background(bg),
	}

	return t
}
