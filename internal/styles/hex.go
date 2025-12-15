package styles

import (
	"image/color"

	"github.com/lucasb-eyer/go-colorful"
)

// MustHex parses a hex color and panics if invalid
func MustHex(hex string) color.Color {
	c, err := colorful.Hex(hex)
	if err != nil {
		panic("invalid hex color: " + hex)
	}
	return c
}

// ColorToHex converts color.Color to hex string
func ColorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return "#" +
		toHex(uint8(r>>8)) +
		toHex(uint8(g>>8)) +
		toHex(uint8(b>>8))
}

func toHex(n uint8) string {
	const hex = "0123456789ABCDEF"
	return string([]byte{hex[n>>4], hex[n&0x0f]})
}
