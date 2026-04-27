package main

import "net/http"

// loginHTML is picked so chromedp.WaitVisible succeeds in headless Chrome (non-zero
// layout box). Includes .foo and .search-total__num for WORK_HOUR_WAIT_CSS overrides.
const loginHTML = `<!doctype html><html lang="zh"><head><meta charset="utf-8">
<style>
.foo,.search-total__num{display:inline-block!important;width:32px!important;height:24px!important;
line-height:24px!important;visibility:visible!important;opacity:1!important;font-size:16px!important;}
</style></head><body>
<span class="foo search-total__num" id="mock-login-ready">0</span>
</body></html>`

func registerLoginRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /login", handleLogin)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "mock_session", Value: "1", Path: "/"})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(loginHTML))
}
