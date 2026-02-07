package messagelist

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/usetero/cli/internal/app/chat/messagelist/block"
)

// stubBlock is a minimal block.Block for testing selection math.
type stubBlock struct {
	text    string
	kind    block.Kind
	width   int
	focused bool
}

func newStubBlock(text string, kind block.Kind) *stubBlock {
	return &stubBlock{text: text, kind: kind}
}

func (b *stubBlock) View() string           { return b.text }
func (b *stubBlock) Height() int            { return len(strings.Split(b.text, "\n")) }
func (b *stubBlock) Update(tea.Msg) tea.Cmd { return nil }
func (b *stubBlock) SetWidth(w int)         { b.width = w }
func (b *stubBlock) SetFocused(f bool)      { b.focused = f }
func (b *stubBlock) Focused() bool          { return b.focused }
func (b *stubBlock) Kind() block.Kind       { return b.kind }

func TestGetHighlightRange(t *testing.T) {
	t.Parallel()

	t.Run("no selection", func(t *testing.T) {
		t.Parallel()
		m := &Model{mouseDownBlock: -1}
		sb, _, _, eb, _, _ := m.getHighlightRange()
		if sb != -1 || eb != -1 {
			t.Errorf("expected (-1, -1), got (%d, %d)", sb, eb)
		}
	})

	t.Run("forward drag", func(t *testing.T) {
		t.Parallel()
		m := &Model{
			mouseDownBlock: 0, mouseDownY: 1, mouseDownX: 5,
			mouseDragBlock: 2, mouseDragY: 3, mouseDragX: 10,
		}
		sb, sl, sc, eb, el, ec := m.getHighlightRange()
		if sb != 0 || sl != 1 || sc != 5 || eb != 2 || el != 3 || ec != 10 {
			t.Errorf("forward: got (%d,%d,%d) -> (%d,%d,%d)", sb, sl, sc, eb, el, ec)
		}
	})

	t.Run("backward drag normalizes", func(t *testing.T) {
		t.Parallel()
		m := &Model{
			mouseDownBlock: 2, mouseDownY: 3, mouseDownX: 10,
			mouseDragBlock: 0, mouseDragY: 1, mouseDragX: 5,
		}
		sb, sl, sc, eb, el, ec := m.getHighlightRange()
		if sb != 0 || sl != 1 || sc != 5 || eb != 2 || el != 3 || ec != 10 {
			t.Errorf("backward: got (%d,%d,%d) -> (%d,%d,%d)", sb, sl, sc, eb, el, ec)
		}
	})

	t.Run("same block backward drag normalizes by line", func(t *testing.T) {
		t.Parallel()
		m := &Model{
			mouseDownBlock: 1, mouseDownY: 5, mouseDownX: 3,
			mouseDragBlock: 1, mouseDragY: 2, mouseDragX: 8,
		}
		sb, sl, sc, eb, el, ec := m.getHighlightRange()
		if sb != 1 || sl != 2 || sc != 8 || eb != 1 || el != 5 || ec != 3 {
			t.Errorf("same-block-backward: got (%d,%d,%d) -> (%d,%d,%d)", sb, sl, sc, eb, el, ec)
		}
	})
}

func TestBlockHighlightRange(t *testing.T) {
	t.Parallel()

	// Selection from block 1 (line 2, col 5) to block 3 (line 4, col 10)
	startBlock, startLine, startCol := 1, 2, 5
	endBlock, endLine, endCol := 3, 4, 10

	t.Run("before selection", func(t *testing.T) {
		t.Parallel()
		sl, sc, el, ec := blockHighlightRange(0, startBlock, startLine, startCol, endBlock, endLine, endCol)
		if sl != -1 || sc != -1 || el != -1 || ec != -1 {
			t.Errorf("expected not selected, got (%d,%d,%d,%d)", sl, sc, el, ec)
		}
	})

	t.Run("start block", func(t *testing.T) {
		t.Parallel()
		sl, sc, el, ec := blockHighlightRange(1, startBlock, startLine, startCol, endBlock, endLine, endCol)
		if sl != 2 || sc != 5 || el != -1 || ec != -1 {
			t.Errorf("start: got (%d,%d,%d,%d)", sl, sc, el, ec)
		}
	})

	t.Run("middle block", func(t *testing.T) {
		t.Parallel()
		sl, sc, el, ec := blockHighlightRange(2, startBlock, startLine, startCol, endBlock, endLine, endCol)
		if sl != 0 || sc != 0 || el != -1 || ec != -1 {
			t.Errorf("middle: got (%d,%d,%d,%d)", sl, sc, el, ec)
		}
	})

	t.Run("end block", func(t *testing.T) {
		t.Parallel()
		sl, sc, el, ec := blockHighlightRange(3, startBlock, startLine, startCol, endBlock, endLine, endCol)
		if sl != 0 || sc != 0 || el != 4 || ec != 10 {
			t.Errorf("end: got (%d,%d,%d,%d)", sl, sc, el, ec)
		}
	})

	t.Run("after selection", func(t *testing.T) {
		t.Parallel()
		sl, sc, el, ec := blockHighlightRange(4, startBlock, startLine, startCol, endBlock, endLine, endCol)
		if sl != -1 || sc != -1 || el != -1 || ec != -1 {
			t.Errorf("expected not selected, got (%d,%d,%d,%d)", sl, sc, el, ec)
		}
	})

	t.Run("single block selection", func(t *testing.T) {
		t.Parallel()
		sl, sc, el, ec := blockHighlightRange(2, 2, 1, 3, 2, 5, 8)
		if sl != 1 || sc != 3 || el != 5 || ec != 8 {
			t.Errorf("single: got (%d,%d,%d,%d)", sl, sc, el, ec)
		}
	})
}

func TestExtractHighlight(t *testing.T) {
	t.Parallel()

	t.Run("extracts content without border", func(t *testing.T) {
		t.Parallel()
		m := &Model{
			mouseDownBlock: 0, mouseDownY: 0, mouseDownX: 0,
			mouseDragBlock: 0, mouseDragY: 0, mouseDragX: 5,
		}
		m.blocks = []blockEntry{
			{block: newStubBlock("hello world", block.KindAssistantText)},
		}

		text := m.extractHighlight()
		if !strings.Contains(text, "hello") {
			t.Errorf("expected extracted text to contain 'hello', got %q", text)
		}
		// The key invariant: no border character should appear
		if strings.Contains(text, "│") {
			t.Errorf("extracted text should not contain border character, got %q", text)
		}
	})

	t.Run("no selection returns empty", func(t *testing.T) {
		t.Parallel()
		m := &Model{mouseDownBlock: -1}
		if text := m.extractHighlight(); text != "" {
			t.Errorf("expected empty, got %q", text)
		}
	})

	t.Run("multi-block extraction", func(t *testing.T) {
		t.Parallel()
		m := &Model{
			mouseDownBlock: 0, mouseDownY: 0, mouseDownX: 0,
			mouseDragBlock: 1, mouseDragY: 0, mouseDragX: 3,
		}
		m.blocks = []blockEntry{
			{block: newStubBlock("first", block.KindAssistantText)},
			{block: newStubBlock("second", block.KindAssistantText)},
		}

		text := m.extractHighlight()
		if !strings.Contains(text, "first") || !strings.Contains(text, "sec") {
			t.Errorf("expected both blocks in extraction, got %q", text)
		}
	})
}
