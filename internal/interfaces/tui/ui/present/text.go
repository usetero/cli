package present

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/usetero/cli/internal/interfaces/tui/ui/theme"
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

// Title renders section/title text.
func Title(appTheme theme.Theme, value string) string {
	return roleStyle(appTheme, RoleTitle).Render(strings.TrimSpace(value))
}

// Body renders primary body text.
func Body(appTheme theme.Theme, value string) string {
	return roleStyle(appTheme, RoleBody).Render(value)
}

// Muted renders muted supporting text.
func Muted(appTheme theme.Theme, value string) string {
	return roleStyle(appTheme, RoleMuted).Render(value)
}

// Subtle renders tertiary supporting text.
func Subtle(appTheme theme.Theme, value string) string {
	return roleStyle(appTheme, RoleSubtle).Render(value)
}

// Error renders error text.
func Error(appTheme theme.Theme, value string) string {
	return roleStyle(appTheme, RoleError).Render(value)
}

// Success renders success text.
func Success(appTheme theme.Theme, value string) string {
	return roleStyle(appTheme, RoleSuccess).Render(value)
}

// Warning renders warning text.
func Warning(appTheme theme.Theme, value string) string {
	return roleStyle(appTheme, RoleWarning).Render(value)
}

// Label renders field labels.
func Label(appTheme theme.Theme, value string) string {
	return roleStyle(appTheme, RoleLabel).Render(value)
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
