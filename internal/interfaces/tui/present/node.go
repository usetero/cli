package present

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/theme"
)

// TextRole is a semantic text style.
type TextRole uint8

const (
	RoleTitle TextRole = iota
	RoleBody
	RoleMuted
	RoleSubtle
	RoleError
	RoleSuccess
	RoleWarning
	RoleLabel
)

// SurfaceKind is a semantic content surface.
type SurfaceKind uint8

const (
	SurfaceCard SurfaceKind = iota
	SurfaceErrorCard
)

// Context is the current render context.
type Context struct {
	Theme theme.Theme
}

// Node is one presentation node.
type Node interface {
	render(Context) string
}

// Block is structured multiline content that can be safely rendered inside a
// surface.
type Block interface {
	Node
	surfaceLines(Context) []surfaceLine
}

// BlockItem is one line-oriented item within a surface block.
type BlockItem interface {
	surfaceLines(Context) []surfaceLine
}

type textNode struct {
	role  TextRole
	value string
}

type rawNode struct {
	value string
}

type stackNode struct {
	gap      int
	children []Node
}

type rowNode struct {
	children []Node
}

type surfaceNode struct {
	kind SurfaceKind
	body Block
}

type blockNode struct {
	gap   int
	items []BlockItem
}

type surfaceLine struct {
	hasRole bool
	role    TextRole
	text    string
}

// Render compiles a node into terminal output.
func Render(appTheme theme.Theme, node Node) string {
	if node == nil {
		return ""
	}
	return node.render(Context{Theme: appTheme})
}

// View compiles a node into a tea.View.
func View(appTheme theme.Theme, node Node) tea.View {
	return tea.NewView(Render(appTheme, node))
}

// Title renders section/title text.
func Title(value string) textNode { return textNode{role: RoleTitle, value: strings.TrimSpace(value)} }

// Body renders primary body text.
func Body(value string) textNode { return textNode{role: RoleBody, value: value} }

// Muted renders muted supporting text.
func Muted(value string) textNode { return textNode{role: RoleMuted, value: value} }

// Subtle renders tertiary supporting text.
func Subtle(value string) textNode { return textNode{role: RoleSubtle, value: value} }

// Error renders error text.
func Error(value string) textNode { return textNode{role: RoleError, value: value} }

// Success renders success text.
func Success(value string) textNode { return textNode{role: RoleSuccess, value: value} }

// Warning renders warning text.
func Warning(value string) textNode { return textNode{role: RoleWarning, value: value} }

// Label renders field labels.
func Label(value string) textNode { return textNode{role: RoleLabel, value: value} }

// Raw embeds already-rendered content from components.
func Raw(value string) rawNode { return rawNode{value: value} }

// Stack vertically stacks nodes with no gap.
func Stack(children ...Node) Node { return StackGap(0, children...) }

// StackGap vertically stacks nodes with a line gap between each child.
func StackGap(gap int, children ...Node) Node {
	filtered := make([]Node, 0, len(children))
	for _, child := range children {
		if child != nil {
			filtered = append(filtered, child)
		}
	}
	return stackNode{gap: max(0, gap), children: filtered}
}

// Row horizontally joins child nodes.
func Row(children ...Node) Node {
	filtered := make([]Node, 0, len(children))
	for _, child := range children {
		if child != nil {
			filtered = append(filtered, child)
		}
	}
	return rowNode{children: filtered}
}

// BlockOf builds structured multiline block content with no extra gap.
func BlockOf(items ...BlockItem) Block { return BlockGap(0, items...) }

// BlockGap builds structured multiline block content with a line gap between items.
func BlockGap(gap int, items ...BlockItem) Block {
	filtered := make([]BlockItem, 0, len(items))
	for _, item := range items {
		if item != nil {
			filtered = append(filtered, item)
		}
	}
	return blockNode{gap: max(0, gap), items: filtered}
}

// Card renders a raised surface around structured content.
func Card(body Block) Node { return surfaceNode{kind: SurfaceCard, body: body} }

// ErrorCard renders an error-toned raised surface.
func ErrorCard(body Block) Node { return surfaceNode{kind: SurfaceErrorCard, body: body} }

// Notice renders a standard non-interactive notice.
func Notice(title string, body string) Node {
	if strings.TrimSpace(body) == "" {
		return BlockOf(Title(title))
	}
	return BlockGap(1, Title(title), Muted(body))
}

// Section renders a standard step section.
func Section(title string, body Node) Node {
	return StackGap(1, Title(title), body)
}

// Field renders one labeled field row.
func Field(prefix string, label string, value Node) Node {
	return Row(Raw(prefix), Label(label), value)
}

// StatusBlock renders a card-backed status block.
func StatusBlock(title string, parts ...BlockItem) Node {
	children := []BlockItem{Title(title)}
	for _, part := range parts {
		if part == nil {
			continue
		}
		children = append(children, part)
	}
	return Card(BlockGap(1, children...))
}

func (n textNode) render(ctx Context) string {
	style := roleStyle(ctx.Theme, n.role)
	return style.Render(n.value)
}

func (n rawNode) render(Context) string { return n.value }

func (n stackNode) render(ctx Context) string {
	if len(n.children) == 0 {
		return ""
	}
	lines := make([]string, 0, len(n.children)*2)
	for i, child := range n.children {
		if i > 0 && n.gap > 0 {
			for range n.gap {
				lines = append(lines, "")
			}
		}
		lines = append(lines, child.render(ctx))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (n rowNode) render(ctx Context) string {
	parts := make([]string, 0, len(n.children))
	for _, child := range n.children {
		parts = append(parts, child.render(ctx))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

func (n surfaceNode) render(ctx Context) string {
	surfaceTheme := ctx.Theme.WithBackground(ctx.Theme.Surface)
	content := renderSurfaceContent(Context{Theme: surfaceTheme}, n.body)

	switch n.kind {
	case SurfaceErrorCard:
		return ctx.Theme.Card.ErrorContainer.Render(content)
	default:
		return ctx.Theme.Card.Container.Render(content)
	}
}

func (n blockNode) render(ctx Context) string {
	lines := n.surfaceLines(ctx)
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		if line.hasRole {
			rendered = append(rendered, roleStyle(ctx.Theme, line.role).Render(line.text))
			continue
		}
		rendered = append(rendered, line.text)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}

func renderSurfaceContent(ctx Context, body Block) string {
	lines := body.surfaceLines(ctx)
	width := 1
	for _, line := range lines {
		if w := lipgloss.Width(line.text); w > width {
			width = w
		}
	}

	rendered := make([]string, 0, len(lines))
	whitespace := lipgloss.NewStyle().Background(ctx.Theme.Background).ColorWhitespace(true)
	for _, line := range lines {
		if line.hasRole {
			rendered = append(rendered, roleStyle(ctx.Theme, line.role).Width(width).ColorWhitespace(true).Render(line.text))
			continue
		}
		rendered = append(rendered, lipgloss.PlaceHorizontal(
			width,
			lipgloss.Left,
			line.text,
			lipgloss.WithWhitespaceStyle(whitespace),
		))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}

func roleStyle(t theme.Theme, role TextRole) lipgloss.Style {
	switch role {
	case RoleBody:
		return t.Text.Body
	case RoleMuted:
		return t.Text.Muted
	case RoleSubtle:
		return t.Text.Subtle
	case RoleError:
		return t.Text.Error
	case RoleSuccess:
		return t.Text.Success
	case RoleWarning:
		return t.Text.Warning
	case RoleLabel:
		return t.Input.Label
	default:
		return t.Text.Section
	}
}

func fillBackground(content string, background color.Color) string {
	lines := strings.Split(content, "\n")
	width := 1
	for _, line := range lines {
		if w := lipgloss.Width(line); w > width {
			width = w
		}
	}
	fill := lipgloss.NewStyle().Background(background).ColorWhitespace(true)
	for i := range lines {
		lines[i] = lipgloss.PlaceHorizontal(
			width,
			lipgloss.Left,
			lines[i],
			lipgloss.WithWhitespaceStyle(fill),
		)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (n textNode) surfaceLines(Context) []surfaceLine {
	return splitSurfaceText(n.role, n.value)
}

func (n rawNode) surfaceLines(Context) []surfaceLine {
	return splitSurfaceRaw(n.value)
}

func (n blockNode) surfaceLines(ctx Context) []surfaceLine {
	lines := make([]surfaceLine, 0, len(n.items)*2)
	for i, item := range n.items {
		if i > 0 && n.gap > 0 {
			for range n.gap {
				lines = append(lines, surfaceLine{text: ""})
			}
		}
		lines = append(lines, item.surfaceLines(ctx)...)
	}
	return lines
}

func splitSurfaceText(role TextRole, value string) []surfaceLine {
	lines := strings.Split(value, "\n")
	out := make([]surfaceLine, 0, len(lines))
	for _, line := range lines {
		out = append(out, surfaceLine{
			hasRole: true,
			role:    role,
			text:    line,
		})
	}
	return out
}

func splitSurfaceRaw(value string) []surfaceLine {
	lines := strings.Split(value, "\n")
	out := make([]surfaceLine, 0, len(lines))
	for _, line := range lines {
		out = append(out, surfaceLine{text: line})
	}
	return out
}
