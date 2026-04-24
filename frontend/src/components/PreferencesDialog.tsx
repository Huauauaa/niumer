import { useEffect, useState } from "react";
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
  GetWorkHourDBPath,
  SetBlogWorkDir,
  SetJsonFormatterWorkDir,
  SetReminderDBPath,
  SetWorkHourDBPath,
} from "../../wailsjs/go/main/App";

type Props = {
  open: boolean;
  onClose: () => void;
  onSaved?: () => void;
};

export function PreferencesDialog({ open, onClose, onSaved }: Props) {
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

  useEffect(() => {
    if (!open) return;
    setError(null);
    void (async () => {
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
    })();
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onClose]);

  if (!open) return null;

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

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50"
      role="presentation"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        className="max-h-[min(100%-2rem,90vh)] w-[min(100%-2rem,640px)] overflow-y-auto rounded border border-[var(--vscode-border)] bg-[#252526] p-4 shadow-xl"
        role="dialog"
        aria-labelledby="prefs-title"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <h2
          id="prefs-title"
          className="m-0 mb-3 min-h-[28px] text-[13px] font-semibold text-[#cccccc]"
        >
          Preference
        </h2>

        <section className="mb-5">
          <p className="mb-2 text-[12px] leading-relaxed text-[#858585]">
            博客 Markdown
            仅保存在本机该目录；在此修改路径并保存后会写入应用配置。Blog files
            stay on disk only under the path below. Built-in default if unset:{" "}
            <span className="allow-select font-mono text-[#b5cea8]">
              {defaultBlogPath || "—"}
            </span>
          </p>
          <label className="mb-1 block text-[11px] uppercase text-[#858585]">
            Directory
          </label>
          <div className="flex gap-2">
            <input
              type="text"
              className="allow-select min-w-0 flex-1 rounded border border-[var(--vscode-border)] bg-[#3c3c3c] px-2 py-1.5 font-mono text-[12px] text-[#cccccc] focus:border-[#007fd4] focus:outline-none"
              value={blogPath}
              onChange={(e) => setBlogPath(e.target.value)}
              spellCheck={false}
            />
            <button
              type="button"
              className="shrink-0 rounded border border-[var(--vscode-border)] bg-[#3c3c3c] px-3 py-1.5 text-[12px] text-[#cccccc] hover:bg-[#454545]"
              onClick={() => void handleBrowseBlog()}
            >
              Browse…
            </button>
          </div>
        </section>

        <section className="mb-5">
          <p className="mb-2 text-[12px] leading-relaxed text-[#858585]">
            JSON 格式化器草稿保存在本机目录下的{" "}
            <span className="font-mono">draft.json</span>
            。macOS / Linux / Windows 默认均为用户「文档」下的{" "}
            <span className="font-mono">niumer-json-formatter</span>
            （与博客目录规则一致）。Built-in default if unset:{" "}
            <span className="allow-select font-mono text-[#b5cea8]">
              {defaultJsonFormatterPath || "—"}
            </span>
          </p>
          <label className="mb-1 block text-[11px] uppercase text-[#858585]">
            JSON formatter directory
          </label>
          <div className="flex gap-2">
            <input
              type="text"
              className="allow-select min-w-0 flex-1 rounded border border-[var(--vscode-border)] bg-[#3c3c3c] px-2 py-1.5 font-mono text-[12px] text-[#cccccc] focus:border-[#007fd4] focus:outline-none"
              value={jsonFormatterPath}
              onChange={(e) => setJsonFormatterPath(e.target.value)}
              spellCheck={false}
            />
            <button
              type="button"
              className="shrink-0 rounded border border-[var(--vscode-border)] bg-[#3c3c3c] px-3 py-1.5 text-[12px] text-[#cccccc] hover:bg-[#454545]"
              onClick={() => void handleBrowseJsonFormatter()}
            >
              Browse…
            </button>
          </div>
        </section>

        <section className="mb-5">
          <p className="mb-2 text-[12px] text-[#858585]">
            Work hour SQLite file path. Default:{" "}
            <span className="allow-select font-mono text-[#b5cea8]">
              {defaultWorkHourPath || "—"}
            </span>
            {" · "}
            Clear the field and save to use the default path.
          </p>
          <label className="mb-1 block text-[11px] uppercase text-[#858585]">
            Workhour SQLite file
          </label>
          <div className="flex gap-2">
            <input
              type="text"
              className="allow-select min-w-0 flex-1 rounded border border-[var(--vscode-border)] bg-[#3c3c3c] px-2 py-1.5 font-mono text-[12px] text-[#cccccc] focus:border-[#007fd4] focus:outline-none"
              value={workHourPath}
              onChange={(e) => setWorkHourPath(e.target.value)}
              spellCheck={false}
              placeholder={defaultWorkHourPath || "work_hour.db"}
            />
            <button
              type="button"
              className="shrink-0 rounded border border-[var(--vscode-border)] bg-[#3c3c3c] px-3 py-1.5 text-[12px] text-[#cccccc] hover:bg-[#454545]"
              onClick={() => void handleBrowseWorkHour()}
            >
              Browse…
            </button>
          </div>
        </section>

        <section className="mb-2">
          <p className="mb-2 text-[12px] text-[#858585]">
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
          <label className="mb-1 block text-[11px] uppercase text-[#858585]">
            Reminder SQLite file
          </label>
          <div className="flex gap-2">
            <input
              type="text"
              className="allow-select min-w-0 flex-1 rounded border border-[var(--vscode-border)] bg-[#3c3c3c] px-2 py-1.5 font-mono text-[12px] text-[#cccccc] focus:border-[#007fd4] focus:outline-none"
              value={reminderDbPath}
              onChange={(e) => setReminderDbPath(e.target.value)}
              spellCheck={false}
              placeholder={defaultReminderDbPath || "reminder.db"}
            />
            <button
              type="button"
              className="shrink-0 rounded border border-[var(--vscode-border)] bg-[#3c3c3c] px-3 py-1.5 text-[12px] text-[#cccccc] hover:bg-[#454545]"
              onClick={() => void handleBrowseReminderDb()}
            >
              Browse…
            </button>
          </div>
        </section>

        {error && <p className="mt-2 text-[12px] text-[#f48771]">{error}</p>}
        <div className="mt-4 flex justify-end gap-2">
          <button
            type="button"
            className="rounded px-3 py-1.5 text-[12px] text-[#cccccc] hover:bg-white/10"
            onClick={onClose}
          >
            Cancel
          </button>
          <button
            type="button"
            disabled={saving}
            className="rounded bg-[#0e639c] px-3 py-1.5 text-[12px] text-white hover:bg-[#1177bb] disabled:opacity-50"
            onClick={() => void handleSave()}
          >
            {saving ? "Saving…" : "Save"}
          </button>
        </div>
      </div>
    </div>
  );
}
