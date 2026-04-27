// Work-hour API mock: same routes and JSON shapes as the placeholder
// http://127.0.0.1:17890/id, /user-info, /work-hour (defaults in workhour_fetch.go).
//
// Handlers are split by area: middleware.go, login.go, workhour.go, pull_request.go, chat.go.
//
// Run from repo root:
//
//	go run ./cmd/mockserver
//
// Then point the app at it, for example:
//
//	export WORK_HOUR_TENANT_URL=http://127.0.0.1:17890/id
//	export WORK_HOUR_USER_INFO_URL=http://127.0.0.1:17890/user-info
//	export WORK_HOUR_WORKHOUR_URL=http://127.0.0.1:17890/work-hour
//	export PULL_REQUEST_LIST_URL=http://127.0.0.1:17890/pull-request
//	export PULL_REQUEST_TOTAL_URL=http://127.0.0.1:17890/pull-request/total
//
// OpenAI-compatible chat (for niumer AI local test):
//
//	export AI base URL in app to http://127.0.0.1:17890  →  POST /v1/chat/completions
//	(alias: POST /chat/completions — same handler). With JSON "stream": true, responds as text/event-stream (SSE).
package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:17890", "listen address")
	flag.Parse()

	mux := http.NewServeMux()
	registerHealthRoute(mux)
	registerLoginRoutes(mux)
	registerWorkHourRoutes(mux)
	registerPullRequestRoutes(mux)
	registerChatRoutes(mux)

	log.Printf("mockserver listening on http://%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, withCORS(logRequests(mux))))
}
