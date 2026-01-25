package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Surface defines colors for a background surface.
// Text and TextMuted are guaranteed to have good contrast with Bg.
type Surface struct {
	Bg        color.Color
	Text      color.Color
	TextMuted color.Color
}

// InputSurface defines colors for input elements.
type InputSurface struct {
	Bg          color.Color
	Text        color.Color
	Placeholder color.Color
	Border      color.Color
	BorderFocus color.Color
}

// BrandColors defines brand gradient colors.
type BrandColors struct {
	GradientStart color.Color // Lime - start of gradient
	GradientEnd   color.Color // Emerald - end of gradient
}

// StatusColors defines colors for a status type.
type StatusColors struct {
	Fg color.Color // Foreground (text/icon)
	Bg color.Color // Background
}

// Colors holds all semantic color tokens, organized by surface/context.
type Colors struct {
	IsDark bool

	// Surfaces - colors grouped by where they appear
	Page  Surface      // Main page background
	Panel Surface      // Cards, sidebar, footer, modals
	Input InputSurface // Input fields

	// Brand
	Brand  BrandColors // Logo gradient, header diagonals
	Accent color.Color // Interactive: links, selections, focus rings

	// Borders
	BorderDefault color.Color // Dividers, separators

	// Status
	Error   StatusColors
	Success StatusColors
	Warning StatusColors
}

// Styles holds pre-built lipgloss styles derived from colors.
type Styles struct {
	Title    lipgloss.Style // Page/step titles (Accent + Bold)
	Subtitle lipgloss.Style // Subtitle/explanatory text (TextMuted)
	Body     lipgloss.Style // Main text content (Text)
	Help     lipgloss.Style // Secondary help text (TextMuted)
	Action   lipgloss.Style // User action prompts (Accent)
	URL      lipgloss.Style // URL displays (TextMuted)
	Success  lipgloss.Style // Success messages (Success + Bold)
	Error    lipgloss.Style // Error messages (Error)
}

// Theme holds colors and pre-built styles. Created once at startup.
type Theme struct {
	Colors *Colors
	Styles *Styles
}

// NewTheme creates a new theme. Use isDark=true for dark mode.
func NewTheme(isDark bool) *Theme {
	colors := buildColors(DefaultPalette(), isDark)
	return &Theme{
		Colors: colors,
		Styles: buildStyles(colors),
	}
}

// buildColors constructs colors from a palette
func buildColors(p Palette, isDark bool) *Colors {
	if isDark {
		return &Colors{
			IsDark: true,

			Page: Surface{
				Bg:        MustHex(p.Neutral[S900]),
				Text:      MustHex(p.Neutral[S50]),
				TextMuted: MustHex(p.Neutral[S300]),
			},
			Panel: Surface{
				Bg:        MustHex(p.Neutral[S800]),
				Text:      MustHex(p.Neutral[S50]),
				TextMuted: MustHex(p.Neutral[S300]),
			},
			Input: InputSurface{
				Bg:          MustHex(p.Neutral[S800]),
				Text:        MustHex(p.Neutral[S50]),
				Placeholder: MustHex(p.Neutral[S400]),
				Border:      MustHex(p.Neutral[S500]),
				BorderFocus: MustHex(p.Brand[S300]),
			},

			Brand: BrandColors{
				GradientStart: MustHex(p.Accent[S300]),
				GradientEnd:   MustHex(p.Brand[S300]),
			},
			Accent: MustHex(p.Brand[S300]),

			BorderDefault: MustHex(p.Neutral[S600]),

			Error: StatusColors{
				Fg: MustHex(p.Error[S400]),
				Bg: MustHex(p.Error[S800]),
			},
			Success: StatusColors{
				Fg: MustHex(p.Success[S400]),
				Bg: MustHex(p.Success[S800]),
			},
			Warning: StatusColors{
				Fg: MustHex(p.Warning[S400]),
				Bg: MustHex(p.Warning[S800]),
			},
		}
	}

	// Light theme - flip the shade numbers
	return &Colors{
		IsDark: false,

		Page: Surface{
			Bg:        MustHex(p.Neutral[S50]),
			Text:      MustHex(p.Neutral[S900]),
			TextMuted: MustHex(p.Neutral[S600]),
		},
		Panel: Surface{
			Bg:        MustHex(p.Neutral[S100]),
			Text:      MustHex(p.Neutral[S900]),
			TextMuted: MustHex(p.Neutral[S600]),
		},
		Input: InputSurface{
			Bg:          MustHex(White),
			Text:        MustHex(p.Neutral[S900]),
			Placeholder: MustHex(p.Neutral[S400]),
			Border:      MustHex(p.Neutral[S300]),
			BorderFocus: MustHex(p.Brand[S600]),
		},

		Brand: BrandColors{
			GradientStart: MustHex(p.Accent[S600]),
			GradientEnd:   MustHex(p.Brand[S600]),
		},
		Accent: MustHex(p.Brand[S600]),

		BorderDefault: MustHex(p.Neutral[S300]),

		Error: StatusColors{
			Fg: MustHex(p.Error[S600]),
			Bg: MustHex(p.Error[S100]),
		},
		Success: StatusColors{
			Fg: MustHex(p.Success[S600]),
			Bg: MustHex(p.Success[S100]),
		},
		Warning: StatusColors{
			Fg: MustHex(p.Warning[S600]),
			Bg: MustHex(p.Warning[S100]),
		},
	}
}

// buildStyles creates pre-built styles from colors
func buildStyles(c *Colors) *Styles {
	return &Styles{
		Title:    lipgloss.NewStyle().Foreground(c.Accent).Bold(true),
		Subtitle: lipgloss.NewStyle().Foreground(c.Page.TextMuted),
		Body:     lipgloss.NewStyle().Foreground(c.Page.Text),
		Help:     lipgloss.NewStyle().Foreground(c.Page.TextMuted),
		Action:   lipgloss.NewStyle().Foreground(c.Accent),
		URL:      lipgloss.NewStyle().Foreground(c.Page.TextMuted),
		Success:  lipgloss.NewStyle().Foreground(c.Success.Fg).Bold(true),
		Error:    lipgloss.NewStyle().Foreground(c.Error.Fg),
	}
}
