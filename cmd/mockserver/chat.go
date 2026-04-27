package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func registerChatRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/chat/completions", handleChatCompletions)
	mux.HandleFunc("POST /chat/completions", handleChatCompletions)
}

// handleChatCompletions implements OpenAI-compatible chat mock (niumer AIChatStream).
func handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	var body struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		Stream bool `json:"stream"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{
				"message": "invalid JSON body",
				"type":    "invalid_request_error",
			},
		})
		return
	}
	lastUser := ""
	for i := len(body.Messages) - 1; i >= 0; i-- {
		if strings.EqualFold(strings.TrimSpace(body.Messages[i].Role), "user") {
			lastUser = body.Messages[i].Content
			break
		}
	}
	model := strings.TrimSpace(body.Model)
	if model == "" {
		model = "mock"
	}
	reply := buildMockChatReply(r.URL.Path, model, lastUser)
	if body.Stream {
		writeChatCompletionSSE(w, model, reply)
		return
	}
	now := time.Now().Unix()
	writeJSON(w, http.StatusOK, map[string]any{
		"id":      fmt.Sprintf("chatcmpl-mock-%d", now),
		"object":  "chat.completion",
		"created": now,
		"model":   model,
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]string{
					"role":    "assistant",
					"content": reply,
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]int{
			"prompt_tokens":     1,
			"completion_tokens": len(reply) / 4,
			"total_tokens":      10,
		},
	})
}

// buildMockChatReply returns a long assistant body so stream mode has many chunks
// (short echo + repeated paragraphs for local UX testing).
func buildMockChatReply(path, model, lastUser string) string {
	last := strings.TrimSpace(lastUser)
	if last == "" {
		last = "（未在 messages 里找到 user 文本）"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "「mockserver」%s · model=%q\n\n", path, model)
	fmt.Fprintf(&b, "### 回声\n\n%q\n\n---\n\n### 流式体感正文（假数据）\n\n", last)
	for i := 1; i <= 12; i++ {
		fmt.Fprintf(&b,
			"**[%d]** 这是刻意写长的占位段落：真实线上接口往往一口气推很多 token；"+
				"本地 mock 则把 Unicode 切成很小的 delta，并在两次 flush 之间 sleep 几毫秒，"+
				"这样桌面端 UI 更容易看出「字在往外冒」，而不是整屏瞬间铺满。"+
				"你可以一边滚动一边观察输入框上方的气泡是否在持续增长。\n\n", i)
	}
	b.WriteString("---\n\n（以上为 mock 假回复，与真实模型无关。）\n")
	return b.String()
}

func splitReplyForStream(reply string, runesPerChunk int) []string {
	if reply == "" {
		return nil
	}
	if runesPerChunk < 1 {
		runesPerChunk = 6
	}
	var chunks []string
	var b strings.Builder
	n := 0
	for _, r := range reply {
		b.WriteRune(r)
		n++
		if n >= runesPerChunk {
			chunks = append(chunks, b.String())
			b.Reset()
			n = 0
		}
	}
	if b.Len() > 0 {
		chunks = append(chunks, b.String())
	}
	return chunks
}

func writeChatCompletionSSE(w http.ResponseWriter, model, reply string) {
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	fl, ok := w.(http.Flusher)
	if !ok {
		return
	}
	id := fmt.Sprintf("chatcmpl-mock-%d", time.Now().UnixNano())
	created := time.Now().Unix()

	emit := func(v any) {
		raw, err := json.Marshal(v)
		if err != nil {
			return
		}
		_, _ = fmt.Fprintf(w, "data: %s\n\n", raw)
		fl.Flush()
	}

	emit(map[string]any{
		"id":      id,
		"object":  "chat.completion.chunk",
		"created": created,
		"model":   model,
		"choices": []any{
			map[string]any{"index": 0, "delta": map[string]string{"role": "assistant"}, "finish_reason": nil},
		},
	})
	// Small chunks + brief pause so the desktop client visibly “types” (dev only).
	const runesPerChunk = 3
	for _, piece := range splitReplyForStream(reply, runesPerChunk) {
		emit(map[string]any{
			"id":      id,
			"object":  "chat.completion.chunk",
			"created": created,
			"model":   model,
			"choices": []any{
				map[string]any{"index": 0, "delta": map[string]string{"content": piece}, "finish_reason": nil},
			},
		})
		time.Sleep(5 * time.Millisecond)
	}
	emit(map[string]any{
		"id":      id,
		"object":  "chat.completion.chunk",
		"created": created,
		"model":   model,
		"choices": []any{
			map[string]any{"index": 0, "delta": map[string]any{}, "finish_reason": "stop"},
		},
	})
	_, _ = fmt.Fprintf(w, "data: [DONE]\n\n")
	fl.Flush()
}
