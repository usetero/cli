package spinner

import "image/color"

const (
	// DefaultSize is the default number of cycling characters.
	DefaultSize = 10

	// DefaultFPS is the animation frame rate.
	DefaultFPS = 20

	// EllipsisAnimSpeed is how many frames between ellipsis changes.
	// At 20 FPS, 8 frames = 400ms per ellipsis state.
	EllipsisAnimSpeed = 8

	// MaxBirthOffset is the maximum stagger delay for character appearance.
	MaxBirthOffset = 1_000_000_000 // 1 second in nanoseconds

	// PrerenderedFrames is how many frames to pre-render when not cycling colors.
	PrerenderedFrames = 10
)

// Default colors for gradient.
var (
	DefaultGradColorA = color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}
	DefaultGradColorB = color.RGBA{R: 0, G: 0, B: 0xff, A: 0xff}
	DefaultLabelColor = color.RGBA{R: 0xcc, G: 0xcc, B: 0xcc, A: 0xff}
)

// Settings configures the spinner animation.
type Settings struct {
	// Size is the number of cycling characters.
	Size int

	// Label is the text shown after the cycling characters.
	Label string

	// LabelColor is the color of the label and ellipsis.
	LabelColor color.Color

	// GradColorA is the start color of the gradient.
	GradColorA color.Color

	// GradColorB is the end color of the gradient.
	GradColorB color.Color

	// CycleColors enables continuous color cycling through the gradient.
	CycleColors bool
}

// normalize applies default values to unset fields.
func (s *Settings) normalize() {
	if s.Size < 1 {
		s.Size = DefaultSize
	}
	if colorIsUnset(s.GradColorA) {
		s.GradColorA = DefaultGradColorA
	}
	if colorIsUnset(s.GradColorB) {
		s.GradColorB = DefaultGradColorB
	}
	if colorIsUnset(s.LabelColor) {
		s.LabelColor = DefaultLabelColor
	}
}

func colorIsUnset(c color.Color) bool {
	if c == nil {
		return true
	}
	_, _, _, a := c.RGBA()
	return a == 0
}
