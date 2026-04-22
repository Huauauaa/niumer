package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type appConfig struct {
	BlogWorkDir          string `json:"blogWorkDir"`
	WorkHourDBPath       string `json:"workHourDbPath"`
	ReminderDBPath       string `json:"reminderDbPath"`
	JsonFormatterWorkDir string `json:"jsonFormatterWorkDir"`
}

func configFilePath() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "niumer", "config.json"), nil
}

func readAppConfig() (appConfig, error) {
	p, err := configFilePath()
	if err != nil {
		return appConfig{}, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return appConfig{}, nil
		}
		return appConfig{}, err
	}
	var c appConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return appConfig{}, err
	}
	return c, nil
}

func writeAppConfig(c appConfig) error {
	p, err := configFilePath()
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
