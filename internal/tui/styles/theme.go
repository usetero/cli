package styles

import (
	"image/color"
)

// Theme holds all colors and styles for the TUI
type Theme struct {
	IsDark bool

	// Brand colors
	Primary   color.Color // Main brand color (Teal 400/500)
	Secondary color.Color // Secondary brand (Teal 500/600)

	// Text hierarchy
	Text       color.Color // Primary text
	TextMuted  color.Color // Supporting text, less emphasis
	TextSubtle color.Color // Minimal emphasis

	// Surfaces
	Background    color.Color // Main background
	BackgroundAlt color.Color // Panels, footers, cards

	// UI elements
	Border color.Color // Borders, separators, dividers
	Field  color.Color // Diagonal patterns, decorative elements

	// Selection
	Selected           color.Color // Selected item text
	SelectedBackground color.Color // Selected item background

	// Status/Feedback (foreground)
	Error   color.Color // Error text/icons
	Success color.Color // Success text/icons
	Warning color.Color // Warning text/icons
	Info    color.Color // Info text/icons

	// Status/Feedback (backgrounds)
	ErrorBackground   color.Color // Error banner background
	SuccessBackground color.Color // Success banner background
	WarningBackground color.Color // Warning banner background
	InfoBackground    color.Color // Info banner background
}

var currentTheme *Theme

// CurrentTheme returns the current theme (always dark mode like Crush)
func CurrentTheme() *Theme {
	if currentTheme == nil {
		// Always use dark theme like Crush
		currentTheme = getTheme(true)
	}
	return currentTheme
}

func getTheme(isDark bool) *Theme {
	if isDark {
		return &Theme{
			IsDark:             true,
			Primary:            MustHex(Teal300),   // Bright teal
			Secondary:          MustHex(Indigo400), // Cyan
			Text:               MustHex(Zinc50),    // Almost white
			TextMuted:          MustHex(Zinc400),   // Gray
			TextSubtle:         MustHex(Zinc500),   // Darker gray
			Background:         MustHex(Zinc900),   // Very dark
			BackgroundAlt:      MustHex(Zinc800),   // Slightly lighter for panels/footers
			Border:             MustHex(Zinc700),   // Subtle border/separator
			Field:              MustHex(Indigo400), // Diagonal patterns
			Selected:           MustHex(Teal300),   // Selected item text
			SelectedBackground: MustHex(Zinc700),   // Selected item background
			Error:              MustHex(Red500),    // Red
			Success:            MustHex(Green500),  // Green
			Warning:            MustHex(Amber500),  // Amber
			Info:               MustHex(Blue500),   // Blue
			ErrorBackground:    MustHex(Red700),    // Dark red for error banners
			SuccessBackground:  MustHex(Green700),  // Dark green for success banners
			WarningBackground:  MustHex(Amber700),  // Dark amber for warning banners
			InfoBackground:     MustHex(Blue700),   // Dark blue for info banners
		}
	}

	// Light theme
	return &Theme{
		IsDark:             false,
		Primary:            MustHex(Teal300),   // Bright teal
		Secondary:          MustHex(Indigo400), // Cyan
		Text:               MustHex(Zinc900),   // Almost black
		TextMuted:          MustHex(Zinc600),   // Gray
		TextSubtle:         MustHex(Zinc400),   // Lighter gray
		Background:         MustHex(Zinc50),    // Almost white
		BackgroundAlt:      MustHex(Zinc100),   // Slightly darker for panels/footers
		Border:             MustHex(Zinc300),   // Borders
		Field:              MustHex(Indigo400), // Diagonal patterns
		Selected:           MustHex(Teal600),   // Selected item text
		SelectedBackground: MustHex(Teal50),    // Selected item background
		Error:              MustHex(Red600),    // Red
		Success:            MustHex(Green600),  // Green
		Warning:            MustHex(Amber600),  // Amber
		Info:               MustHex(Blue600),   // Blue
		ErrorBackground:    MustHex(Red100),    // Light red for error banners
		SuccessBackground:  MustHex(Green100),  // Light green for success banners
		WarningBackground:  MustHex(Amber100),  // Light amber for warning banners
		InfoBackground:     MustHex(Blue100),   // Light blue for info banners
	}
}
