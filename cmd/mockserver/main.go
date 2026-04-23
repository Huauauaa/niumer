// Work-hour API mock: same routes and JSON shapes as the placeholder
// http://127.0.0.1:17890/id, /user-info, /work-hour (defaults in workhour_fetch.go).
// Also: GET /pull-request (paginated PR list), GET /pr-preview/{id} (HTML for iframe).
//
// Run from repo root:
//
//	go run ./cmd/mockserver
//
// Then point the app at it, for example:
//
//	export WORK_HOUR_TENANT_URL=http://127.0.0.1:17890/id
//	export WORK_HOUR_USER_INFO_URL=http://127.0.0.1:17890/user-info
//	export WORK_HOUR_API_URL=http://127.0.0.1:17890/work-hour
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:17890", "listen address")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
	// Minimal page for chromedp / playwright (WORK_HOUR_WAIT_CSS default).
	mux.HandleFunc("GET /login", handleLogin)
	mux.HandleFunc("POST /id", handleTenant)
	mux.HandleFunc("POST /user-info", handleUserInfo)
	mux.HandleFunc("POST /work-hour", handleWorkHour)
	mux.HandleFunc("GET /pull-request", handlePullRequestList)
	mux.HandleFunc("GET /pr-preview/{id}", handlePRPreview)
	log.Printf("mockserver listening on http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, withCORS(logRequests(mux))))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// loginHTML is picked so chromedp.WaitVisible succeeds in headless Chrome (non-zero
// layout box). Includes .foo and .search-total__num for WORK_HOUR_WAIT_CSS overrides.
const loginHTML = `<!doctype html><html lang="zh"><head><meta charset="utf-8">
<style>
.foo,.search-total__num{display:inline-block!important;width:32px!important;height:24px!important;
line-height:24px!important;visibility:visible!important;opacity:1!important;font-size:16px!important;}
</style></head><body>
<span class="foo search-total__num" id="mock-login-ready">0</span>
</body></html>`

func handleLogin(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "mock_session", Value: "1", Path: "/"})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(loginHTML))
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		log.Printf("encode json: %v", err)
	}
}

func handleTenant(w http.ResponseWriter, r *http.Request) {
	// Stable fake account: first rune is skipped as employeeQuery in fetchWorkHourUserInfo.
	const userAccount = "M10001"
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"tenant": map[string]string{
				"userAccount": userAccount,
			},
		},
	})
}

func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	eq, _ := body["employeeQuery"].(string)
	hrID := int64(900001)
	if eq != "" {
		if n, err := strconv.ParseInt(strings.TrimSpace(eq), 10, 64); err == nil && n > 0 {
			hrID = n
		} else {
			hrID = int64(len(eq))*1000 + 42
		}
	}
	// Same shape as real /user-info: hrId may be string or number in production.
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      200,
		"messageCode": "attendance.response.ok",
		"messageText": "Success",
		"ok":          true,
		"data": map[string]any{
			"hrId":             strconv.FormatInt(hrID, 10),
			"attendanceScheme": "MOCK",
			"departmentDTO":    map[string]any{"departmentChineseName": "Mock Dept"},
			"shiftInformationDTO": map[string]any{
				"shiftNameZh": "China/Flex,Work:08:00-17:30,Rest 12:00-13:30/17:30-18:00,Core :09:00-17:30,Card: 05:00-04:59",
			},
			"isTopmanager": false,
		},
	})
}

func handleWorkHour(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	hrID := jsonNumberToInt64(body["hrId"])
	if hrID == 0 {
		hrID = jsonNumberToInt64(body["hrId"])
	}
	today := time.Now().Format("2006-01-02")
	rec := map[string]any{
		"id":                       1,
		"creationDate":             today + " 09:00:00",
		"createdBy":                "mock",
		"lastUpdateDate":           today + " 09:00:00",
		"lastUpdatedBy":            "mock",
		"originalId":               "mock-orig-1",
		"hrId":                     hrID,
		"dataSource":               "MOCK",
		"clockInReason":            "",
		"attendanceDate":           today,
		"clockInDate":              today,
		"clockInTime":              "09:00",
		"dayId":                    today,
		"clockingInSequenceNumber": 1,
		"earlyClockInTime":         "09:00",
		"lateClockInTime":          "21:00",
		"clockInType":              "NORMAL",
		"earlyClockInType":         "",
		"lateClockInType":          "",
		"attendanceStatus":         "Present",
		"minuteNumber":             "480",
		"hourNumber":               "8",
		"attendProcessId":          "",
		"workDay":                  "1",
		"attendanceStatusCode":     "OK",
		"earlyClockInReason":       "",
		"lateClockInReason":        "",
		"earlyClockTag":            "",
		"lateClockTag":             "",
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"records": []map[string]any{rec},
		},
	})
}

func jsonNumberToInt64(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case json.Number:
		i, _ := x.Int64()
		return i
	case string:
		i, err := strconv.ParseInt(strings.TrimSpace(x), 10, 64)
		if err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}

const pullRequestMockTotal = 47

func requestBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

func prState(n int) string {
	switch n % 3 {
	case 0:
		return "open"
	case 1:
		return "merged"
	default:
		return "closed"
	}
}

func pullRequestItemMap(n int, base string) map[string]any {
	target := "main"
	if n%4 == 0 {
		target = "develop"
	}
	now := time.Now().UTC().Format(time.RFC3339)
	return map[string]any{
		"id":           n,
		"number":       n,
		"url":          fmt.Sprintf("%s/pr-preview/%d", strings.TrimSuffix(base, "/"), n),
		"title":        fmt.Sprintf("feat: mock pull request #%d (%s)", n, prState(n)),
		"author":       fmt.Sprintf("dev%d", (n%9)+1),
		"sourceBranch": fmt.Sprintf("feature/pr-%d", n),
		"targetBranch": target,
		"state":        prState(n),
		"createdAt":    now,
		"updatedAt":    now,
	}
}

func handlePullRequestList(w http.ResponseWriter, r *http.Request) {
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
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}
	start := (page - 1) * pageSize
	if start >= total {
		writeJSON(w, http.StatusOK, map[string]any{
			"items":      []any{},
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": totalPages,
		})
		return
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	items := make([]map[string]any, 0, end-start)
	for i := start; i < end; i++ {
		items = append(items, pullRequestItemMap(i+1, base))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":      items,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	})
}

func handlePRPreview(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	n, err := strconv.Atoi(idStr)
	if err != nil || n < 1 || n > pullRequestMockTotal {
		http.NotFound(w, r)
		return
	}
	base := requestBaseURL(r)
	m := pullRequestItemMap(n, base)
	title, _ := m["title"].(string)
	author, _ := m["author"].(string)
	src, _ := m["sourceBranch"].(string)
	tgt, _ := m["targetBranch"].(string)
	st, _ := m["state"].(string)
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
		html.EscapeString(fmt.Sprint(m["url"])),
	)
}
