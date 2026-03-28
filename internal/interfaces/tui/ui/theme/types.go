package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

const AppName = "Tero"

// Palette is the source design-token set for a theme.
type Palette struct {
	Surface       color.Color
	Border        color.Color
	Text          color.Color
	TextMuted     color.Color
	TextSubtle    color.Color
	Accent        color.Color
	AccentAlt     color.Color
	GradientStart color.Color
	GradientEnd   color.Color
	Success       color.Color
	Warning       color.Color
	Error         color.Color
}

// Theme defines the active render context for one TUI subtree.
type Theme struct {
	Background color.Color
	Palette    Palette
	Gradients  GradientStyles

	Shell    ShellStyles
	Card     CardStyles
	Text     TextStyles
	List     ListStyles
	Input    InputStyles
	Progress ProgressStyles
}

// WithBackground returns a copy of the theme whose styles render against the
// provided background color. Use this at surface boundaries instead of letting
// children guess their background.
func (t Theme) WithBackground(background color.Color) Theme {
	t.Background = background
	t = applyStyles(t)
	return t
}

// OnSurface returns a copy of the theme that renders against the surface background.
func (t Theme) OnSurface() Theme {
	return t.WithBackground(t.Palette.Surface)
}

// GradientStyles are semantic gradients used across shared chrome and components.
type GradientStyles struct {
	Brand Gradient
	Motif Gradient
}

// ShellStyles are the app-wide chrome styles.
type ShellStyles struct {
	Outer       lipgloss.Style
	HeaderBar   lipgloss.Style
	HeaderBrand lipgloss.Style
	HeaderLead  lipgloss.Style
	HeaderBadge lipgloss.Style
	HeaderRule  lipgloss.Style
	Body        lipgloss.Style
	Footer      lipgloss.Style
	FooterLead  lipgloss.Style
	FooterRule  lipgloss.Style
}

// CardStyles are reusable card styles for emphasized content blocks.
type CardStyles struct {
	Container      lipgloss.Style
	ErrorContainer lipgloss.Style
	Title          lipgloss.Style
	ErrorTitle     lipgloss.Style
	Body           lipgloss.Style
}

// TextStyles are shared semantic text styles.
type TextStyles struct {
	Section lipgloss.Style
	Body    lipgloss.Style
	Muted   lipgloss.Style
	Subtle  lipgloss.Style
	Error   lipgloss.Style
	Success lipgloss.Style
	Warning lipgloss.Style
}

// ListStyles are shared styles for selectable lists.
type ListStyles struct {
	Container       lipgloss.Style
	ActiveContainer lipgloss.Style
	Cursor          lipgloss.Style
	CursorInactive  lipgloss.Style
	Item            lipgloss.Style
	ItemActive      lipgloss.Style
	Subtitle        lipgloss.Style
	SubtitleActive  lipgloss.Style
	Empty           lipgloss.Style
}

// InputStyles are shared styles for text inputs.
type InputStyles struct {
	Container       lipgloss.Style
	ActiveContainer lipgloss.Style
	Label           lipgloss.Style
	Value           lipgloss.Style
	Placeholder     lipgloss.Style
	Active          lipgloss.Style
	Inactive        lipgloss.Style
}

// ProgressStyles are shared styles for progress bars.
type ProgressStyles struct {
	Fill  lipgloss.Style
	Empty lipgloss.Style
}
