package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// GetUserSettingsFilePath returns the absolute path to User/settings.json (VS Code–style).
func (a *App) GetUserSettingsFilePath() string {
	p, err := userSettingsFilePath()
	if err != nil {
		return ""
	}
	return p
}

// ReadUserSettingsJSON returns pretty-printed JSON for editing. If the file is missing,
// returns a minimal template built from the current effective configuration.
func (a *App) ReadUserSettingsJSON() (string, error) {
	p, err := userSettingsFilePath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return a.defaultUserSettingsJSON()
		}
		return "", err
	}
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return a.defaultUserSettingsJSON()
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (a *App) defaultUserSettingsJSON() (string, error) {
	wh, err := a.GetWorkHourDBPath()
	if err != nil {
		wh = ""
	}
	rm, err := a.GetReminderDBPath()
	if err != nil {
		rm = ""
	}
	th := a.GetUITheme()
	if th == "" {
		th = "dark"
	}
	c := appConfig{
		BlogWorkDir:          a.GetBlogWorkDir(),
		JsonFormatterWorkDir: a.GetJsonFormatterWorkDir(),
		WorkHourDBPath:       wh,
		ReminderDBPath:       rm,
		Theme:                th,
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// WriteUserSettingsJSON parses JSON, applies the same validation as the Preferences form,
// and persists via the existing setters (creates directories / parent paths as needed).
func (a *App) WriteUserSettingsJSON(content string) error {
	var m map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &m); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	blog := strings.TrimSpace(stringField(m, "blogWorkDir"))
	jf := strings.TrimSpace(stringField(m, "jsonFormatterWorkDir"))
	wh := strings.TrimSpace(stringField(m, "workHourDbPath"))
	rm := strings.TrimSpace(stringField(m, "reminderDbPath"))

	for _, kv := range []struct {
		key string
	}{
		{"blogWorkDir"},
		{"jsonFormatterWorkDir"},
		{"workHourDbPath"},
		{"reminderDbPath"},
		{"theme"},
	} {
		if v, ok := m[kv.key]; ok && v != nil {
			if _, ok := v.(string); !ok {
				return fmt.Errorf("%q must be a JSON string", kv.key)
			}
		}
	}

	if blog == "" {
		return errors.New("blogWorkDir cannot be empty")
	}
	if jf == "" {
		return errors.New("jsonFormatterWorkDir cannot be empty")
	}
	if err := a.SetBlogWorkDir(blog); err != nil {
		return err
	}
	if err := a.SetJsonFormatterWorkDir(jf); err != nil {
		return err
	}
	if err := a.SetWorkHourDBPath(wh); err != nil {
		return err
	}
	if err := a.SetReminderDBPath(rm); err != nil {
		return err
	}

	if _, has := m["theme"]; !has {
		return nil
	}
	c, err := readAppConfig()
	if err != nil {
		c = appConfig{}
	}
	if m["theme"] == nil {
		c.Theme = ""
	} else if s, ok := m["theme"].(string); ok {
		s = strings.TrimSpace(strings.ToLower(s))
		if s == "" {
			c.Theme = ""
		} else if s != "dark" && s != "light" {
			return fmt.Errorf(`theme must be "dark" or "light"`)
		} else {
			c.Theme = s
		}
	} else {
		return fmt.Errorf(`theme must be a JSON string`)
	}
	return writeAppConfig(c)
}

func stringField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
