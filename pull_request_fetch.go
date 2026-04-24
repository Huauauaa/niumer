package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"niumer/internal/config"
)

// Pull request list JSON shape (mockserver GET /pull-request and production upstream).
type PullRequestListItem struct {
	ID            int    `json:"id"`
	Number        int    `json:"number"`
	URL           string `json:"url"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	SourceBranch  string `json:"sourceBranch"`
	TargetBranch  string `json:"targetBranch"`
	State         string `json:"state"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// PullRequestListResponse matches the frontend and mockserver envelope.
type PullRequestListResponse struct {
	Items      []PullRequestListItem `json:"items"`
	Total      int                   `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"pageSize"`
	TotalPages int                   `json:"totalPages"`
}

func pullRequestListURLOrDefault() string {
	return envOr("PULL_REQUEST_LIST_URL", config.GetWorkHour().PullRequestListURL)
}

// RefreshPullRequestList 使用与考勤相同的 Cookie 请求 pull_request_list_url（GET，query: page, page_size），
// 与本地 mockserver 或生产环境接口一致。默认 URL 见 configs/config.yaml，可用 PULL_REQUEST_LIST_URL 覆盖。
func (a *App) RefreshPullRequestList(page, pageSize int) (PullRequestListResponse, error) {
	var zero PullRequestListResponse
	if a == nil {
		return zero, errors.New("app is nil")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	listURL := strings.TrimSpace(pullRequestListURLOrDefault())
	if listURL == "" {
		return zero, errors.New("未配置 pull_request_list_url 或 PULL_REQUEST_LIST_URL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	select {
	case <-a.workHourBootstrapDone:
	case <-ctx.Done():
		return zero, errors.New("等待考勤登录阶段结束超时")
	}

	a.muWorkHourAuth.RLock()
	hasCookies := len(a.workHourCookies) > 0
	bootErr := a.workHourBootstrapErr
	a.muWorkHourAuth.RUnlock()
	if !hasCookies {
		if bootErr != nil {
			return zero, fmt.Errorf("无有效会话 Cookie: %w", bootErr)
		}
		return zero, errors.New("无有效会话 Cookie（需先完成考勤登录与 tenant / user-info）")
	}

	u, err := url.Parse(listURL)
	if err != nil {
		return zero, fmt.Errorf("pull request 列表 URL: %w", err)
	}
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))
	u.RawQuery = q.Encode()

	client := workHourHTTPClient()
	cookies := a.workHourCookiesForHTTP()
	raw, code, err := getWithCookies(ctx, client, u.String(), cookies)
	if err != nil {
		return zero, err
	}
	if code < 200 || code >= 300 {
		return zero, fmt.Errorf("pull request 列表 HTTP %d: %s", code, truncate(string(raw), 500))
	}

	var resp PullRequestListResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return zero, fmt.Errorf("解析 pull request 列表 JSON: %w", err)
	}
	return resp, nil
}
