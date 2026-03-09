package theme

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/rivo/uniseg"
)

// Gradient is a semantic foreground gradient that can render grapheme-safe text.
type Gradient struct {
	Start color.Color
	End   color.Color
}

// Render applies the gradient across grapheme clusters in the input string.
func (g Gradient) Render(input string, bold bool) string {
	if input == "" {
		return ""
	}

	clusters := graphemeClusters(input)
	ramp := g.Ramp(len(clusters))

	var out strings.Builder
	for i, cluster := range clusters {
		c, _ := colorful.MakeColor(ramp[i])
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(c.Clamped().Hex()))
		if bold {
			style = style.Bold(true)
		}
		fmt.Fprint(&out, style.Render(cluster))
	}
	return out.String()
}

// Ramp returns a blended color ramp from start to end.
func (g Gradient) Ramp(size int) []color.Color {
	if size <= 0 {
		return nil
	}
	if g.Start == nil || g.End == nil {
		return make([]color.Color, size)
	}

	start, _ := colorful.MakeColor(g.Start)
	end, _ := colorful.MakeColor(g.End)

	colors := make([]color.Color, 0, size)
	for i := range size {
		t := 0.0
		if size > 1 {
			t = float64(i) / float64(size-1)
		}
		colors = append(colors, start.BlendHcl(end, t))
	}
	return colors
}

func graphemeClusters(input string) []string {
	var clusters []string
	graphemes := uniseg.NewGraphemes(input)
	for graphemes.Next() {
		clusters = append(clusters, graphemes.Str())
	}
	return clusters
}
