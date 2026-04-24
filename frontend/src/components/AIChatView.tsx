import { useCallback, useEffect, useRef, useState } from "react";
import { AIChatStream, GetAISettings, SetAISettings } from "../../wailsjs/go/main/App";
import { EventsOn } from "../../wailsjs/runtime/runtime";

type Phase = "loading" | "setup" | "chat";

type Msg = {
  id: string;
  role: "user" | "assistant";
  content: string;
};

function uid() {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

type Props = {
  /** Increment to clear the current conversation. */
  resetKey: number;
  /** Increment (from sidebar) to open the connection form. */
  settingsNonce: number;
};

export function AIChatView({ resetKey, settingsNonce }: Props) {
  const [phase, setPhase] = useState<Phase>("loading");
  const [baseUrl, setBaseUrl] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [model, setModel] = useState("deepseek-chat");
  const [messages, setMessages] = useState<Msg[]>([]);
  const [input, setInput] = useState("");
  const [sending, setSending] = useState(false);
  const [setupSaving, setSetupSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  /** 侧栏「连接设置」：已配置时在对话上方弹出表单，不离开聊天页 */
  const [connectionDialogOpen, setConnectionDialogOpen] = useState(false);
  const listRef = useRef<HTMLDivElement>(null);
  const connectionUrlInputRef = useRef<HTMLInputElement>(null);
  const prevSettingsNonce = useRef(0);
  /** Matches `streamId` in Wails events for the active request. */
  const currentStreamIdRef = useRef<string | null>(null);
  /** Assistant bubble id receiving streamed tokens. */
  const assistantStreamMsgIdRef = useRef<string | null>(null);

  useEffect(() => {
    const offDelta = EventsOn("niumer:ai:delta", (...args: unknown[]) => {
      const raw = args[0];
      if (!raw || typeof raw !== "object") return;
      const o = raw as Record<string, string>;
      const streamId = o.streamId ?? "";
      const text = o.text ?? "";
      if (!streamId || text === "") return;
      if (streamId !== currentStreamIdRef.current) return;
      const aid = assistantStreamMsgIdRef.current;
      if (!aid) return;
      setMessages((prev) =>
        prev.map((m) => (m.id === aid ? { ...m, content: m.content + text } : m)),
      );
    });
    const offDone = EventsOn("niumer:ai:done", (...args: unknown[]) => {
      const sid = typeof args[0] === "string" ? args[0] : "";
      if (!sid || sid !== currentStreamIdRef.current) return;
      assistantStreamMsgIdRef.current = null;
      currentStreamIdRef.current = null;
      setSending(false);
    });
    const offErr = EventsOn("niumer:ai:error", (...args: unknown[]) => {
      const raw = args[0];
      if (!raw || typeof raw !== "object") return;
      const o = raw as Record<string, string>;
      const sid = o.streamId ?? "";
      if (!sid || sid !== currentStreamIdRef.current) return;
      setError(o.message ?? "Unknown error");
      assistantStreamMsgIdRef.current = null;
      currentStreamIdRef.current = null;
      setSending(false);
    });
    return () => {
      offDelta();
      offDone();
      offErr();
    };
  }, []);

  const applyAISettingsFromServer = useCallback(async (): Promise<boolean> => {
    try {
      const s = await GetAISettings();
      setBaseUrl((s.baseUrl ?? "").trim());
      setApiKey(s.apiKey ?? "");
      setModel((s.model ?? "").trim() || "deepseek-chat");
      return Boolean((s.baseUrl ?? "").trim() && (s.apiKey ?? "").trim());
    } catch {
      setBaseUrl("");
      setApiKey("");
      setModel("deepseek-chat");
      return false;
    }
  }, []);

  const loadSettings = useCallback(async () => {
    const ok = await applyAISettingsFromServer();
    setPhase(ok ? "chat" : "setup");
  }, [applyAISettingsFromServer]);

  useEffect(() => {
    void loadSettings();
  }, [loadSettings]);

  useEffect(() => {
    setMessages([]);
    setInput("");
    setError(null);
    setConnectionDialogOpen(false);
  }, [resetKey]);

  useEffect(() => {
    if (settingsNonce <= prevSettingsNonce.current) return;
    prevSettingsNonce.current = settingsNonce;
    setError(null);
    void (async () => {
      const ok = await applyAISettingsFromServer();
      if (ok) {
        setConnectionDialogOpen(true);
      } else {
        setPhase("setup");
      }
    })();
  }, [settingsNonce, applyAISettingsFromServer]);

  useEffect(() => {
    if (!connectionDialogOpen) return;
    const t = window.requestAnimationFrame(() => {
      connectionUrlInputRef.current?.focus();
    });
    return () => window.cancelAnimationFrame(t);
  }, [connectionDialogOpen]);

  useEffect(() => {
    if (!connectionDialogOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        setConnectionDialogOpen(false);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [connectionDialogOpen]);

  useEffect(() => {
    if (phase !== "chat") return;
    const el = listRef.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  }, [messages, phase, sending]);

  const saveAIConnection = useCallback(async (): Promise<boolean> => {
    const b = baseUrl.trim();
    const k = apiKey.trim();
    if (!b || !k) {
      setError("请填写 Base URL 与 API Key。");
      return false;
    }
    setSetupSaving(true);
    setError(null);
    try {
      await SetAISettings(b, k, model.trim() || "deepseek-chat");
      return true;
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
      return false;
    } finally {
      setSetupSaving(false);
    }
  }, [baseUrl, apiKey, model]);

  const handleSaveSetup = async () => {
    if (await saveAIConnection()) setPhase("chat");
  };

  const handleSaveConnectionModal = async () => {
    if (await saveAIConnection()) setConnectionDialogOpen(false);
  };

  const handleSend = async () => {
    const text = input.trim();
    if (!text || sending) return;
    const userMsg: Msg = { id: uid(), role: "user", content: text };
    const assistantId = uid();
    const streamId = uid();
    assistantStreamMsgIdRef.current = assistantId;
    currentStreamIdRef.current = streamId;
    setInput("");
    setError(null);
    setMessages((prev) => [
      ...prev,
      userMsg,
      { id: assistantId, role: "assistant", content: "" },
    ]);
    setSending(true);
    const history = [...messages, userMsg].map((m) => ({
      role: m.role,
      content: m.content,
    }));
    try {
      await AIChatStream(streamId, history);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
      assistantStreamMsgIdRef.current = null;
      currentStreamIdRef.current = null;
      setMessages((prev) => prev.filter((m) => m.id !== assistantId));
      setSending(false);
    }
  };

  if (phase === "loading") {
    return (
      <div
        className="flex min-h-0 flex-1 items-center justify-center text-[13px] text-[var(--vscode-fg-muted)]"
        style={{ background: "var(--vscode-editor-bg)" }}
      >
        加载中…
      </div>
    );
  }

  if (phase === "setup") {
    return (
      <div
        className="flex min-h-0 flex-1 flex-col items-center justify-center px-4 py-8"
        style={{ background: "var(--vscode-editor-bg)" }}
      >
        <div className="w-full max-w-md rounded-2xl border border-[var(--vscode-border)] bg-[var(--vscode-sideBar-bg)] px-6 py-7 shadow-lg">
          <h1 className="mb-1 text-center text-[18px] font-semibold tracking-tight text-[var(--vscode-fg)]">
            AI 对话
          </h1>
          <p className="mb-6 text-center text-[12px] leading-relaxed text-[var(--vscode-fg-muted)]">
            配置 OpenAI 兼容接口（如 DeepSeek：
            <span className="font-mono text-[11px]"> https://api.deepseek.com</span>
            ），密钥保存在本机{" "}
            <code className="rounded bg-[var(--vscode-textBlockQuote-background)] px-1 py-0.5 font-mono text-[10px]">
              User/settings.json
            </code>
            。
          </p>
          {error ? (
            <div className="mb-4 rounded-lg border border-[#f4877144] bg-[#f4877114] px-3 py-2 text-[12px] text-[#f48771]">
              {error}
            </div>
          ) : null}
          <label className="mb-1 block text-[11px] font-medium uppercase tracking-wide text-[var(--vscode-fg-muted)]">
            Base URL
          </label>
          <input
            className="mb-4 w-full rounded-lg border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-3 py-2.5 text-[13px] text-[var(--vscode-fg)] outline-none ring-[var(--vscode-focus-ring)] focus:ring-1"
            placeholder="https://api.deepseek.com"
            value={baseUrl}
            onChange={(e) => setBaseUrl(e.target.value)}
            autoComplete="off"
          />
          <label className="mb-1 block text-[11px] font-medium uppercase tracking-wide text-[var(--vscode-fg-muted)]">
            API Key
          </label>
          <input
            className="mb-4 w-full rounded-lg border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-3 py-2.5 text-[13px] text-[var(--vscode-fg)] outline-none ring-[var(--vscode-focus-ring)] focus:ring-1"
            placeholder="sk-…"
            type="password"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            autoComplete="off"
          />
          <label className="mb-1 block text-[11px] font-medium uppercase tracking-wide text-[var(--vscode-fg-muted)]">
            模型（可选）
          </label>
          <input
            className="mb-6 w-full rounded-lg border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-3 py-2.5 text-[13px] text-[var(--vscode-fg)] outline-none ring-[var(--vscode-focus-ring)] focus:ring-1"
            placeholder="deepseek-chat"
            value={model}
            onChange={(e) => setModel(e.target.value)}
            autoComplete="off"
          />
          <div className="flex flex-wrap gap-2">
            {baseUrl.trim() && apiKey.trim() ? (
              <button
                type="button"
                className="rounded-xl border border-[var(--vscode-border)] px-4 py-2.5 text-[13px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-list-hover)]"
                onClick={() => {
                  setError(null);
                  setPhase("chat");
                }}
              >
                返回对话
              </button>
            ) : null}
            <button
              type="button"
              className="min-w-[140px] flex-1 rounded-xl bg-[#4a9eff] py-2.5 text-[13px] font-medium text-white transition hover:bg-[#3d8eef] disabled:opacity-50"
              disabled={setupSaving}
              onClick={() => void handleSaveSetup()}
            >
              {setupSaving ? "保存中…" : "保存并开始"}
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div
      className="relative flex min-h-0 flex-1 flex-col"
      style={{ background: "var(--vscode-editor-bg)" }}
    >
      {connectionDialogOpen ? (
        <div
          className="absolute inset-0 z-20 flex items-center justify-center bg-black/45 px-4 py-8"
          role="presentation"
          onMouseDown={(e) => {
            if (e.target === e.currentTarget) setConnectionDialogOpen(false);
          }}
        >
          <div
            role="dialog"
            aria-labelledby="ai-connection-title"
            className="w-full max-w-md rounded-2xl border border-[var(--vscode-border)] bg-[var(--vscode-sideBar-bg)] px-6 py-6 shadow-xl"
            onMouseDown={(e) => e.stopPropagation()}
          >
            <h2
              id="ai-connection-title"
              className="mb-1 text-center text-[16px] font-semibold text-[var(--vscode-fg)]"
            >
              连接设置
            </h2>
            <p className="mb-5 text-center text-[12px] leading-relaxed text-[var(--vscode-fg-muted)]">
              Base URL、API Key 与模型写入本机{" "}
              <code className="rounded bg-[var(--vscode-textBlockQuote-background)] px-1 py-0.5 font-mono text-[10px]">
                User/settings.json
              </code>
            </p>
            {error ? (
              <div className="mb-4 rounded-lg border border-[#f4877144] bg-[#f4877114] px-3 py-2 text-[12px] text-[#f48771]">
                {error}
              </div>
            ) : null}
            <label className="mb-1 block text-[11px] font-medium uppercase tracking-wide text-[var(--vscode-fg-muted)]">
              Base URL
            </label>
            <input
              ref={connectionUrlInputRef}
              className="mb-3 w-full rounded-lg border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-3 py-2.5 text-[13px] text-[var(--vscode-fg)] outline-none ring-[var(--vscode-focus-ring)] focus:ring-1"
              placeholder="https://api.deepseek.com"
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
              autoComplete="off"
            />
            <label className="mb-1 block text-[11px] font-medium uppercase tracking-wide text-[var(--vscode-fg-muted)]">
              API Key
            </label>
            <input
              className="mb-3 w-full rounded-lg border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-3 py-2.5 text-[13px] text-[var(--vscode-fg)] outline-none ring-[var(--vscode-focus-ring)] focus:ring-1"
              placeholder="sk-…"
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              autoComplete="off"
            />
            <label className="mb-1 block text-[11px] font-medium uppercase tracking-wide text-[var(--vscode-fg-muted)]">
              模型
            </label>
            <input
              className="mb-5 w-full rounded-lg border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-3 py-2.5 text-[13px] text-[var(--vscode-fg)] outline-none ring-[var(--vscode-focus-ring)] focus:ring-1"
              placeholder="deepseek-chat"
              value={model}
              onChange={(e) => setModel(e.target.value)}
              autoComplete="off"
            />
            <div className="flex justify-end gap-2">
              <button
                type="button"
                className="rounded-lg border border-[var(--vscode-border)] px-4 py-2 text-[13px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-list-hover)]"
                onClick={() => {
                  setError(null);
                  setConnectionDialogOpen(false);
                  void applyAISettingsFromServer();
                }}
              >
                取消
              </button>
              <button
                type="button"
                className="rounded-lg bg-[#4a9eff] px-4 py-2 text-[13px] font-medium text-white hover:bg-[#3d8eef] disabled:opacity-50"
                disabled={setupSaving}
                onClick={() => void handleSaveConnectionModal()}
              >
                {setupSaving ? "保存中…" : "保存"}
              </button>
            </div>
          </div>
        </div>
      ) : null}
      <div
        ref={listRef}
        className="min-h-0 flex-1 overflow-y-auto px-4 py-6"
      >
        <div className="mx-auto flex max-w-3xl flex-col gap-5">
          {messages.length === 0 ? (
            <div className="py-16 text-center text-[14px] text-[var(--vscode-fg-muted)]">
              开始与 AI 对话。支持 OpenAI 兼容的{" "}
              <code className="rounded px-1 font-mono text-[12px]">/v1/chat/completions</code>
              。
            </div>
          ) : null}
          {messages.map((m) => (
            <div
              key={m.id}
              className={`flex ${m.role === "user" ? "justify-end" : "justify-start"}`}
            >
              <div
                className={`max-w-[85%] rounded-2xl px-4 py-2.5 text-[14px] leading-relaxed ${
                  m.role === "user"
                    ? "bg-[#4a9eff] text-white"
                    : "border border-[var(--vscode-border)] bg-[var(--vscode-sideBar-bg)] text-[var(--vscode-editor-fg)]"
                }`}
              >
                <div className="whitespace-pre-wrap break-words">{m.content}</div>
              </div>
            </div>
          ))}
          {sending ? (
            <div className="flex justify-start">
              <div className="rounded-2xl border border-[var(--vscode-border)] bg-[var(--vscode-sideBar-bg)] px-4 py-2.5 text-[13px] text-[var(--vscode-fg-muted)]">
                思考中…
              </div>
            </div>
          ) : null}
        </div>
      </div>
      {error ? (
        <div className="shrink-0 border-t border-[var(--vscode-border)] bg-[#f4877114] px-4 py-2 text-center text-[12px] text-[#f48771]">
          {error}
        </div>
      ) : null}
      <div className="shrink-0 border-t border-[var(--vscode-border)] px-4 pb-4 pt-3">
        <div className="mx-auto flex max-w-3xl gap-2">
          <textarea
            className="min-h-[48px] flex-1 resize-none rounded-2xl border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-4 py-3 text-[14px] text-[var(--vscode-fg)] outline-none ring-[var(--vscode-focus-ring)] focus:ring-1"
            rows={2}
            placeholder="给 AI 发送消息…"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key !== "Enter" || e.shiftKey) return;
              if (e.nativeEvent.isComposing) return;
              e.preventDefault();
              void handleSend();
            }}
          />
          <button
            type="button"
            className="self-end rounded-2xl bg-[#4a9eff] px-5 py-3 text-[14px] font-medium text-white transition hover:bg-[#3d8eef] disabled:opacity-40"
            disabled={sending || !input.trim()}
            onClick={() => void handleSend()}
          >
            发送
          </button>
        </div>
      </div>
    </div>
  );
}
