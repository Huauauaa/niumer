import { useCallback, useEffect, useState } from "react";
import {
  ChooseBlogWorkDir,
  ChooseJsonFormatterWorkDir,
  ChooseReminderDBPath,
  ChooseWorkHourDBPath,
  GetBlogWorkDir,
  GetDefaultBlogWorkDir,
  GetDefaultJsonFormatterWorkDir,
  GetDefaultReminderDBPath,
  GetDefaultWorkHourDBPath,
  GetJsonFormatterWorkDir,
  GetReminderDBPath,
  GetUITheme,
  GetUserSettingsFilePath,
  GetWorkHourDBPath,
  ReadUserSettingsJSON,
  SetBlogWorkDir,
  SetJsonFormatterWorkDir,
  SetReminderDBPath,
  SetUITheme,
  SetWorkHourDBPath,
  WriteUserSettingsJSON,
} from "../../wailsjs/go/main/App";
import { applyTheme, getStoredTheme, type UITheme } from "../theme";

type Props = {
  open: boolean;
  onClose: () => void;
  onSaved?: () => void;
};

type PrefsSection = "path" | "theme";

function IconOpenSettingsJson({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="currentColor"
      aria-hidden
    >
      <path d="M3 1.5h6.5L12 4v10.5a.5.5 0 01-.5.5H3A.5.5 0 012.5 14V2A.5.5 0 013 1.5zm1 .5v11h7V4.5H8.5V2H4zm5.5.8L10.2 4H9.5V2.8zM5 7h5v1H5V7zm0 2h4v1H5V9zm0 2h5v1H5v-1z" />
    </svg>
  );
}

export function PreferencesDialog({ open, onClose, onSaved }: Props) {
  const [section, setSection] = useState<PrefsSection>("path");
  const [uiTheme, setUiTheme] = useState<UITheme>(() => getStoredTheme());
  const [blogPath, setBlogPath] = useState("");
  const [defaultBlogPath, setDefaultBlogPath] = useState("");
  const [jsonFormatterPath, setJsonFormatterPath] = useState("");
  const [defaultJsonFormatterPath, setDefaultJsonFormatterPath] = useState("");
  const [workHourPath, setWorkHourPath] = useState("");
  const [defaultWorkHourPath, setDefaultWorkHourPath] = useState("");
  const [reminderDbPath, setReminderDbPath] = useState("");
  const [defaultReminderDbPath, setDefaultReminderDbPath] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const [jsonEditorOpen, setJsonEditorOpen] = useState(false);
  const [jsonText, setJsonText] = useState("");
  const [jsonLoading, setJsonLoading] = useState(false);
  const [userSettingsPath, setUserSettingsPath] = useState("");

  const loadPathState = useCallback(async () => {
    try {
      const [curBlog, defBlog, curJson, defJson, curWh, defWh, curRm, defRm] =
        await Promise.all([
          GetBlogWorkDir(),
          GetDefaultBlogWorkDir(),
          GetJsonFormatterWorkDir(),
          GetDefaultJsonFormatterWorkDir(),
          GetWorkHourDBPath(),
          GetDefaultWorkHourDBPath(),
          GetReminderDBPath(),
          GetDefaultReminderDBPath(),
        ]);
      setBlogPath(curBlog);
      setDefaultBlogPath(defBlog);
      setJsonFormatterPath(curJson);
      setDefaultJsonFormatterPath(defJson);
      setWorkHourPath(curWh);
      setDefaultWorkHourPath(defWh);
      setReminderDbPath(curRm);
      setDefaultReminderDbPath(defRm);
    } catch {
      setBlogPath("");
      setDefaultBlogPath("");
      setJsonFormatterPath("");
      setDefaultJsonFormatterPath("");
      setWorkHourPath("");
      setDefaultWorkHourPath("");
      setReminderDbPath("");
      setDefaultReminderDbPath("");
    }
  }, []);

  useEffect(() => {
    if (!open) return;
    setJsonEditorOpen(false);
    setSection("path");
    void GetUITheme()
      .then((t) => {
        if (t === "light" || t === "dark") setUiTheme(t);
        else setUiTheme(getStoredTheme());
      })
      .catch(() => setUiTheme(getStoredTheme()));
    setError(null);
    void loadPathState();
    void GetUserSettingsFilePath()
      .then(setUserSettingsPath)
      .catch(() => setUserSettingsPath(""));
  }, [open, loadPathState]);

  useEffect(() => {
    if (!open || !jsonEditorOpen) return;
    setJsonLoading(true);
    setError(null);
    void ReadUserSettingsJSON()
      .then((t) => setJsonText(t))
      .catch((e) => {
        setError(e instanceof Error ? e.message : String(e));
        setJsonText("");
      })
      .finally(() => setJsonLoading(false));
  }, [open, jsonEditorOpen]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== "Escape") return;
      if (jsonEditorOpen) {
        e.preventDefault();
        setJsonEditorOpen(false);
        setError(null);
        return;
      }
      onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onClose, jsonEditorOpen]);

  if (!open) return null;

  const navBtn = (id: PrefsSection, label: string) => {
    const sel = section === id;
    return (
      <button
        key={id}
        type="button"
        className={`w-full rounded px-3 py-2 text-left text-[13px] focus:outline-none focus-visible:ring-1 focus-visible:ring-[#007fd4] ${
          sel
            ? "bg-[var(--vscode-list-hover)] font-medium text-[var(--vscode-fg)]"
            : "text-[var(--vscode-fg-muted)] hover:bg-[var(--vscode-list-hover)] hover:text-[var(--vscode-fg)]"
        }`}
        onClick={() => setSection(id)}
      >
        {label}
      </button>
    );
  };

  const handleBrowseBlog = async () => {
    setError(null);
    try {
      const picked = await ChooseBlogWorkDir();
      if (picked) setBlogPath(picked);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  };

  const handleBrowseJsonFormatter = async () => {
    setError(null);
    try {
      const picked = await ChooseJsonFormatterWorkDir();
      if (picked) setJsonFormatterPath(picked);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  };

  const handleBrowseWorkHour = async () => {
    setError(null);
    try {
      const picked = await ChooseWorkHourDBPath();
      if (picked) setWorkHourPath(picked);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  };

  const handleBrowseReminderDb = async () => {
    setError(null);
    try {
      const picked = await ChooseReminderDBPath();
      if (picked) setReminderDbPath(picked);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  };

  const handleSave = async () => {
    const b = blogPath.trim();
    if (!b) {
      setError("Blog directory cannot be empty.");
      return;
    }
    const j = jsonFormatterPath.trim();
    if (!j) {
      setError("JSON formatter directory cannot be empty.");
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await SetBlogWorkDir(b);
      await SetJsonFormatterWorkDir(j);
      await SetWorkHourDBPath(workHourPath.trim());
      await SetReminderDBPath(reminderDbPath.trim());
      onSaved?.();
      onClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setSaving(false);
    }
  };

  const handleSaveJson = async () => {
    setSaving(true);
    setError(null);
    try {
      await WriteUserSettingsJSON(jsonText);
      onSaved?.();
      await loadPathState();
      void GetUITheme()
        .then((t) => {
          if (t === "light" || t === "dark") applyTheme(t);
        })
        .catch(() => {});
      setJsonEditorOpen(false);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setSaving(false);
    }
  };

  const handleFooterCancel = () => {
    if (jsonEditorOpen) {
      setJsonEditorOpen(false);
      setError(null);
      return;
    }
    onClose();
  };

  const handleFooterSave = () => {
    if (jsonEditorOpen) void handleSaveJson();
    else void handleSave();
  };

  const setTheme = (t: UITheme) => {
    applyTheme(t);
    setUiTheme(t);
    void SetUITheme(t).catch(() => {});
  };

  const themeChoiceClass = (t: UITheme) =>
    `flex w-full items-center justify-between rounded border px-3 py-2.5 text-left text-[13px] transition-colors focus:outline-none focus-visible:ring-1 focus-visible:ring-[#007fd4] ${
      uiTheme === t
        ? "border-[#007fd4] bg-[var(--vscode-list-hover)] text-[var(--vscode-fg)]"
        : "border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] text-[var(--vscode-fg)] hover:border-[#007fd4]/60"
    }`;

  const modalWidthClass = jsonEditorOpen
    ? "w-[min(100%-2rem,880px)]"
    : "w-[min(100%-2rem,760px)]";

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50"
      role="presentation"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        className={`flex max-h-[min(100%-2rem,90vh)] ${modalWidthClass} min-h-0 flex-col overflow-hidden rounded border border-[var(--vscode-border)] bg-[var(--vscode-dialog-bg)] shadow-xl`}
        role="dialog"
        aria-labelledby="prefs-title"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className="flex shrink-0 items-center justify-between gap-2 border-b border-[var(--vscode-border)] px-4 py-2.5">
          <h2
            id="prefs-title"
            className="m-0 min-w-0 flex-1 text-[13px] font-semibold text-[var(--vscode-fg)]"
          >
            {jsonEditorOpen ? "settings.json" : "Preference"}
          </h2>
          {!jsonEditorOpen ? (
            <button
              type="button"
              className="shrink-0 rounded p-1 text-[var(--vscode-fg-muted)] hover:bg-[var(--vscode-menu-hover)] hover:text-[var(--vscode-fg)] focus:outline-none focus-visible:ring-1 focus-visible:ring-[#007fd4]"
              title="Open Settings (JSON)"
              aria-label="Open Settings (JSON)"
              onClick={() => {
                setError(null);
                setJsonEditorOpen(true);
              }}
            >
              <IconOpenSettingsJson />
            </button>
          ) : (
            <button
              type="button"
              className="shrink-0 rounded px-2 py-1 text-[12px] text-[var(--vscode-fg-muted)] hover:bg-[var(--vscode-menu-hover)] hover:text-[var(--vscode-fg)] focus:outline-none focus-visible:ring-1 focus-visible:ring-[#007fd4]"
              title="Back to form"
              aria-label="Back to form"
              onClick={() => {
                setJsonEditorOpen(false);
                setError(null);
              }}
            >
              Form
            </button>
          )}
        </div>

        {jsonEditorOpen ? (
          <div className="flex min-h-0 min-w-0 flex-1 flex-col">
            {userSettingsPath ? (
              <div className="shrink-0 border-b border-[var(--vscode-border)] px-4 py-1.5 font-mono text-[11px] text-[var(--vscode-fg-muted)]">
                <span className="allow-select break-all">{userSettingsPath}</span>
              </div>
            ) : null}
            <div className="min-h-0 flex-1 p-2">
              {jsonLoading ? (
                <div className="px-2 py-4 text-[12px] text-[var(--vscode-fg-muted)]">
                  Loading…
                </div>
              ) : (
                <textarea
                  className="allow-select h-full min-h-[280px] w-full resize-none rounded border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] p-3 font-mono text-[12px] leading-relaxed text-[var(--vscode-fg)] focus:border-[#007fd4] focus:outline-none"
                  spellCheck={false}
                  value={jsonText}
                  onChange={(e) => setJsonText(e.target.value)}
                  aria-label="User settings JSON"
                />
              )}
            </div>
            {error && (
              <p className="shrink-0 px-4 pb-1 text-[12px] text-[#f48771]">
                {error}
              </p>
            )}
            <div className="shrink-0 border-t border-[var(--vscode-border)] px-4 py-3">
              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  className="rounded px-3 py-1.5 text-[12px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-menu-hover)]"
                  onClick={handleFooterCancel}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  disabled={saving || jsonLoading}
                  className="rounded bg-[#0e639c] px-3 py-1.5 text-[12px] text-white hover:bg-[#1177bb] disabled:opacity-50"
                  onClick={() => void handleFooterSave()}
                >
                  {saving ? "Saving…" : "Save"}
                </button>
              </div>
            </div>
          </div>
        ) : (
          <div className="flex min-h-0 flex-1">
            <nav
              className="flex w-40 shrink-0 flex-col gap-0.5 border-r border-[var(--vscode-border)] p-2"
              aria-label="Preference sections"
            >
              {navBtn("path", "Path")}
              {navBtn("theme", "Theme")}
            </nav>

            <div className="flex min-h-0 min-w-0 flex-1 flex-col">
              <div className="min-h-0 flex-1 overflow-y-auto p-4">
                {section === "path" ? (
                  <>
                    <section className="mb-5">
                      <p className="mb-2 text-[12px] leading-relaxed text-[var(--vscode-fg-muted)]">
                        博客 Markdown
                        仅保存在本机该目录；在此修改路径并保存后会写入应用配置。Blog
                        files stay on disk only under the path below. Built-in
                        default if unset:{" "}
                        <span className="allow-select font-mono text-[#b5cea8]">
                          {defaultBlogPath || "—"}
                        </span>
                      </p>
                      <label className="mb-1 block text-[11px] uppercase text-[var(--vscode-fg-muted)]">
                        Directory
                      </label>
                      <div className="flex gap-2">
                        <input
                          type="text"
                          className="allow-select min-w-0 flex-1 rounded border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-2 py-1.5 font-mono text-[12px] text-[var(--vscode-fg)] focus:border-[#007fd4] focus:outline-none"
                          value={blogPath}
                          onChange={(e) => setBlogPath(e.target.value)}
                          spellCheck={false}
                        />
                        <button
                          type="button"
                          className="shrink-0 rounded border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-3 py-1.5 text-[12px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-button-hover)]"
                          onClick={() => void handleBrowseBlog()}
                        >
                          Browse…
                        </button>
                      </div>
                    </section>

                    <section className="mb-5">
                      <p className="mb-2 text-[12px] leading-relaxed text-[var(--vscode-fg-muted)]">
                        JSON 格式化器草稿保存在本机目录下的{" "}
                        <span className="font-mono">draft.json</span>
                        。macOS / Linux / Windows 默认均为用户「文档」下的{" "}
                        <span className="font-mono">niumer-json-formatter</span>
                        （与博客目录规则一致）。Built-in default if unset:{" "}
                        <span className="allow-select font-mono text-[#b5cea8]">
                          {defaultJsonFormatterPath || "—"}
                        </span>
                      </p>
                      <label className="mb-1 block text-[11px] uppercase text-[var(--vscode-fg-muted)]">
                        JSON formatter directory
                      </label>
                      <div className="flex gap-2">
                        <input
                          type="text"
                          className="allow-select min-w-0 flex-1 rounded border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-2 py-1.5 font-mono text-[12px] text-[var(--vscode-fg)] focus:border-[#007fd4] focus:outline-none"
                          value={jsonFormatterPath}
                          onChange={(e) => setJsonFormatterPath(e.target.value)}
                          spellCheck={false}
                        />
                        <button
                          type="button"
                          className="shrink-0 rounded border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-3 py-1.5 text-[12px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-button-hover)]"
                          onClick={() => void handleBrowseJsonFormatter()}
                        >
                          Browse…
                        </button>
                      </div>
                    </section>

                    <section className="mb-5">
                      <p className="mb-2 text-[12px] text-[var(--vscode-fg-muted)]">
                        Work hour SQLite file path. Default:{" "}
                        <span className="allow-select font-mono text-[#b5cea8]">
                          {defaultWorkHourPath || "—"}
                        </span>
                        {" · "}
                        Clear the field and save to use the default path.
                      </p>
                      <label className="mb-1 block text-[11px] uppercase text-[var(--vscode-fg-muted)]">
                        Workhour SQLite file
                      </label>
                      <div className="flex gap-2">
                        <input
                          type="text"
                          className="allow-select min-w-0 flex-1 rounded border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-2 py-1.5 font-mono text-[12px] text-[var(--vscode-fg)] focus:border-[#007fd4] focus:outline-none"
                          value={workHourPath}
                          onChange={(e) => setWorkHourPath(e.target.value)}
                          spellCheck={false}
                          placeholder={defaultWorkHourPath || "work_hour.db"}
                        />
                        <button
                          type="button"
                          className="shrink-0 rounded border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-3 py-1.5 text-[12px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-button-hover)]"
                          onClick={() => void handleBrowseWorkHour()}
                        >
                          Browse…
                        </button>
                      </div>
                    </section>

                    <section className="mb-2">
                      <p className="mb-2 text-[12px] text-[var(--vscode-fg-muted)]">
                        个人提醒数据保存在 SQLite 文件（表{" "}
                        <span className="font-mono">custom_reminders</span>
                        ）。默认与工时库同目录下的{" "}
                        <span className="font-mono">reminder.db</span>。Default:{" "}
                        <span className="allow-select font-mono text-[#b5cea8]">
                          {defaultReminderDbPath || "—"}
                        </span>
                        {" · "}
                        清空路径并保存则恢复默认位置。
                      </p>
                      <label className="mb-1 block text-[11px] uppercase text-[var(--vscode-fg-muted)]">
                        Reminder SQLite file
                      </label>
                      <div className="flex gap-2">
                        <input
                          type="text"
                          className="allow-select min-w-0 flex-1 rounded border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-2 py-1.5 font-mono text-[12px] text-[var(--vscode-fg)] focus:border-[#007fd4] focus:outline-none"
                          value={reminderDbPath}
                          onChange={(e) => setReminderDbPath(e.target.value)}
                          spellCheck={false}
                          placeholder={defaultReminderDbPath || "reminder.db"}
                        />
                        <button
                          type="button"
                          className="shrink-0 rounded border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-3 py-1.5 text-[12px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-button-hover)]"
                          onClick={() => void handleBrowseReminderDb()}
                        >
                          Browse…
                        </button>
                      </div>
                    </section>
                  </>
                ) : (
                  <div>
                    <p className="mb-3 text-[12px] leading-relaxed text-[var(--vscode-fg-muted)]">
                      选择窗口配色（Dark / Light）。选择后立即生效，并写入{" "}
                      <span className="font-mono">User/settings.json</span> 中的{" "}
                      <span className="font-mono">theme</span> 字段；也可在 JSON
                      视图中直接编辑该字段。
                    </p>
                    <div className="flex max-w-md flex-col gap-2">
                      <button
                        type="button"
                        className={themeChoiceClass("dark")}
                        onClick={() => setTheme("dark")}
                      >
                        <span>Dark</span>
                        {uiTheme === "dark" ? (
                          <span className="text-[11px] text-[#007fd4]">✓</span>
                        ) : (
                          <span className="w-3" />
                        )}
                      </button>
                      <button
                        type="button"
                        className={themeChoiceClass("light")}
                        onClick={() => setTheme("light")}
                      >
                        <span>Light</span>
                        {uiTheme === "light" ? (
                          <span className="text-[11px] text-[#007fd4]">✓</span>
                        ) : (
                          <span className="w-3" />
                        )}
                      </button>
                    </div>
                  </div>
                )}

                {error && (
                  <p className="mt-2 text-[12px] text-[#f48771]">{error}</p>
                )}
              </div>

              <div className="shrink-0 border-t border-[var(--vscode-border)] px-4 py-3">
                <div className="flex justify-end gap-2">
                  <button
                    type="button"
                    className="rounded px-3 py-1.5 text-[12px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-menu-hover)]"
                    onClick={handleFooterCancel}
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    disabled={saving}
                    className="rounded bg-[#0e639c] px-3 py-1.5 text-[12px] text-white hover:bg-[#1177bb] disabled:opacity-50"
                    onClick={() => void handleFooterSave()}
                  >
                    {saving ? "Saving…" : "Save"}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
