package theme

import (
	"strings"
	"testing"
)

func TestNew_BuildsRenderableStyles(t *testing.T) {
	light := New(false)
	dark := New(true)

	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "light header bar", value: light.Shell.HeaderBar.Render("TERO")},
		{name: "dark header bar", value: dark.Shell.HeaderBar.Render("TERO")},
		{name: "light header brand", value: light.Shell.HeaderBrand.Render("TERO")},
		{name: "dark header brand", value: dark.Shell.HeaderBrand.Render("TERO")},
		{name: "light body", value: light.Shell.Body.Render("body")},
		{name: "dark body", value: dark.Shell.Body.Render("body")},
		{name: "light card", value: light.Card.Container.Render("body")},
		{name: "dark card", value: dark.Card.Container.Render("body")},
		{name: "light section", value: light.Text.Section.Render("section")},
		{name: "dark section", value: dark.Text.Section.Render("section")},
	} {
		if strings.TrimSpace(tc.value) == "" {
			t.Fatalf("expected %s to render non-empty value", tc.name)
		}
	}
}
