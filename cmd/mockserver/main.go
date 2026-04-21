// Work-hour API mock: same routes and JSON shapes as the placeholder
// http://127.0.0.1:17890/id, /hr-id, /work-hour (defaults in workhour_fetch.go).
//
// Run from repo root:
//
//	go run ./cmd/mockserver
//
// Then point the app at it, for example:
//
//	export WORK_HOUR_TENANT_URL=http://127.0.0.1:17890/id
//	export WORK_HOUR_HR_ID_URL=http://127.0.0.1:17890/hr-id
//	export WORK_HOUR_API_URL=http://127.0.0.1:17890/work-hour
package main

import (
	"encoding/json"
	"flag"
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
	mux.HandleFunc("POST /hr-id", handleHrID)
	mux.HandleFunc("POST /work-hour", handleWorkHour)

	log.Printf("mockserver listening on http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, logRequests(mux)))
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
	// Stable fake account: first rune is skipped as employeeQuery in fetchHrID.
	const userAccount = "M10001"
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"tenant": map[string]string{
				"userAccount": userAccount,
			},
		},
	})
}

func handleHrID(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"hrId": hrID,
		},
	})
}

func handleWorkHour(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	hrID := jsonNumberToInt64(body["hr_id"])
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
