package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// defaultWorkHourDBPath returns the built-in path for work_hour.db (under the user config directory).
func defaultWorkHourDBPath() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(d, "niumer")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "work_hour.db"), nil
}

// resolvedWorkHourDBPath returns configured SQLite file path or the default file path.
func (a *App) resolvedWorkHourDBPath() (string, error) {
	c, err := readAppConfig()
	if err != nil {
		return defaultWorkHourDBPath()
	}
	p := strings.TrimSpace(c.WorkHourDBPath)
	if p == "" {
		return defaultWorkHourDBPath()
	}
	return filepath.Clean(p), nil
}

// GetDefaultWorkHourDBPath returns the built-in default work hour database path (for display).
func (a *App) GetDefaultWorkHourDBPath() string {
	s, err := defaultWorkHourDBPath()
	if err != nil {
		return ""
	}
	return s
}

// GetWorkHourDBPath returns the effective SQLite database file path used by the app.
func (a *App) GetWorkHourDBPath() (string, error) {
	return a.resolvedWorkHourDBPath()
}

// SetWorkHourDBPath sets the SQLite database file path. Pass empty string to use the default location.
func (a *App) SetWorkHourDBPath(path string) error {
	path = strings.TrimSpace(path)
	c, err := readAppConfig()
	if err != nil {
		c = appConfig{}
	}
	if path == "" {
		c.WorkHourDBPath = ""
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
	c.WorkHourDBPath = path
	return writeAppConfig(c)
}

// ChooseWorkHourDBPath opens a save dialog to pick the SQLite file path. Returns "" if cancelled.
func (a *App) ChooseWorkHourDBPath() (string, error) {
	if a.ctx == nil {
		return "", errors.New("app not ready")
	}
	opts := runtime.SaveDialogOptions{
		Title:                "Choose SQLite database file",
		DefaultFilename:      "work_hour.db",
		CanCreateDirectories: true,
		Filters: []runtime.FileFilter{
			{DisplayName: "SQLite (*.db;*.sqlite)", Pattern: "*.db;*.sqlite"},
		},
	}
	current, err := a.resolvedWorkHourDBPath()
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
		if d, err := defaultWorkHourDBPath(); err == nil {
			if parent := filepath.Dir(d); parent != d {
				if st, err := os.Stat(parent); err == nil && st.IsDir() {
					opts.DefaultDirectory = parent
				}
			}
		}
	}
	return runtime.SaveFileDialog(a.ctx, opts)
}
