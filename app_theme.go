package main

import (
	"fmt"
	"strings"
)

// GetUITheme returns "dark", "light", or "" when settings.json has no theme key yet
// (frontend may keep using localStorage until the user saves theme to disk).
func (a *App) GetUITheme() string {
	c, err := readAppConfig()
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(strings.ToLower(c.Theme))
	switch s {
	case "light":
		return "light"
	case "dark":
		return "dark"
	default:
		return ""
	}
}

// SetUITheme persists color theme to User/settings.json ("dark" or "light").
func (a *App) SetUITheme(theme string) error {
	theme = strings.TrimSpace(strings.ToLower(theme))
	if theme != "dark" && theme != "light" {
		return fmt.Errorf(`theme must be "dark" or "light"`)
	}
	c, err := readAppConfig()
	if err != nil {
		c = appConfig{}
	}
	c.Theme = theme
	return writeAppConfig(c)
}
