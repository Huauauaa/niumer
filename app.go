package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

type appConfig struct {
	BlogWorkDir          string `json:"blogWorkDir"`
	WorkHourDBPath       string `json:"workHourDbPath"`
	JsonFormatterWorkDir string `json:"jsonFormatterWorkDir"`
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

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// On first launch, persist the default blog directory so Markdown always has an
	// explicit on-disk location and Preferences shows the resolved path.
	if err := a.ensureDefaultBlogWorkDirInConfig(); err != nil {
		log.Printf("niumer: default blog directory: %v", err)
	}
	if err := a.ensureDefaultJsonFormatterWorkDirInConfig(); err != nil {
		log.Printf("niumer: default JSON formatter directory: %v", err)
	}
}

// ensureDefaultBlogWorkDirInConfig writes ~/Documents/niumer-blog (or OS equivalent)
// into config when blogWorkDir has never been set.
func (a *App) ensureDefaultBlogWorkDirInConfig() error {
	c, err := readAppConfig()
	if err != nil {
		c = appConfig{}
	}
	if strings.TrimSpace(c.BlogWorkDir) != "" {
		return nil
	}
	d, err := defaultBlogWorkDir()
	if err != nil {
		return err
	}
	return a.SetBlogWorkDir(d)
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

func configFilePath() (string, error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "niumer", "config.json"), nil
}

// defaultBlogWorkDir returns ~/Documents/niumer-blog (OS path rules apply).
func defaultBlogWorkDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Documents", "niumer-blog"), nil
}

// GetDefaultBlogWorkDir returns the built-in default blog directory (not necessarily created on disk).
func (a *App) GetDefaultBlogWorkDir() string {
	s, err := defaultBlogWorkDir()
	if err != nil {
		return ""
	}
	return s
}

// GetBlogWorkDir returns the configured blog working directory, or the default if unset.
func (a *App) GetBlogWorkDir() string {
	c, err := readAppConfig()
	if err != nil || c.BlogWorkDir == "" {
		d, _ := defaultBlogWorkDir()
		return d
	}
	return c.BlogWorkDir
}

// SetBlogWorkDir persists the blog working directory and ensures it exists.
func (a *App) SetBlogWorkDir(path string) error {
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
	c.BlogWorkDir = path
	return writeAppConfig(c)
}

// ChooseBlogWorkDir opens a native folder picker. Returns "" if the user cancels.
func (a *App) ChooseBlogWorkDir() (string, error) {
	if a.ctx == nil {
		return "", errors.New("app not ready")
	}
	opts := runtime.OpenDialogOptions{
		Title: "Choose blog working directory",
	}
	current := a.GetBlogWorkDir()
	if st, err := os.Stat(current); err == nil && st.IsDir() {
		opts.DefaultDirectory = current
	} else if d, err := defaultBlogWorkDir(); err == nil {
		if parent := filepath.Dir(d); parent != d {
			if st, err := os.Stat(parent); err == nil && st.IsDir() {
				opts.DefaultDirectory = parent
			}
		}
	}
	return runtime.OpenDirectoryDialog(a.ctx, opts)
}

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

const seedWelcomeMD = `# Welcome

Use **Blog** in the activity bar. Markdown files are stored **only on this computer** under your **blog working directory**.

- Open **Preferences** (menu bar) to see or change that folder. By default it is under your user **Documents** (for example **Documents/niumer-blog** on macOS).
- **New File** creates a **.md** file on disk in that directory.
- **Cmd+S / Ctrl+S** saves the active document.
`

func sanitizeBlogMarkdownName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("empty file name")
	}
	base := filepath.Base(name)
	if base != name || strings.Contains(name, "..") {
		return "", errors.New("invalid file name")
	}
	if len(base) > 200 {
		return "", errors.New("file name too long")
	}
	if len(base) < 4 || !strings.HasSuffix(strings.ToLower(base), ".md") {
		return "", errors.New("must be a .md file")
	}
	stem := base[:len(base)-3]
	for _, r := range stem {
		if r > unicode.MaxASCII {
			return "", errors.New("invalid character in file name")
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' || r == '_' || r == '.' {
			continue
		}
		return "", errors.New("invalid character in file name")
	}
	return base, nil
}

// ListBlogMarkdownFiles returns sorted file names (e.g. "a.md") in the blog directory.
func (a *App) ListBlogMarkdownFiles() ([]string, error) {
	dir := a.GetBlogWorkDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasSuffix(strings.ToLower(n), ".md") {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names, nil
}

// ReadBlogFile reads UTF-8 content of a markdown file in the blog directory.
func (a *App) ReadBlogFile(fileName string) (string, error) {
	base, err := sanitizeBlogMarkdownName(fileName)
	if err != nil {
		return "", err
	}
	path := filepath.Join(a.GetBlogWorkDir(), base)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteBlogFile writes content to a markdown file (create or overwrite).
func (a *App) WriteBlogFile(fileName string, content string) error {
	base, err := sanitizeBlogMarkdownName(fileName)
	if err != nil {
		return err
	}
	dir := a.GetBlogWorkDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, base)
	return os.WriteFile(path, []byte(content), 0o644)
}

// DeleteBlogFile removes a markdown file from the blog directory.
func (a *App) DeleteBlogFile(fileName string) error {
	base, err := sanitizeBlogMarkdownName(fileName)
	if err != nil {
		return err
	}
	path := filepath.Join(a.GetBlogWorkDir(), base)
	return os.Remove(path)
}

// RenameBlogFile renames a markdown file within the blog directory.
func (a *App) RenameBlogFile(oldName, newName string) error {
	oldBase, err := sanitizeBlogMarkdownName(oldName)
	if err != nil {
		return err
	}
	newBase, err := sanitizeBlogMarkdownName(newName)
	if err != nil {
		return err
	}
	if oldBase == newBase {
		return nil
	}
	dir := a.GetBlogWorkDir()
	oldPath := filepath.Join(dir, oldBase)
	newPath := filepath.Join(dir, newBase)
	if _, err := os.Stat(oldPath); err != nil {
		if os.IsNotExist(err) {
			return errors.New("source file does not exist")
		}
		return err
	}
	if _, err := os.Stat(newPath); err == nil {
		return errors.New("a file with that name already exists")
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.Rename(oldPath, newPath)
}

// EnsureWelcomeBlogFile creates Welcome.md with default content if the blog folder has no .md files.
func (a *App) EnsureWelcomeBlogFile() error {
	names, err := a.ListBlogMarkdownFiles()
	if err != nil {
		return err
	}
	if len(names) > 0 {
		return nil
	}
	return a.WriteBlogFile("Welcome.md", seedWelcomeMD)
}
