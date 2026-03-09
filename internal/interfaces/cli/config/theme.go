package config

// ThemeMode controls startup theme selection for the TUI.
type ThemeMode string

const (
	ThemeModeAuto  ThemeMode = "auto"
	ThemeModeLight ThemeMode = "light"
	ThemeModeDark  ThemeMode = "dark"
)

func (m ThemeMode) Valid() bool {
	switch m {
	case ThemeModeAuto, ThemeModeLight, ThemeModeDark:
		return true
	default:
		return false
	}
}

// ThemeConfig controls TUI theme resolution.
type ThemeConfig struct {
	Mode ThemeMode `name:"theme" help:"TUI theme mode." enum:"auto,light,dark" env:"TERO_THEME" default:"auto"`
}
