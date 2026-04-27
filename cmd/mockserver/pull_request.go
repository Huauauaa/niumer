// Pull-request mock: list, total count, HTML preview. See registerPullRequestRoutes.
package main

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// registerPullRequestRoutes registers GET /pull-request, /pull-request/total, /pr-preview/{id}.
func registerPullRequestRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /pull-request", handlePullRequestList)
	mux.HandleFunc("GET /pull-request/total", handlePullRequestTotal)
	mux.HandleFunc("GET /pr-preview/{id}", handlePRPreview)
}

const pullRequestMockTotal = 47

func requestBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func mockMergeRequestProject(pid int, repoName string) map[string]any {
	created := "2024-04-24T15:05:07.034+08:00"
	return map[string]any{
		"id":                  pid,
		"description":         "",
		"name":                repoName,
		"name_with_namespace": fmt.Sprintf("Network_Automation_Platform / TPE / %s", repoName),
		"path":                repoName,
		"path_with_namespace": "Network_Automation_Platform/TPE/" + repoName,
		"develop_mode":        nil,
		"created_at":          created,
		"updated_at":          created,
		"archived":            false,
		"is_kia":              false,
		"ssh_url_to_repo":     "git@example.com:group/" + repoName + ".git",
		"http_url_to_repo":    "https://example.com/group/" + repoName,
		"s3_url_to_repo":      nil,
		"repo_model":          nil,
		"web_url":             nil,
		"readme_url":          nil,
		"repo_type":           "code",
		"product_id":          nil,
		"product_name":        nil,
	}
}

// pullRequestWireMockMap matches production merge-request list elements (snake_case); web_url points at local pr-preview.
func pullRequestWireMockMap(n int, base string) map[string]any {
	base = strings.TrimSuffix(base, "/")
	pid := 3939700 + (n % 50)
	repoName := "TPEController-ops-console"
	target := "master"
	if n%4 == 0 {
		target = "develop"
	}
	states := []string{"opened", "merged", "closed"}
	state := states[n%3]
	now := time.Now().In(time.FixedZone("GMT+8", 8*3600)).Format("2006-01-02T15:04:05.000Z07:00")
	webURL := fmt.Sprintf("%s/pr-preview/%d", base, n)
	return map[string]any{
		"id":                int64(46573000 + n),
		"iid":               n,
		"title":             fmt.Sprintf("feat: mock merge request #%d (%s)", n, state),
		"source_branch":     fmt.Sprintf("feature/pr-%d", n),
		"target_branch":     target,
		"state":             state,
		"created_at":        now,
		"updated_at":        now,
		"source_project_id": pid,
		"review_mode":       "vote",
		"author": map[string]any{
			"id":          884800 + n,
			"name":        fmt.Sprintf("Dev User %d", n),
			"username":    fmt.Sprintf("dev%d", (n%9)+1),
			"state":       "active",
			"avatar_url":  "",
			"avatar_path": nil,
			"email":       "dev@example.com",
			"name_cn":     fmt.Sprintf("开发者%d", n%20),
			"web_url":     "",
			"nick_name":   nil,
			"tenant_name": nil,
		},
		"closed_at":                         nil,
		"closed_by":                         nil,
		"merged_at":                         nil,
		"merged_by":                         nil,
		"pipeline_status":                   "success",
		"codequality_status":                "success",
		"pipeline_status_with_code_quality": "success",
		"notes":                             0,
		"source_project":                    mockMergeRequestProject(pid, repoName),
		"target_project":                    mockMergeRequestProject(pid, repoName),
		"web_url":                           webURL,
		"added_lines":                       2,
		"removed_lines":                     2,
		"merge_request_type":                "MergeRequest",
		"source_git_url":                    "git@example.com:group/" + repoName + ".git",
		"labels":                            []any{},
		"score":                             1,
		"min_merged_score":                  3,
		"source_product_id":                 nil,
		"target_product_id":                 nil,
		"product_name":                      nil,
		"notes_count": map[string]any{
			"notes_count":            0,
			"unresolved_notes_count": 0,
			"already_resolved_count": 0,
			"need_resolved_count":    0,
		},
		"approval_approvers_required_passed": nil,
		"approval_reviewers_required_passed": nil,
		"approved_count":                     nil,
		"reviewed_count":                     nil,
		"is_conflict":                        false,
		"custom_ctrl_items_passed":           false,
		"region":                             "prod-dgg-green",
	}
}

func requirePullRequestAjaxHeader(w http.ResponseWriter, r *http.Request) bool {
	if strings.TrimSpace(r.Header.Get("X-Requested-With")) != "XMLHttpRequest" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "X-Requested-With must be XMLHttpRequest (same as production PR list client)",
		})
		return false
	}
	return true
}

func handlePullRequestTotal(w http.ResponseWriter, r *http.Request) {
	if !requirePullRequestAjaxHeader(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"total": pullRequestMockTotal})
}

func handlePullRequestList(w http.ResponseWriter, r *http.Request) {
	if !requirePullRequestAjaxHeader(w, r) {
		return
	}
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	base := requestBaseURL(r)
	total := pullRequestMockTotal
	start := (page - 1) * pageSize
	if start >= total {
		writeJSON(w, http.StatusOK, []map[string]any{})
		return
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	items := make([]map[string]any, 0, end-start)
	for i := start; i < end; i++ {
		items = append(items, pullRequestWireMockMap(i+1, base))
	}
	writeJSON(w, http.StatusOK, items)
}

func handlePRPreview(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	n, err := strconv.Atoi(idStr)
	if err != nil || n < 1 || n > pullRequestMockTotal {
		http.NotFound(w, r)
		return
	}
	base := requestBaseURL(r)
	m := pullRequestWireMockMap(n, base)
	title, _ := m["title"].(string)
	author := ""
	if auth, ok := m["author"].(map[string]any); ok {
		if s, ok := auth["name_cn"].(string); ok && strings.TrimSpace(s) != "" {
			author = s
		} else if s, ok := auth["name"].(string); ok && strings.TrimSpace(s) != "" {
			author = s
		} else if s, ok := auth["username"].(string); ok {
			author = s
		}
	}
	src, _ := m["source_branch"].(string)
	tgt, _ := m["target_branch"].(string)
	st, _ := m["state"].(string)
	previewURL := ""
	if u, ok := m["web_url"].(string); ok {
		previewURL = u
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>PR #%d</title>
<style>body{font-family:system-ui,sans-serif;background:#1e1e1e;color:#ccc;padding:1.5rem;line-height:1.5;}
code{color:#ce9178}a{color:#4fc1ff}</style></head><body>
<h1 style="font-size:1.1rem;margin:0 0 0.5rem">Pull request #%d</h1>
<p style="color:#858585;margin:0 0 1rem">Mock preview from <code>%s</code></p>
<dl style="margin:0;display:grid;grid-template-columns:10rem 1fr;gap:0.35rem 1rem;font-size:14px">
<dt>Title</dt><dd>%s</dd>
<dt>Author</dt><dd>%s</dd>
<dt>Branches</dt><dd><code>%s</code> → <code>%s</code></dd>
<dt>State</dt><dd>%s</dd>
</dl>
<p style="margin-top:1.5rem"><a href="%s">Open this URL</a> in a full browser if the embedded view is limited.</p>
</body></html>`,
		n, n, html.EscapeString(r.URL.Path),
		html.EscapeString(title), html.EscapeString(author),
		html.EscapeString(src), html.EscapeString(tgt), html.EscapeString(st),
		html.EscapeString(previewURL),
	)
}
