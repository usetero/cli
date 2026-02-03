// Package logo renders the Tero wordmark.
package logo

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/usetero/cli/internal/styles"
)

// Opts are the options for rendering the Tero logo.
type Opts struct {
	TitleColorA color.Color // left gradient ramp point (start)
	TitleColorB color.Color // right gradient ramp point (end)
}

// Render renders just the TERO wordmark (using Crush-style letterforms).
func Render(o Opts) string {
	teroArt := `▀▀▀▀▀ █▀▀▀▀ █▀▀▀▄ ▄▀▀▀▄
  █   █▀▀▀▀ █▀▀▀▄ █   █
  ▀   ▀▀▀▀▀ ▀   ▀  ▀▀▀ `

	b := new(strings.Builder)
	for _, line := range strings.Split(teroArt, "\n") {
		fmt.Fprintln(b, styles.ApplyForegroundGrad(line, o.TitleColorA, o.TitleColorB))
	}

	return strings.TrimSpace(b.String())
}
