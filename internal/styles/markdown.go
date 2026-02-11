package styles

import (
	"fmt"
	"image/color"
	"sync"

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

// cachedRenderer holds a cached glamour renderer for a specific width and bg.
type cachedRenderer struct {
	renderer *glamour.TermRenderer
	width    int
	bgHex    string
}

var (
	rendererCache *cachedRenderer
	rendererMu    sync.Mutex
)

// getRenderer returns a cached glamour renderer, creating one if the width or bg changed.
// Must be called with rendererMu held.
func getRenderer(theme Theme, width int) *glamour.TermRenderer {
	bgHex := colorToHex(theme.Bg)
	if rendererCache != nil && rendererCache.width == width && rendererCache.bgHex == bgHex {
		return rendererCache.renderer
	}
	r, _ := glamour.NewTermRenderer(
		glamour.WithStyles(markdownStyle(theme)),
		glamour.WithWordWrap(width),
	)
	rendererCache = &cachedRenderer{renderer: r, width: width, bgHex: bgHex}
	return r
}

// RenderMarkdown renders markdown text with theme styling.
func RenderMarkdown(theme Theme, text string, width int) string {
	rendererMu.Lock()
	r := getRenderer(theme, width)
	out, err := r.Render(text)
	rendererMu.Unlock()

	if err != nil {
		return text // fallback to plain text
	}
	return out
}

// markdownStyle builds a glamour StyleConfig from the active theme.
func markdownStyle(t Theme) ansi.StyleConfig {
	text := stringPtr(colorToHex(t.Text))
	muted := stringPtr(colorToHex(t.TextMuted))
	accent := stringPtr(colorToHex(t.Accent))
	codeBg := stringPtr(colorToHex(t.BgElevated))

	bg := stringPtr(colorToHex(t.Bg))

	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:           text,
				BackgroundColor: bg,
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
			CrossedOut: boolPtr(false),
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
				// Base
				Text:  ansi.StylePrimitive{Color: text},
				Error: ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Error))},

				// Comments
				Comment:        ansi.StylePrimitive{Color: muted, Italic: boolPtr(true)},
				CommentPreproc: ansi.StylePrimitive{Color: muted, Italic: boolPtr(true)},

				// Keywords (if/else/for, true/false/null, import, int/string, class/struct)
				Keyword:          ansi.StylePrimitive{Color: accent},
				KeywordReserved:  ansi.StylePrimitive{Color: accent},
				KeywordNamespace: ansi.StylePrimitive{Color: accent},
				KeywordType:      ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Info))},

				// Operators & punctuation ({}, :, =, +)
				Operator:    ansi.StylePrimitive{Color: text},
				Punctuation: ansi.StylePrimitive{Color: muted},

				// Names (variables, functions, classes, JSON keys, decorators)
				Name:          ansi.StylePrimitive{Color: text},
				NameBuiltin:   ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Info))},
				NameTag:       ansi.StylePrimitive{Color: accent},
				NameAttribute: ansi.StylePrimitive{Color: accent},
				NameClass:     ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Info)), Bold: boolPtr(true)},
				NameConstant:  ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Warning))},
				NameDecorator: ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Info))},
				NameException: ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Error))},
				NameFunction:  ansi.StylePrimitive{Color: accent},
				NameOther:     ansi.StylePrimitive{Color: text},

				// Literals (strings, numbers, dates)
				Literal:             ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Success))},
				LiteralNumber:       ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Warning))},
				LiteralDate:         ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Warning))},
				LiteralString:       ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Success))},
				LiteralStringEscape: ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Warning))},

				// Diff
				GenericDeleted:    ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Error))},
				GenericEmph:       ansi.StylePrimitive{Italic: boolPtr(true)},
				GenericInserted:   ansi.StylePrimitive{Color: stringPtr(colorToHex(t.Success))},
				GenericStrong:     ansi.StylePrimitive{Bold: boolPtr(true)},
				GenericSubheading: ansi.StylePrimitive{Color: muted},
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
