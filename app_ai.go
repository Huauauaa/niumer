package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// AISettingsView is bound to the frontend for configuring OpenAI-compatible APIs.
type AISettingsView struct {
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey"`
	Model   string `json:"model"`
}

// AIChatMessage is one turn in POST /v1/chat/completions (OpenAI format).
type AIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GetAISettings returns stored AI endpoint settings (API key is included for local editing only).
func (a *App) GetAISettings() AISettingsView {
	c, err := readAppConfig()
	if err != nil {
		return AISettingsView{}
	}
	return AISettingsView{
		BaseURL: strings.TrimSpace(c.AIBaseURL),
		APIKey:  c.AIAPIKey,
		Model:   strings.TrimSpace(c.AIModel),
	}
}

// SetAISettings persists base URL, API key, and optional model to User/settings.json.
func (a *App) SetAISettings(baseURL, apiKey, model string) error {
	c, err := readAppConfig()
	if err != nil {
		c = appConfig{}
	}
	c.AIBaseURL = strings.TrimSpace(baseURL)
	c.AIAPIKey = strings.TrimSpace(apiKey)
	c.AIModel = strings.TrimSpace(model)
	return writeAppConfig(c)
}

// AIChatStream starts a streaming POST to {baseURL}/v1/chat/completions (stream: true).
// streamID must be a non-empty client-generated id so the UI can subscribe to events
// before any chunk arrives. Events:
//   - niumer:ai:delta  map[string]string{"streamId","text"}
//   - niumer:ai:done  streamID string
//   - niumer:ai:error map[string]string{"streamId","message"}
func (a *App) AIChatStream(streamID string, messages []AIChatMessage) error {
	streamID = strings.TrimSpace(streamID)
	if streamID == "" {
		return fmt.Errorf("streamId is empty")
	}
	if a.ctx == nil {
		return fmt.Errorf("app not ready")
	}
	c, err := readAppConfig()
	if err != nil {
		return err
	}
	base := strings.TrimRight(strings.TrimSpace(c.AIBaseURL), "/")
	if base == "" {
		return fmt.Errorf("AI base URL is empty")
	}
	key := strings.TrimSpace(c.AIAPIKey)
	if key == "" {
		return fmt.Errorf("API key is empty")
	}
	model := strings.TrimSpace(c.AIModel)
	if model == "" {
		model = "deepseek-chat"
	}
	if len(messages) == 0 {
		return fmt.Errorf("no messages")
	}
	msgs := append([]AIChatMessage(nil), messages...)
	go a.runAIChatStream(streamID, base, key, model, msgs)
	return nil
}

func (a *App) runAIChatStream(streamID, base, key, model string, messages []AIChatMessage) {
	emitErr := func(msg string) {
		runtime.EventsEmit(a.ctx, "niumer:ai:error", map[string]string{
			"streamId": streamID,
			"message":  msg,
		})
	}
	ctx := a.ctx
	body := map[string]any{
		"model":    model,
		"messages": messages,
		"stream":   true,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		emitErr(err.Error())
		return
	}
	url := base + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		emitErr(err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+key)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		emitErr(err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(b))
		if len(msg) > 800 {
			msg = msg[:800] + "…"
		}
		if msg == "" {
			msg = resp.Status
		}
		emitErr(fmt.Sprintf("API %s: %s", resp.Status, msg))
		return
	}

	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(ct, "text/event-stream") {
		a.parseAIChatSSE(streamID, resp.Body)
		return
	}

	// Fallback: non-stream JSON (older mocks or misconfigured Content-Type).
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		emitErr(err.Error())
		return
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		emitErr(fmt.Sprintf("invalid API JSON: %v", err))
		return
	}
	if len(parsed.Choices) == 0 {
		emitErr("empty choices in API response")
		return
	}
	out := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if out != "" {
		runtime.EventsEmit(a.ctx, "niumer:ai:delta", map[string]string{
			"streamId": streamID,
			"text":     out,
		})
	}
	runtime.EventsEmit(a.ctx, "niumer:ai:done", streamID)
}

func (a *App) parseAIChatSSE(streamID string, r io.Reader) {
	emitErr := func(msg string) {
		runtime.EventsEmit(a.ctx, "niumer:ai:error", map[string]string{
			"streamId": streamID,
			"message":  msg,
		})
	}
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			runtime.EventsEmit(a.ctx, "niumer:ai:done", streamID)
			return
		}
		var envelope struct {
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
			continue
		}
		if envelope.Error != nil && envelope.Error.Message != "" {
			emitErr(envelope.Error.Message)
			return
		}
		if len(envelope.Choices) == 0 {
			continue
		}
		piece := envelope.Choices[0].Delta.Content
		if piece != "" {
			runtime.EventsEmit(a.ctx, "niumer:ai:delta", map[string]string{
				"streamId": streamID,
				"text":     piece,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		emitErr(err.Error())
		return
	}
	runtime.EventsEmit(a.ctx, "niumer:ai:done", streamID)
}
