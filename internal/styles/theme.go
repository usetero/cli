package styles

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
)

// --- Layer 1: Palette (raw color scales) ---
// Already defined in palette.go and colors.go.
// Palette holds ColorFamily maps keyed by Shade.

// --- Layer 2: Tokens (semantic names, resolved from palette + dark/light) ---

// Tokens holds all semantic color assignments for a color mode.
// This is the full vocabulary of colors available to the app.
// Built once from a Palette, never mutated.
type Tokens struct {
	IsDark bool

	// Backgrounds
	Bg         color.Color // Default page/app background
	BgElevated color.Color // Raised surfaces (blocks, cards, toasts)

	// Text
	Text       color.Color // Primary text
	TextMuted  color.Color // Secondary text (keys, labels)
	TextSubtle color.Color // Tertiary text (descriptions, hints)

	// Brand
	GradientStart color.Color // Brand gradient start
	GradientEnd   color.Color // Brand gradient end
	Accent        color.Color // Interactive: links, selections, prompts
	AccentAlt     color.Color // Alternate accent: focus rings, highlights

	// Selection
	SelectionBg color.Color
	SelectionFg color.Color

	// Borders
	Border color.Color // Dividers, separators

	// Input
	InputBg          color.Color
	InputText        color.Color
	InputPlaceholder color.Color
	InputBorder      color.Color
	InputBorderFocus color.Color

	// Status — two-layer contrast system:
	//   Error/Success/Warning/Info = the status color, high contrast against Bg.
	//     Use for text, icons, dots placed directly on the page background.
	//   OnError/OnSuccess/OnWarning/OnInfo = text on a status-colored surface.
	//     Use for text inside badges, chips, toasts whose background is the status color.
	Error     color.Color
	OnError   color.Color
	Success   color.Color
	OnSuccess color.Color
	Warning   color.Color
	OnWarning color.Color
	Info      color.Color
	OnInfo    color.Color
}

// buildTokens resolves semantic tokens from a palette and color mode.
func buildTokens(p Palette, isDark bool) Tokens {
	if isDark {
		return Tokens{
			IsDark: true,

			Bg:         MustHex(p.Neutral[S950]),
			BgElevated: MustHex(p.Neutral[S900]),

			Text:       MustHex(p.Neutral[S200]),
			TextMuted:  MustHex(p.Neutral[S400]),
			TextSubtle: MustHex(p.Neutral[S500]),

			GradientStart: MustHex(p.Brand[S300]),
			GradientEnd:   MustHex(p.Accent[S600]),
			Accent:        MustHex(p.Brand[S300]),
			AccentAlt:     MustHex(p.Accent[S600]),

			SelectionBg: MustHex(p.Brand[S600]),
			SelectionFg: MustHex(p.Neutral[S50]),

			Border: MustHex(p.Neutral[S600]),

			InputBg:          MustHex(p.Neutral[S800]),
			InputText:        MustHex(p.Neutral[S50]),
			InputPlaceholder: MustHex(p.Neutral[S400]),
			InputBorder:      MustHex(p.Neutral[S500]),
			InputBorderFocus: MustHex(p.Brand[S300]),

			Error:     MustHex(p.Error[S400]),
			OnError:   MustHex(p.Error[S950]),
			Success:   MustHex(p.Success[S400]),
			OnSuccess: MustHex(p.Success[S950]),
			Warning:   MustHex(p.Warning[S400]),
			OnWarning: MustHex(p.Warning[S950]),
			Info:      MustHex(p.Info[S400]),
			OnInfo:    MustHex(p.Info[S950]),
		}
	}

	return Tokens{
		IsDark: false,

		Bg:         MustHex(p.Neutral[S50]),
		BgElevated: MustHex(p.Neutral[S100]),

		Text:       MustHex(p.Neutral[S900]),
		TextMuted:  MustHex(p.Neutral[S600]),
		TextSubtle: MustHex(p.Neutral[S500]),

		GradientStart: MustHex(p.Brand[S600]),
		GradientEnd:   MustHex(p.Accent[S800]),
		Accent:        MustHex(p.Brand[S600]),
		AccentAlt:     MustHex(p.Accent[S600]),

		SelectionBg: MustHex(p.Brand[S500]),
		SelectionFg: MustHex(p.Neutral[S950]),

		Border: MustHex(p.Neutral[S300]),

		InputBg:          MustHex(White),
		InputText:        MustHex(p.Neutral[S900]),
		InputPlaceholder: MustHex(p.Neutral[S400]),
		InputBorder:      MustHex(p.Neutral[S300]),
		InputBorderFocus: MustHex(p.Brand[S600]),

		Error:     MustHex(p.Error[S600]),
		OnError:   MustHex(p.Error[S50]),
		Success:   MustHex(p.Success[S600]),
		OnSuccess: MustHex(p.Success[S50]),
		Warning:   MustHex(p.Warning[S600]),
		OnWarning: MustHex(p.Warning[S50]),
		Info:      MustHex(p.Info[S600]),
		OnInfo:    MustHex(p.Info[S50]),
	}
}

// --- Layer 3: Theme (active context, flows through the component tree) ---

// Theme is the active color context passed to every component.
// Components use theme.Bg, theme.Text, etc. without knowing what surface they're on.
// Use WithBg at surface boundaries to change the background for children.
type Theme struct {
	// Active context — what components render with
	Bg         color.Color
	Text       color.Color
	TextMuted  color.Color
	TextSubtle color.Color
	Accent     color.Color
	AccentAlt  color.Color

	// Selection
	SelectionBg color.Color
	SelectionFg color.Color

	// Border
	Border color.Color

	// Input
	InputBg          color.Color
	InputText        color.Color
	InputPlaceholder color.Color
	InputBorder      color.Color
	InputBorderFocus color.Color

	// Status — two-layer contrast system:
	//   Error/Success/Warning/Info = the status color, high contrast against Bg.
	//     Use for text, icons, dots placed directly on the page background.
	//   OnError/OnSuccess/OnWarning/OnInfo = text on a status-colored surface.
	//     Use for text inside badges, chips, toasts whose background is the status color.
	Error     color.Color
	OnError   color.Color
	Success   color.Color
	OnSuccess color.Color
	Warning   color.Color
	OnWarning color.Color
	Info      color.Color
	OnInfo    color.Color

	// Brand (gradient rendering)
	GradientStart color.Color
	GradientEnd   color.Color

	// Available backgrounds (not active — options for WithBg)
	BgElevated color.Color

	// Pre-built styles
	Styles Styles
}

// WithBg returns a copy of the theme with Bg changed.
// Use at surface boundaries so children render on the new background.
func (t Theme) WithBg(bg color.Color) Theme {
	t.Bg = bg
	t.Styles = buildStyles(t)
	return t
}

// Styles holds pre-built lipgloss styles derived from the theme.
type Styles struct {
	Title   lipgloss.Style // Accent + Bold
	Body    lipgloss.Style // Text
	Help    lipgloss.Style // TextMuted
	Action  lipgloss.Style // Accent
	URL     lipgloss.Style // TextMuted
	Success lipgloss.Style // Success + Bold
	Error   lipgloss.Style // Error
}

// DetectTheme creates a theme by detecting the terminal's background color.
// For standalone CLI usage only — Bubble Tea apps should use NewTheme with
// tea.BackgroundColorMsg instead.
func DetectTheme() Theme {
	return NewTheme(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
}

// NewTheme creates a new theme from the default palette. Use isDark=true for dark mode.
func NewTheme(isDark bool) Theme {
	tokens := buildTokens(DefaultPalette(), isDark)
	return themeFromTokens(tokens)
}

// themeFromTokens creates a Theme from resolved tokens.
func themeFromTokens(t Tokens) Theme {
	theme := Theme{
		Bg:         t.Bg,
		Text:       t.Text,
		TextMuted:  t.TextMuted,
		TextSubtle: t.TextSubtle,
		Accent:     t.Accent,
		AccentAlt:  t.AccentAlt,

		SelectionBg: t.SelectionBg,
		SelectionFg: t.SelectionFg,

		Border: t.Border,

		InputBg:          t.InputBg,
		InputText:        t.InputText,
		InputPlaceholder: t.InputPlaceholder,
		InputBorder:      t.InputBorder,
		InputBorderFocus: t.InputBorderFocus,

		Error:     t.Error,
		OnError:   t.OnError,
		Success:   t.Success,
		OnSuccess: t.OnSuccess,
		Warning:   t.Warning,
		OnWarning: t.OnWarning,
		Info:      t.Info,
		OnInfo:    t.OnInfo,

		GradientStart: t.GradientStart,
		GradientEnd:   t.GradientEnd,

		BgElevated: t.BgElevated,
	}
	theme.Styles = buildStyles(theme)
	return theme
}

// buildStyles creates pre-built styles from the active theme.
// Styles set foreground only — background is inherited from the terminal (CLI)
// or set explicitly by TUI components at surface boundaries.
func buildStyles(t Theme) Styles {
	return Styles{
		Title:   lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		Body:    lipgloss.NewStyle().Foreground(t.Text),
		Help:    lipgloss.NewStyle().Foreground(t.TextMuted),
		Action:  lipgloss.NewStyle().Foreground(t.Accent),
		URL:     lipgloss.NewStyle().Foreground(t.TextMuted),
		Success: lipgloss.NewStyle().Foreground(t.Success).Bold(true),
		Error:   lipgloss.NewStyle().Foreground(t.Error),
	}
}
