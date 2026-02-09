package cmd

import "github.com/usetero/cli/internal/styles"

const kvKeyWidth = 22

// kv renders a key-value pair with ANSI-aware alignment.
// Key is styled as Help (muted), value as Body (primary text).
func kv(s styles.Styles, key, value string) string {
	return s.Help.Width(kvKeyWidth).Render(key+":") + " " + s.Body.Render(value)
}

// kvStyled renders a key-value pair where the value is already styled.
func kvStyled(s styles.Styles, key, value string) string {
	return s.Help.Width(kvKeyWidth).Render(key+":") + " " + value
}

// section renders a section header with a preceding blank line.
func section(s styles.Styles, title string) string {
	return "\n" + s.Title.Render(title)
}
