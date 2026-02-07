package input

import (
	"testing"

	"github.com/usetero/cli/internal/styles"
	"github.com/usetero/cli/internal/tea/teatest"
)

func TestView(t *testing.T) {
	t.Parallel()
	theme := styles.NewTheme(true)

	t.Run("no raw escape sequences when empty", func(t *testing.T) {
		t.Parallel()
		m := New(theme)
		m.SetWidth(40)

		output := m.View()
		t.Logf("output: %q", output)
		teatest.AssertNoRawEscapes(t, output)
	})

	t.Run("no raw escape sequences with value", func(t *testing.T) {
		t.Parallel()
		m := New(theme)
		m.SetWidth(40)
		m.SetValue("hello world")

		output := m.View()
		t.Logf("output: %q", output)
		teatest.AssertNoRawEscapes(t, output)
	})
}
