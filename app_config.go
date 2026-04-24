package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// appConfig holds user-editable paths from Preferences (like VS Code User settings).
type appConfig struct {
	BlogWorkDir          string `json:"blogWorkDir"`
	WorkHourDBPath       string `json:"workHourDbPath"`
	ReminderDBPath       string `json:"reminderDbPath"`
	JsonFormatterWorkDir string `json:"jsonFormatterWorkDir"`
	// Theme is "dark" or "light" (same as workbench color theme in spirit).
	Theme string `json:"theme,omitempty"`
	// OpenAI-compatible chat API (e.g. https://api.deepseek.com).
	AIBaseURL string `json:"aiBaseUrl,omitempty"`
	AIAPIKey  string `json:"aiApiKey,omitempty"`
	AIModel   string `json:"aiModel,omitempty"`
}

// userDataRoot is the app folder under the OS user config directory, e.g.:
//
//	macOS: ~/Library/Application Support/niumer
//	Linux: ~/.config/niumer (or $XDG_CONFIG_HOME/niumer)
//	Windows: %AppData%\niumer
func userDataRoot() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "niumer"), nil
}

// userSettingsFilePath is the primary preferences file, mirroring VS Code’s
// …/User/settings.json layout: <userDataRoot>/User/settings.json
func userSettingsFilePath() (string, error) {
	root, err := userDataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "User", "settings.json"), nil
}

// legacyConfigFilePath is the legacy path (<userDataRoot>/config.json) before User/settings.json.
func legacyConfigFilePath() (string, error) {
	root, err := userDataRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "config.json"), nil
}

func readAppConfig() (appConfig, error) {
	settingsPath, err := userSettingsFilePath()
	if err != nil {
		return appConfig{}, err
	}
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		var c appConfig
		if err := json.Unmarshal(data, &c); err != nil {
			return appConfig{}, err
		}
		return c, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return appConfig{}, err
	}

	legacyPath, err := legacyConfigFilePath()
	if err != nil {
		return appConfig{}, err
	}
	data, err = os.ReadFile(legacyPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return appConfig{}, nil
		}
		return appConfig{}, err
	}
	var c appConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return appConfig{}, err
	}
	// One-time migration to User/settings.json
	if werr := writeAppConfig(c); werr == nil {
		_ = os.Remove(legacyPath)
	}
	return c, nil
}

func writeAppConfig(c appConfig) error {
	p, err := userSettingsFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}
