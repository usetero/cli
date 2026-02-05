package styles

import (
	"fmt"
	"image/color"

	"charm.land/glamour/v2"
	"charm.land/glamour/v2/ansi"
)

// Helper functions for style pointers
func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }
func uintPtr(u uint) *uint       { return &u }

// colorToHex converts a color.Color to hex string for glamour
func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

// cachedRenderer holds a cached glamour renderer for a specific width.
type cachedRenderer struct {
	renderer *glamour.TermRenderer
	width    int
}

var rendererCache *cachedRenderer

// MarkdownRenderer returns a glamour renderer configured with theme colors.
// The renderer is cached and reused for the same width.
func MarkdownRenderer(theme *Theme, width int) *glamour.TermRenderer {
	if rendererCache != nil && rendererCache.width == width {
		return rendererCache.renderer
	}
	r, _ := glamour.NewTermRenderer(
		glamour.WithStyles(markdownStyle(theme.Colors)),
		glamour.WithWordWrap(width),
	)
	rendererCache = &cachedRenderer{renderer: r, width: width}
	return r
}

// RenderMarkdown renders markdown text with theme styling.
func RenderMarkdown(theme *Theme, text string, width int) string {
	r := MarkdownRenderer(theme, width)
	out, err := r.Render(text)
	if err != nil {
		return text // fallback to plain text
	}
	return out
}

// markdownStyle builds a glamour StyleConfig from our theme colors.
func markdownStyle(c *Colors) ansi.StyleConfig {
	text := stringPtr(colorToHex(c.Page.Text))
	muted := stringPtr(colorToHex(c.Page.TextMuted))
	accent := stringPtr(colorToHex(c.Accent))
	codeBg := stringPtr(colorToHex(c.Panel.Bg))

	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: text,
			},
			Margin: uintPtr(0),
		},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  muted,
				Italic: boolPtr(true),
			},
			Indent:      uintPtr(1),
			IndentToken: stringPtr("│ "),
		},
		List: ansi.StyleList{
			LevelIndent: 2,
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BlockSuffix: "\n",
				Color:       accent,
				Bold:        boolPtr(true),
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "# ",
				Color:  accent,
				Bold:   boolPtr(true),
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "## ",
				Color:  accent,
				Bold:   boolPtr(true),
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "### ",
				Color:  accent,
				Bold:   boolPtr(true),
			},
		},
		H4: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "#### ",
				Color:  accent,
			},
		},
		H5: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "##### ",
				Color:  accent,
			},
		},
		H6: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Prefix: "###### ",
				Color:  accent,
			},
		},
		Strikethrough: ansi.StylePrimitive{
			CrossedOut: boolPtr(true),
		},
		Emph: ansi.StylePrimitive{
			Color:  text,
			Italic: boolPtr(true),
		},
		Strong: ansi.StylePrimitive{
			Color: text,
			Bold:  boolPtr(true),
		},
		HorizontalRule: ansi.StylePrimitive{
			Color:  muted,
			Format: "\n────────\n",
		},
		Item: ansi.StylePrimitive{
			BlockPrefix: "• ",
		},
		Enumeration: ansi.StylePrimitive{
			BlockPrefix: ". ",
		},
		Task: ansi.StyleTask{
			Ticked:   "[✓] ",
			Unticked: "[ ] ",
		},
		Link: ansi.StylePrimitive{
			Color:     accent,
			Underline: boolPtr(true),
		},
		LinkText: ansi.StylePrimitive{
			Color: accent,
			Bold:  boolPtr(true),
		},
		Image: ansi.StylePrimitive{
			Color:     muted,
			Underline: boolPtr(true),
		},
		ImageText: ansi.StylePrimitive{
			Color:  muted,
			Format: "Image: {{.text}} →",
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:           accent,
				BackgroundColor: codeBg,
				Prefix:          " ",
				Suffix:          " ",
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{
					Color: text,
				},
				Margin: uintPtr(0),
			},
			Chroma: &ansi.Chroma{
				Text: ansi.StylePrimitive{
					Color: text,
				},
				Keyword: ansi.StylePrimitive{
					Color: accent,
				},
				Name: ansi.StylePrimitive{
					Color: text,
				},
				NameFunction: ansi.StylePrimitive{
					Color: accent,
				},
				LiteralString: ansi.StylePrimitive{
					Color: stringPtr(colorToHex(c.Success.Fg)),
				},
				LiteralNumber: ansi.StylePrimitive{
					Color: stringPtr(colorToHex(c.Warning.Fg)),
				},
				Comment: ansi.StylePrimitive{
					Color:  muted,
					Italic: boolPtr(true),
				},
			},
		},
		Table: ansi.StyleTable{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{},
			},
			CenterSeparator: stringPtr("┼"),
			ColumnSeparator: stringPtr("│"),
			RowSeparator:    stringPtr("─"),
		},
		DefinitionDescription: ansi.StylePrimitive{
			BlockPrefix: "\n  ",
		},
	}
}
