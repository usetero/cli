package styles

// Palette defines the base color families for the theme.
// Change these to rebrand the entire application.
type Palette struct {
	Brand   ColorFamily // Primary brand color (Emerald)
	Accent  ColorFamily // Secondary brand for gradients (Lime)
	Neutral ColorFamily // Grays (Neutral)
	Error   ColorFamily // Error states (Red)
	Success ColorFamily // Success states (Emerald)
	Warning ColorFamily // Warning states (Amber)
	Info    ColorFamily // Info states (Cyan)
}

// DefaultPalette returns the Tero brand palette.
// To rebrand, change the color families here.
func DefaultPalette() Palette {
	return Palette{
		Brand:   EmeraldFamily, // Primary brand color
		Accent:  CyanFamily,    // Gradient: Emerald → Cyan
		Neutral: ZincFamily,
		Error:   RedFamily,
		Success: EmeraldFamily,
		Warning: AmberFamily,
		Info:    CyanFamily,
	}
}
