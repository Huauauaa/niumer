import { Fragment, useEffect, useMemo, useState } from "react";
import type { CustomReminder } from "../types/reminder";

type Props = {
  items: CustomReminder[];
  selectedId: string | null;
  onSelect: (id: string | null) => void;
  onAdd: (name: string, date: string) => Promise<void>;
  onUpdate: (id: string, name: string, date: string) => Promise<void>;
  onDelete: (row: CustomReminder) => Promise<void>;
};

const rowClass =
  "group flex w-full cursor-pointer items-center gap-1 rounded px-2 py-1.5 text-left text-[12px] text-[#cccccc] hover:bg-[var(--vscode-list-hover)]";
const selectedRow = "bg-[#37373d] hover:bg-[#37373d]";

export function ReminderCustomSidebar({
  items,
  selectedId,
  onSelect,
  onAdd,
  onUpdate,
  onDelete,
}: Props) {
  const [editingId, setEditingId] = useState<string | null>(null);
  const [draftName, setDraftName] = useState("");
  const [draftDate, setDraftDate] = useState("");
  const [newName, setNewName] = useState("");
  const [newDate, setNewDate] = useState("");
  const [busy, setBusy] = useState(false);
  /** Wails webview 中 `window.confirm` 常不可用，用内置确认层代替 */
  const [pendingDelete, setPendingDelete] = useState<CustomReminder | null>(
    null,
  );

  const sorted = useMemo(
    () => [...items].sort((a, b) => a.date.localeCompare(b.date)),
    [items],
  );

  useEffect(() => {
    if (!pendingDelete) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        setPendingDelete(null);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [pendingDelete]);

  const startEdit = (r: CustomReminder) => {
    setEditingId(r.id);
    setDraftName(r.name);
    setDraftDate(r.date);
  };

  const commitEdit = async () => {
    if (!editingId) return;
    const name = draftName.trim();
    if (!name) {
      window.alert("名称不能为空。");
      return;
    }
    if (!/^\d{4}-\d{2}-\d{2}$/.test(draftDate)) {
      window.alert("请选择有效日期。");
      return;
    }
    setBusy(true);
    try {
      await onUpdate(editingId, name, draftDate);
      setEditingId(null);
    } catch (e) {
      window.alert(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  };

  const cancelEdit = () => setEditingId(null);

  const requestDelete = (row: CustomReminder) => {
    setPendingDelete(row);
  };

  const cancelDelete = () => setPendingDelete(null);

  const executeConfirmedDelete = async () => {
    const row = pendingDelete;
    if (!row) return;
    setPendingDelete(null);
    setBusy(true);
    try {
      await onDelete(row);
      if (editingId === row.id) setEditingId(null);
    } catch (e) {
      window.alert(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  };

  const add = async () => {
    const name = newName.trim();
    if (!name) {
      window.alert("请输入名称。");
      return;
    }
    if (!/^\d{4}-\d{2}-\d{2}$/.test(newDate)) {
      window.alert("请选择日期。");
      return;
    }
    setBusy(true);
    try {
      await onAdd(name, newDate);
      setNewName("");
      setNewDate("");
    } catch (e) {
      window.alert(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <Fragment>
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="shrink-0 px-3 py-2 text-[11px] font-bold uppercase tracking-wide text-[#bbbbbb]">
        我的提醒
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto px-1 pb-2">
        <ul className="flex flex-col gap-px">
          {sorted.length === 0 ? (
            <li className="px-2 py-2 text-[12px] text-[#858585]">
              暂无条目。在下方填写名称与日期后添加（数据保存在 reminder.db）。
            </li>
          ) : (
            sorted.map((r) => {
              const sel = r.id === selectedId;
              const editing = r.id === editingId;
              return (
                <li key={r.id}>
                  <div
                    className={`${rowClass} ${sel ? selectedRow : ""} flex-col items-stretch gap-1`}
                    role="button"
                    tabIndex={0}
                    onClick={() => !editing && !busy && onSelect(r.id)}
                    onKeyDown={(e) => {
                      if (editing || busy) return;
                      if (e.key === "Enter" || e.key === " ") {
                        e.preventDefault();
                        onSelect(r.id);
                      }
                    }}
                  >
                    {editing ? (
                      <div
                        className="flex flex-col gap-1.5"
                        onClick={(e) => e.stopPropagation()}
                      >
                        <input
                          type="text"
                          className="allow-select rounded border border-[#007fd4] bg-[#3c3c3c] px-1.5 py-1 text-[12px] text-[#cccccc] outline-none"
                          value={draftName}
                          onChange={(e) => setDraftName(e.target.value)}
                          placeholder="名称"
                        />
                        <input
                          type="date"
                          className="allow-select rounded border border-[#007fd4] bg-[#3c3c3c] px-1.5 py-1 font-mono text-[11px] text-[#cccccc] outline-none"
                          value={draftDate}
                          onChange={(e) => setDraftDate(e.target.value)}
                        />
                        <div className="flex gap-1">
                          <button
                            type="button"
                            disabled={busy}
                            className="rounded bg-[#0e639c] px-2 py-0.5 text-[11px] text-white hover:bg-[#1177bb] disabled:opacity-50"
                            onClick={() => void commitEdit()}
                          >
                            保存
                          </button>
                          <button
                            type="button"
                            disabled={busy}
                            className="rounded px-2 py-0.5 text-[11px] text-[#858585] hover:bg-white/10 disabled:opacity-50"
                            onClick={cancelEdit}
                          >
                            取消
                          </button>
                        </div>
                      </div>
                    ) : (
                      <>
                        <div className="flex min-w-0 items-center justify-between gap-1">
                          <span className="min-w-0 truncate font-medium">
                            {r.name}
                          </span>
                          <div className="flex shrink-0 gap-0.5 opacity-0 group-hover:opacity-100">
                            <button
                              type="button"
                              title="编辑"
                              disabled={busy}
                              className="rounded p-0.5 hover:bg-white/15 disabled:opacity-40"
                              aria-label={`编辑 ${r.name}`}
                              onClick={(e) => {
                                e.stopPropagation();
                                startEdit(r);
                              }}
                            >
                              <svg
                                width="14"
                                height="14"
                                viewBox="0 0 24 24"
                                fill="currentColor"
                                className="text-[#858585]"
                              >
                                <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z" />
                              </svg>
                            </button>
                            <button
                              type="button"
                              title="删除"
                              disabled={busy}
                              className="rounded p-0.5 hover:bg-white/15 disabled:opacity-40"
                              aria-label={`删除 ${r.name}`}
                              onClick={(e) => {
                                e.stopPropagation();
                                requestDelete(r);
                              }}
                            >
                              <svg
                                width="14"
                                height="14"
                                viewBox="0 0 24 24"
                                fill="currentColor"
                                className="text-[#858585]"
                              >
                                <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z" />
                              </svg>
                            </button>
                          </div>
                        </div>
                        <div className="font-mono text-[11px] text-[#858585]">
                          {r.date}
                        </div>
                      </>
                    )}
                  </div>
                </li>
              );
            })
          )}
        </ul>
      </div>
      <div className="shrink-0 border-t border-[var(--vscode-border)] px-2 py-2">
        <div className="mb-1.5 text-[10px] font-bold uppercase tracking-wide text-[#858585]">
          添加
        </div>
        <div className="flex flex-col gap-1.5">
          <input
            type="text"
            className="allow-select w-full rounded border border-[var(--vscode-border)] bg-[#3c3c3c] px-2 py-1 text-[12px] text-[#cccccc] outline-none focus:border-[#007fd4]"
            placeholder="名称"
            value={newName}
            disabled={busy}
            onChange={(e) => setNewName(e.target.value)}
          />
          <input
            type="date"
            className="allow-select w-full rounded border border-[var(--vscode-border)] bg-[#3c3c3c] px-2 py-1 font-mono text-[11px] text-[#cccccc] outline-none focus:border-[#007fd4]"
            value={newDate}
            disabled={busy}
            onChange={(e) => setNewDate(e.target.value)}
          />
          <button
            type="button"
            disabled={busy}
            className="w-full rounded border border-[var(--vscode-border)] bg-[#0e639c] py-1.5 text-[12px] text-white hover:bg-[#1177bb] disabled:opacity-50"
            onClick={() => void add()}
          >
            {busy ? "处理中…" : "添加提醒"}
          </button>
        </div>
      </div>
    </div>
    {pendingDelete ? (
      <div
        className="fixed inset-0 z-[200] flex items-center justify-center bg-black/50 p-4"
        role="presentation"
        onMouseDown={(e) => {
          if (e.target === e.currentTarget) cancelDelete();
        }}
      >
        <div
          className="w-full max-w-[320px] rounded border border-[var(--vscode-border)] bg-[#252526] p-4 shadow-xl"
          role="alertdialog"
          aria-modal="true"
          aria-labelledby="del-title"
          onMouseDown={(e) => e.stopPropagation()}
        >
          <p id="del-title" className="text-[13px] leading-relaxed text-[#cccccc]">
            确定删除「{pendingDelete.name}」（{pendingDelete.date}）？
          </p>
          <div className="mt-4 flex justify-end gap-2">
            <button
              type="button"
              className="rounded px-3 py-1.5 text-[12px] text-[#cccccc] hover:bg-white/10"
              onClick={cancelDelete}
            >
              取消
            </button>
            <button
              type="button"
              disabled={busy}
              className="rounded bg-[#a1260d] px-3 py-1.5 text-[12px] text-white hover:bg-[#c72e1a] disabled:opacity-50"
              onClick={() => void executeConfirmedDelete()}
            >
              {busy ? "删除中…" : "删除"}
            </button>
          </div>
        </div>
      </div>
    ) : null}
    </Fragment>
  );
}
