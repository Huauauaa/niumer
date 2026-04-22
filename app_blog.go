package main

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const seedWelcomeMD = `# Welcome

Use **Blog** in the activity bar. Markdown files are stored **only on this computer** under your **blog working directory**.

- Open **Preferences** (menu bar) to see or change that folder. By default it is under your user **Documents** (for example **Documents/niumer-blog** on macOS).
- **New File** creates a **.md** file on disk in that directory.
- **Cmd+S / Ctrl+S** saves the active document.
`

// defaultBlogWorkDir returns ~/Documents/niumer-blog (OS path rules apply).
func defaultBlogWorkDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Documents", "niumer-blog"), nil
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
