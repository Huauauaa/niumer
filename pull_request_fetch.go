package main

import (
	"bytes"
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

// PullRequestListItem is the normalized shape exposed to Wails / frontend (camelCase JSON).
type PullRequestListItem struct {
	ID           int    `json:"id"`
	Number       int    `json:"number"`
	URL          string `json:"url"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	SourceBranch string `json:"sourceBranch"`
	TargetBranch string `json:"targetBranch"`
	State        string `json:"state"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// PullRequestListResponse is always returned to the UI (derived when upstream is a bare JSON array).
type PullRequestListResponse struct {
	Items      []PullRequestListItem `json:"items"`
	Total      int                   `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"pageSize"`
	TotalPages int                   `json:"totalPages"`
}

// pullRequestWireItem matches the upstream merge-request list element (snake_case).
type pullRequestWireAuthor struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	NameCn   string `json:"name_cn"`
}

type pullRequestWireItem struct {
	ID           int64                  `json:"id"`
	IID          int                    `json:"iid"`
	Title        string                 `json:"title"`
	SourceBranch string                 `json:"source_branch"`
	TargetBranch string                 `json:"target_branch"`
	State        string                 `json:"state"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
	WebURL       string                 `json:"web_url"`
	Author       *pullRequestWireAuthor `json:"author"`
}

func wireAuthorDisplay(a *pullRequestWireAuthor) string {
	if a == nil {
		return ""
	}
	if s := strings.TrimSpace(a.NameCn); s != "" {
		return s
	}
	if s := strings.TrimSpace(a.Name); s != "" {
		return s
	}
	return strings.TrimSpace(a.Username)
}

func normalizePullRequestState(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "opened", "open":
		return "open"
	case "merged":
		return "merged"
	case "closed":
		return "closed"
	default:
		return "open"
	}
}

func wireItemToListItem(w pullRequestWireItem) PullRequestListItem {
	id := int(w.ID)
	if w.ID != 0 && int64(id) != w.ID {
		id = 0
	}
	num := w.IID
	if num == 0 && id != 0 {
		num = id
	}
	u := strings.TrimSpace(w.WebURL)
	return PullRequestListItem{
		ID:           id,
		Number:       num,
		URL:          u,
		Title:        strings.TrimSpace(w.Title),
		Author:       wireAuthorDisplay(w.Author),
		SourceBranch: strings.TrimSpace(w.SourceBranch),
		TargetBranch: strings.TrimSpace(w.TargetBranch),
		State:        normalizePullRequestState(w.State),
		CreatedAt:    strings.TrimSpace(w.CreatedAt),
		UpdatedAt:    strings.TrimSpace(w.UpdatedAt),
	}
}

// derivePullRequestPagination fills total/totalPages when upstream only returns a page slice (no counts).
func derivePullRequestPagination(nItems, page, pageSize int) (total, totalPages int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	switch {
	case nItems == 0 && page <= 1:
		return 0, 1
	case nItems == 0 && page > 1:
		prev := page - 1
		if prev < 1 {
			prev = 1
		}
		return (prev - 1) * pageSize, prev
	case nItems > 0 && nItems < pageSize:
		return (page-1)*pageSize + nItems, page
	default:
		return (page + 1) * pageSize, page + 1
	}
}

func parsePullRequestListJSON(raw []byte, page, pageSize int) (PullRequestListResponse, error) {
	var zero PullRequestListResponse
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return zero, errors.New("empty response body")
	}
	if raw[0] == '[' {
		var wire []pullRequestWireItem
		if err := json.Unmarshal(raw, &wire); err != nil {
			return zero, fmt.Errorf("parse pull request array: %w", err)
		}
		items := make([]PullRequestListItem, 0, len(wire))
		for _, w := range wire {
			items = append(items, wireItemToListItem(w))
		}
		total, totalPages := derivePullRequestPagination(len(items), page, pageSize)
		return PullRequestListResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		}, nil
	}
	var env PullRequestListResponse
	if err := json.Unmarshal(raw, &env); err != nil {
		return zero, fmt.Errorf("parse pull request list JSON: %w", err)
	}
	return env, nil
}

func pullRequestListURLOrDefault() string {
	return envOr("PULL_REQUEST_LIST_URL", config.GetWorkHour().PullRequestListURL)
}

func pullRequestTotalURLOrDefault() string {
	return envOr("PULL_REQUEST_TOTAL_URL", config.GetWorkHour().PullRequestTotalURL)
}

const maxPullRequestPagesForTotal = 500

var prListAjaxHeaders = map[string]string{"X-Requested-With": "XMLHttpRequest"}

func (a *App) pullRequestSessionWait(ctx context.Context) error {
	if a == nil {
		return errors.New("app is nil")
	}
	select {
	case <-a.workHourBootstrapDone:
	case <-ctx.Done():
		return errors.New("等待考勤登录阶段结束超时")
	}
	a.muWorkHourAuth.RLock()
	hasCookies := len(a.workHourCookies) > 0
	bootErr := a.workHourBootstrapErr
	a.muWorkHourAuth.RUnlock()
	if !hasCookies {
		if bootErr != nil {
			return fmt.Errorf("无有效会话 Cookie: %w", bootErr)
		}
		return errors.New("无有效会话 Cookie（需先完成考勤登录与 tenant / user-info）")
	}
	return nil
}

// fetchPullRequestListOnce GET pull_request_list_url with page / page_size（与 RefreshPullRequestList 相同会话与请求头）。
func (a *App) fetchPullRequestListOnce(ctx context.Context, page, pageSize int) ([]byte, int, error) {
	listURL := strings.TrimSpace(pullRequestListURLOrDefault())
	if listURL == "" {
		return nil, 0, errors.New("未配置 pull_request_list_url 或 PULL_REQUEST_LIST_URL")
	}
	u, err := url.Parse(listURL)
	if err != nil {
		return nil, 0, fmt.Errorf("pull request 列表 URL: %w", err)
	}
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))
	u.RawQuery = q.Encode()
	client := workHourHTTPClient()
	return a.workHourGetWithHeaders(ctx, client, u.String(), prListAjaxHeaders)
}

func parsePullRequestTotalResponse(raw []byte) (int, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return 0, errors.New("empty total response body")
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n, nil
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return 0, fmt.Errorf("parse total JSON: %w", err)
	}
	if t, ok := top["total"]; ok {
		var v int
		if err := json.Unmarshal(t, &v); err != nil {
			return 0, fmt.Errorf("total field: %w", err)
		}
		return v, nil
	}
	if d, ok := top["data"]; ok {
		var v int
		if err := json.Unmarshal(d, &v); err == nil {
			return v, nil
		}
		var inner map[string]json.RawMessage
		if err := json.Unmarshal(d, &inner); err == nil {
			if t, ok := inner["total"]; ok {
				var x int
				if err := json.Unmarshal(t, &x); err != nil {
					return 0, fmt.Errorf("data.total: %w", err)
				}
				return x, nil
			}
		}
	}
	return 0, errors.New("total response: expected a JSON number or object with total")
}

// extractPullRequestTotalFromListBody 若列表 JSON 顶层含 total 则直接采用；否则返回 definitive=false 由调用方分页累加。
func extractPullRequestTotalFromListBody(raw []byte) (total int, definitive bool, err error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return 0, false, errors.New("empty list response body")
	}
	if raw[0] == '[' {
		return 0, false, nil
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return 0, false, err
	}
	if t, ok := top["total"]; ok {
		var v int
		if err := json.Unmarshal(t, &v); err != nil {
			return 0, true, err
		}
		return v, true, nil
	}
	return 0, false, nil
}

func countPullRequestItemsInListBody(raw []byte) (int, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return 0, nil
	}
	if raw[0] == '[' {
		var wire []json.RawMessage
		if err := json.Unmarshal(raw, &wire); err != nil {
			return 0, fmt.Errorf("parse pull request array: %w", err)
		}
		return len(wire), nil
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return 0, err
	}
	it, ok := top["items"]
	if !ok {
		return 0, errors.New("list object missing items")
	}
	var wire []json.RawMessage
	if err := json.Unmarshal(it, &wire); err != nil {
		return 0, fmt.Errorf("parse items: %w", err)
	}
	return len(wire), nil
}

// RefreshPullRequestList 使用与考勤相同的 Cookie 请求 pull_request_list_url（GET，query: page, page_size），
// 并带 X-Requested-With: XMLHttpRequest（与上游 Ajax 列表一致）。上游体为 MR 对象 JSON 数组（snake_case）；
// 仍兼容旧版 { items, total, page, pageSize, totalPages } 包络。默认 URL 见 configs/config.yaml，可用 PULL_REQUEST_LIST_URL 覆盖。
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := a.pullRequestSessionWait(ctx); err != nil {
		return zero, err
	}

	raw, code, err := a.fetchPullRequestListOnce(ctx, page, pageSize)
	if err != nil {
		return zero, err
	}
	if code < 200 || code >= 300 {
		return zero, fmt.Errorf("pull request 列表 HTTP %d: %s", code, truncate(string(raw), 500))
	}

	return parsePullRequestListJSON(raw, page, pageSize)
}

// GetPullRequestListTotal 返回 MR 总数。优先请求 pull_request_total_url / PULL_REQUEST_TOTAL_URL（JSON 数字或 {"total":N} 等）；
// 未配置时从 pull_request_list_url 解析顶层 total，否则按 page 分页累加条数（每页最多 50 条，最多翻 maxPullRequestPagesForTotal 页）。
func (a *App) GetPullRequestListTotal() (int, error) {
	if a == nil {
		return 0, errors.New("app is nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := a.pullRequestSessionWait(ctx); err != nil {
		return 0, err
	}
	client := workHourHTTPClient()

	totalURL := strings.TrimSpace(pullRequestTotalURLOrDefault())
	if totalURL != "" {
		raw, code, err := a.workHourGetWithHeaders(ctx, client, totalURL, prListAjaxHeaders)
		if err != nil {
			return 0, err
		}
		if code < 200 || code >= 300 {
			return 0, fmt.Errorf("pull request 总数 HTTP %d: %s", code, truncate(string(raw), 500))
		}
		return parsePullRequestTotalResponse(raw)
	}

	listURL := strings.TrimSpace(pullRequestListURLOrDefault())
	if listURL == "" {
		return 0, errors.New("未配置 pull_request_list_url 或 pull_request_total_url")
	}

	pageSize := 50
	raw, code, err := a.fetchPullRequestListOnce(ctx, 1, pageSize)
	if err != nil {
		return 0, err
	}
	if code < 200 || code >= 300 {
		return 0, fmt.Errorf("pull request 列表 HTTP %d: %s", code, truncate(string(raw), 500))
	}
	t, def, err := extractPullRequestTotalFromListBody(raw)
	if err != nil {
		return 0, err
	}
	if def {
		return t, nil
	}
	sum, err := countPullRequestItemsInListBody(raw)
	if err != nil {
		return 0, err
	}
	for page := 2; page <= maxPullRequestPagesForTotal; page++ {
		raw, code, err := a.fetchPullRequestListOnce(ctx, page, pageSize)
		if err != nil {
			return 0, err
		}
		if code < 200 || code >= 300 {
			return 0, fmt.Errorf("pull request 列表 HTTP %d (page %d): %s", code, page, truncate(string(raw), 500))
		}
		n, err := countPullRequestItemsInListBody(raw)
		if err != nil {
			return 0, fmt.Errorf("page %d: %w", page, err)
		}
		sum += n
		if n == 0 || n < pageSize {
			break
		}
	}
	return sum, nil
}
