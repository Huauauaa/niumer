package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// WorkHourSettings holds URLs and selectors for attendance sync (chromedp + HTTP).
type WorkHourSettings struct {
	LoginURL  string `yaml:"login_url"`
	WaitCSS   string `yaml:"wait_css"`
	TenantURL string `yaml:"tenant_url"`
	// UserInfoURL is the POST URL for the attendance user-info endpoint (hrId, shiftNameZh, etc.).
	UserInfoURL string `yaml:"user_info_url"`
	// WorkHourURL is the POST URL for the attendance records / work-hour list API.
	WorkHourURL string `yaml:"workhour_url"`
	// PullRequestListURL is the GET URL for the pull request list (query: page, page_size).
	// Upstream returns a JSON array of MR objects (snake_case); see pull_request_fetch.go.
	PullRequestListURL string `yaml:"pull_request_list_url"`
	// PullRequestTotalURL is an optional GET URL for the total MR count (JSON: {"total": N} or N).
	// When empty, GetPullRequestListTotal derives the count from PullRequestListURL.
	PullRequestTotalURL string `yaml:"pull_request_total_url"`
}

// Root is the top-level application YAML shape.
type Root struct {
	WorkHour WorkHourSettings `yaml:"workhour"`
}

var current Root

// GetWorkHour returns merged work-hour settings (after Load).
func GetWorkHour() WorkHourSettings {
	return current.WorkHour
}

// Load reads configs/config.yaml from fsys, then merges configs/config.<env>.yaml
// when present. env is taken from NIUMER_ENV; if empty, defaults to "dev".
func Load(fsys fs.FS) error {
	baseBytes, err := fs.ReadFile(fsys, "configs/config.yaml")
	if err != nil {
		return fmt.Errorf("read configs/config.yaml: %w", err)
	}
	if err := yaml.Unmarshal(baseBytes, &current); err != nil {
		return fmt.Errorf("parse configs/config.yaml: %w", err)
	}

	env := strings.TrimSpace(os.Getenv("NIUMER_ENV"))
	if env == "" {
		env = "dev"
	}
	overlayPath := fmt.Sprintf("configs/config.%s.yaml", env)
	overlayBytes, err := fs.ReadFile(fsys, overlayPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read %s: %w", overlayPath, err)
	}
	var over Root
	if err := yaml.Unmarshal(overlayBytes, &over); err != nil {
		return fmt.Errorf("parse %s: %w", overlayPath, err)
	}
	current.WorkHour = mergeWorkHour(current.WorkHour, over.WorkHour)
	return nil
}

func mergeWorkHour(base, over WorkHourSettings) WorkHourSettings {
	out := base
	if strings.TrimSpace(over.LoginURL) != "" {
		out.LoginURL = over.LoginURL
	}
	if strings.TrimSpace(over.WaitCSS) != "" {
		out.WaitCSS = over.WaitCSS
	}
	if strings.TrimSpace(over.TenantURL) != "" {
		out.TenantURL = over.TenantURL
	}
	if strings.TrimSpace(over.UserInfoURL) != "" {
		out.UserInfoURL = over.UserInfoURL
	}
	if strings.TrimSpace(over.WorkHourURL) != "" {
		out.WorkHourURL = over.WorkHourURL
	}
	if strings.TrimSpace(over.PullRequestListURL) != "" {
		out.PullRequestListURL = over.PullRequestListURL
	}
	if strings.TrimSpace(over.PullRequestTotalURL) != "" {
		out.PullRequestTotalURL = over.PullRequestTotalURL
	}
	return out
}
