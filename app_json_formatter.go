package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// jsonFormatterDraftFile is the on-disk buffer for the Tool JSON formatter.
const jsonFormatterDraftFile = "draft.json"

// defaultJsonFormatterWorkDir returns ~/Documents/niumer-json-formatter (OS path rules apply).
// Same layout as the blog default: user home + "Documents" + app folder (macOS, Linux, Windows).
func defaultJsonFormatterWorkDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Documents", "niumer-json-formatter"), nil
}

func (a *App) ensureDefaultJsonFormatterWorkDirInConfig() error {
	c, err := readAppConfig()
	if err != nil {
		c = appConfig{}
	}
	if strings.TrimSpace(c.JsonFormatterWorkDir) != "" {
		return nil
	}
	d, err := defaultJsonFormatterWorkDir()
	if err != nil {
		return err
	}
	return a.SetJsonFormatterWorkDir(d)
}

// GetDefaultJsonFormatterWorkDir returns the built-in default directory (for display).
func (a *App) GetDefaultJsonFormatterWorkDir() string {
	s, err := defaultJsonFormatterWorkDir()
	if err != nil {
		return ""
	}
	return s
}

// GetJsonFormatterWorkDir returns the configured directory, or the default if unset.
func (a *App) GetJsonFormatterWorkDir() string {
	c, err := readAppConfig()
	if err != nil || c.JsonFormatterWorkDir == "" {
		d, _ := defaultJsonFormatterWorkDir()
		return d
	}
	return c.JsonFormatterWorkDir
}

// SetJsonFormatterWorkDir persists the JSON formatter working directory and ensures it exists.
func (a *App) SetJsonFormatterWorkDir(path string) error {
	if path == "" {
		return errors.New("path is empty")
	}
	path = filepath.Clean(path)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	c, err := readAppConfig()
	if err != nil {
		c = appConfig{}
	}
	c.JsonFormatterWorkDir = path
	return writeAppConfig(c)
}

// ChooseJsonFormatterWorkDir opens a native folder picker. Returns "" if the user cancels.
func (a *App) ChooseJsonFormatterWorkDir() (string, error) {
	if a.ctx == nil {
		return "", errors.New("app not ready")
	}
	opts := runtime.OpenDialogOptions{
		Title: "Choose JSON formatter directory",
	}
	current := a.GetJsonFormatterWorkDir()
	if st, err := os.Stat(current); err == nil && st.IsDir() {
		opts.DefaultDirectory = current
	} else if d, err := defaultJsonFormatterWorkDir(); err == nil {
		if parent := filepath.Dir(d); parent != d {
			if st, err := os.Stat(parent); err == nil && st.IsDir() {
				opts.DefaultDirectory = parent
			}
		}
	}
	return runtime.OpenDirectoryDialog(a.ctx, opts)
}

func (a *App) jsonFormatterDraftPath() string {
	return filepath.Join(a.GetJsonFormatterWorkDir(), jsonFormatterDraftFile)
}

// ReadJsonFormatterDraft reads UTF-8 draft content. Returns ("", nil) if the file does not exist.
func (a *App) ReadJsonFormatterDraft() (string, error) {
	p := a.jsonFormatterDraftPath()
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// WriteJsonFormatterDraft writes UTF-8 content to the draft file under the configured directory.
func (a *App) WriteJsonFormatterDraft(content string) error {
	dir := a.GetJsonFormatterWorkDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(a.jsonFormatterDraftPath(), []byte(content), 0o644)
}
