package styles

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/rivo/uniseg"
)

// blendColors creates a gradient between multiple color stops
func blendColors(size int, stops ...color.Color) []color.Color {
	if len(stops) < 2 {
		return nil
	}

	stopsPrime := make([]colorful.Color, len(stops))
	for i, k := range stops {
		stopsPrime[i], _ = colorful.MakeColor(k)
	}

	numSegments := len(stopsPrime) - 1
	blended := make([]color.Color, 0, size)

	// Calculate how many colors each segment should have
	segmentSizes := make([]int, numSegments)
	baseSize := size / numSegments
	remainder := size % numSegments

	// Distribute the remainder across segments
	for i := range numSegments {
		segmentSizes[i] = baseSize
		if i < remainder {
			segmentSizes[i]++
		}
	}

	// Generate colors for each segment
	for i := range numSegments {
		c1 := stopsPrime[i]
		c2 := stopsPrime[i+1]
		segmentSize := segmentSizes[i]

		for j := range segmentSize {
			var t float64
			if segmentSize > 1 {
				t = float64(j) / float64(segmentSize-1)
			}
			c := c1.BlendHcl(c2, t)
			blended = append(blended, c)
		}
	}

	return blended
}

// ForegroundGrad creates gradient text clusters
func ForegroundGrad(input string, bold bool, color1, color2 color.Color) []string {
	if input == "" {
		return []string{""}
	}

	if len(input) == 1 {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorToHex(color1)))
		if bold {
			style = style.Bold(true)
		}
		return []string{style.Render(input)}
	}

	var clusters []string
	gr := uniseg.NewGraphemes(input)
	for gr.Next() {
		clusters = append(clusters, string(gr.Runes()))
	}

	ramp := blendColors(len(clusters), color1, color2)
	for i, c := range ramp {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorToHex(c)))
		if bold {
			style = style.Bold(true)
		}
		clusters[i] = style.Render(clusters[i])
	}
	return clusters
}

// ApplyForegroundGrad renders a string with a horizontal gradient foreground
func ApplyForegroundGrad(input string, color1, color2 color.Color) string {
	if input == "" {
		return ""
	}
	var o strings.Builder
	clusters := ForegroundGrad(input, false, color1, color2)
	for _, c := range clusters {
		fmt.Fprint(&o, c)
	}
	return o.String()
}

// ApplyBoldForegroundGrad renders a string with a bold horizontal gradient foreground
func ApplyBoldForegroundGrad(input string, color1, color2 color.Color) string {
	if input == "" {
		return ""
	}
	var o strings.Builder
	clusters := ForegroundGrad(input, true, color1, color2)
	for _, c := range clusters {
		fmt.Fprint(&o, c)
	}
	return o.String()
}
