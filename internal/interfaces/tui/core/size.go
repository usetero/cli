package core

// HeightProvider exposes a model's preferred rendered height for a given width.
type HeightProvider interface {
	PreferredHeight(width int) int
}
