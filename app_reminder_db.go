package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// defaultReminderDBPath returns the built-in path for reminder.db (next to work_hour.db under user config).
func defaultReminderDBPath() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(d, "niumer")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "reminder.db"), nil
}

func (a *App) resolvedReminderDBPath() (string, error) {
	c, err := readAppConfig()
	if err != nil {
		return defaultReminderDBPath()
	}
	p := strings.TrimSpace(c.ReminderDBPath)
	if p == "" {
		return defaultReminderDBPath()
	}
	return filepath.Clean(p), nil
}

// GetDefaultReminderDBPath returns the built-in default reminder database path (for display).
func (a *App) GetDefaultReminderDBPath() string {
	s, err := defaultReminderDBPath()
	if err != nil {
		return ""
	}
	return s
}

// GetReminderDBPath returns the effective SQLite path for custom reminders.
func (a *App) GetReminderDBPath() (string, error) {
	return a.resolvedReminderDBPath()
}

// SetReminderDBPath sets the SQLite file path for reminders. Empty string resets to default.
func (a *App) SetReminderDBPath(path string) error {
	path = strings.TrimSpace(path)
	c, err := readAppConfig()
	if err != nil {
		c = appConfig{}
	}
	if path == "" {
		c.ReminderDBPath = ""
		return writeAppConfig(c)
	}
	path = filepath.Clean(path)
	if st, err := os.Stat(path); err == nil && st.IsDir() {
		return errors.New("path must be a database file, not a directory")
	}
	parent := filepath.Dir(path)
	if parent != "." && parent != "" && parent != path {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return err
		}
	}
	c.ReminderDBPath = path
	return writeAppConfig(c)
}

// ChooseReminderDBPath opens a save dialog to pick the SQLite file path. Returns "" if cancelled.
func (a *App) ChooseReminderDBPath() (string, error) {
	if a.ctx == nil {
		return "", errors.New("app not ready")
	}
	opts := runtime.SaveDialogOptions{
		Title:                "Choose reminder SQLite database file",
		DefaultFilename:      "reminder.db",
		CanCreateDirectories: true,
		Filters: []runtime.FileFilter{
			{DisplayName: "SQLite (*.db;*.sqlite)", Pattern: "*.db;*.sqlite"},
		},
	}
	current, err := a.resolvedReminderDBPath()
	if err != nil {
		current = ""
	}
	if current != "" {
		dir := filepath.Dir(current)
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			opts.DefaultDirectory = dir
		}
	}
	if opts.DefaultDirectory == "" {
		if d, err := defaultReminderDBPath(); err == nil {
			if parent := filepath.Dir(d); parent != d {
				if st, err := os.Stat(parent); err == nil && st.IsDir() {
					opts.DefaultDirectory = parent
				}
			}
		}
	}
	return runtime.SaveFileDialog(a.ctx, opts)
}
