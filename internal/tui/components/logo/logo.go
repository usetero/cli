// Package logo renders the Tero wordmark.
package logo

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/usetero/cli/internal/tui/styles"
)

// Opts are the options for rendering the Tero logo.
type Opts struct {
	TitleColorA color.Color // left gradient ramp point (start)
	TitleColorB color.Color // right gradient ramp point (end)
}

// Render renders just the TERO wordmark (using Crush-style letterforms).
// Layout components (header, sidebar) add diagonal fields as needed.
func Render(o Opts) string {
	// TERO using Crush-style compact letterforms (3 lines tall)
	teroArt := `▀▀▀▀▀ █▀▀▀▀ █▀▀▀▄ ▄▀▀▀▄
  █   █▀▀▀▀ █▀▀▀▄ █   █
  ▀   ▀▀▀▀▀ ▀   ▀  ▀▀▀ `

	// Apply gradient to each line of the logo
	b := new(strings.Builder)
	for _, line := range strings.Split(teroArt, "\n") {
		fmt.Fprintln(b, styles.ApplyForegroundGrad(line, o.TitleColorA, o.TitleColorB))
	}

	return strings.TrimSpace(b.String())
}
