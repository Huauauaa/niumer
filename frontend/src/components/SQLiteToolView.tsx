import { useCallback, useEffect, useMemo, useState } from "react";
import {
  ChooseSQLiteDBPath,
  SQLiteToolDeleteRow,
  SQLiteToolGetDBPath,
  SQLiteToolInsertRow,
  SQLiteToolListTables,
  SQLiteToolOpenDB,
  SQLiteToolQueryTable,
  SQLiteToolUpdateRow,
} from "../../wailsjs/go/main/App";

type Column = {
  name: string;
  type: string;
  notNull: boolean;
  pk: boolean;
  default: unknown;
};

type QueryResult = {
  table: string;
  columns: Column[];
  pkColumns: string[];
  rows: Record<string, unknown>[];
  total: number;
  limit: number;
  offset: number;
};

function ModalShell({
  title,
  children,
  onClose,
}: {
  title: string;
  children: React.ReactNode;
  onClose: () => void;
}) {
  return (
    <div
      className="fixed inset-0 z-30 flex items-center justify-center bg-black/45 px-4 py-8"
      role="presentation"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        role="dialog"
        aria-label={title}
        className="w-full max-w-lg rounded-2xl border border-[var(--vscode-border)] bg-[var(--vscode-sideBar-bg)] px-5 py-4 shadow-xl"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className="mb-3 flex items-center justify-between gap-2">
          <div className="text-[14px] font-semibold text-[var(--vscode-fg)]">
            {title}
          </div>
          <button
            type="button"
            className="rounded px-2 py-1 text-[12px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-list-hover)]"
            onClick={onClose}
          >
            Esc
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

function ConfirmModal({
  title,
  message,
  confirmText = "Delete",
  cancelText = "Cancel",
  confirmTone = "danger",
  loading = false,
  onConfirm,
  onClose,
}: {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  confirmTone?: "danger" | "primary";
  loading?: boolean;
  onConfirm: () => void;
  onClose: () => void;
}) {
  return (
    <ModalShell title={title} onClose={onClose}>
      <div className="text-[13px] leading-relaxed text-[var(--vscode-fg)]">
        {message}
      </div>
      <div className="mt-4 flex justify-end gap-2">
        <button
          type="button"
          className="rounded-lg border border-[var(--vscode-border)] px-4 py-2 text-[13px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-list-hover)] disabled:opacity-50"
          onClick={onClose}
          disabled={loading}
        >
          {cancelText}
        </button>
        <button
          type="button"
          className={[
            "rounded-lg px-4 py-2 text-[13px] font-medium text-white disabled:opacity-50",
            confirmTone === "danger"
              ? "bg-[#f48771] hover:bg-[#e06a59]"
              : "bg-[#4a9eff] hover:bg-[#3d8eef]",
          ].join(" ")}
          onClick={onConfirm}
          disabled={loading}
        >
          {confirmText}
        </button>
      </div>
    </ModalShell>
  );
}

function coerceInput(v: string): unknown {
  const s = v.trim();
  if (s === "") return "";
  if (s === "null" || s === "NULL") return null;
  if (/^-?\d+$/.test(s)) return Number(s);
  if (/^-?\d+\.\d+$/.test(s)) return Number(s);
  return v;
}

const RECENT_KEY = "niumer.sqliteTool.recentDBs";
const RECENT_LIMIT = 8;

function loadRecentDBs(): string[] {
  try {
    const raw = localStorage.getItem(RECENT_KEY);
    if (!raw) return [];
    const arr = JSON.parse(raw);
    if (!Array.isArray(arr)) return [];
    return arr.filter((x) => typeof x === "string" && x.trim() !== "").slice(0, RECENT_LIMIT);
  } catch {
    return [];
  }
}

function saveRecentDBs(list: string[]) {
  try {
    localStorage.setItem(RECENT_KEY, JSON.stringify(list.slice(0, RECENT_LIMIT)));
  } catch {
    // ignore
  }
}

export function SQLiteToolView() {
  const [dbPath, setDbPath] = useState("");
  const [recentDBs, setRecentDBs] = useState<string[]>(() => loadRecentDBs());
  const [tables, setTables] = useState<string[]>([]);
  const [table, setTable] = useState<string>("");
  const [res, setRes] = useState<QueryResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [limit, setLimit] = useState(50);
  const [offset, setOffset] = useState(0);

  const [editOpen, setEditOpen] = useState(false);
  const [editRow, setEditRow] = useState<Record<string, unknown> | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [pendingDeleteRow, setPendingDeleteRow] = useState<Record<string, unknown> | null>(
    null,
  );

  const pkCols = res?.pkColumns ?? [];

  const refreshTables = useCallback(async () => {
    const list = await SQLiteToolListTables();
    setTables(list);
    setTable((cur) => (cur && list.includes(cur) ? cur : list[0] ?? ""));
  }, []);

  const refreshQuery = useCallback(
    async (t: string, nextOffset: number) => {
      if (!t) return;
      setLoading(true);
      setError(null);
      try {
        const r = (await SQLiteToolQueryTable(t, limit, nextOffset)) as QueryResult;
        setRes(r);
      } catch (e) {
        setRes(null);
        setError(e instanceof Error ? e.message : String(e));
      } finally {
        setLoading(false);
      }
    },
    [limit],
  );

  useEffect(() => {
    void SQLiteToolGetDBPath()
      .then((p) => {
        const next = p || "";
        setDbPath(next);
        if (next) {
          setRecentDBs((prev) => {
            const merged = [next, ...prev.filter((x) => x !== next)].slice(0, RECENT_LIMIT);
            saveRecentDBs(merged);
            return merged;
          });
        }
      })
      .catch(() => setDbPath(""));
  }, []);

  useEffect(() => {
    if (!dbPath) return;
    void refreshTables().catch((e) => setError(e instanceof Error ? e.message : String(e)));
  }, [dbPath, refreshTables]);

  useEffect(() => {
    if (!table) return;
    setOffset(0);
    void refreshQuery(table, 0);
  }, [table, refreshQuery]);

  const openDB = useCallback(
    async (picked?: string) => {
      const path = picked ?? (await ChooseSQLiteDBPath());
      if (!path) return;
      setLoading(true);
      setError(null);
      try {
        await SQLiteToolOpenDB(path);
        setDbPath(path);
        setRecentDBs((prev) => {
          const merged = [path, ...prev.filter((x) => x !== path)].slice(0, RECENT_LIMIT);
          saveRecentDBs(merged);
          return merged;
        });
        await refreshTables();
      } catch (e) {
        setError(e instanceof Error ? e.message : String(e));
      } finally {
        setLoading(false);
      }
    },
    [refreshTables],
  );

  const cols = res?.columns ?? [];
  const rows = res?.rows ?? [];

  const canMutate = useMemo(() => pkCols.length > 0, [pkCols.length]);

  const keyOf = useCallback(
    (row: Record<string, unknown>) => {
      const key: Record<string, unknown> = {};
      for (const k of pkCols) key[k] = row[k];
      return key;
    },
    [pkCols],
  );

  const saveEdit = useCallback(async () => {
    if (!res || !editRow) return;
    const key = keyOf(editRow);
    const values: Record<string, unknown> = {};
    for (const c of cols) {
      if (pkCols.includes(c.name)) continue;
      values[c.name] = editRow[c.name];
    }
    setLoading(true);
    setError(null);
    try {
      await SQLiteToolUpdateRow(res.table, key, values);
      setEditOpen(false);
      setEditRow(null);
      await refreshQuery(res.table, offset);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  }, [cols, editRow, keyOf, offset, pkCols, refreshQuery, res]);

  const deleteRow = useCallback(
    async (row: Record<string, unknown>) => {
      if (!res) return;
      setPendingDeleteRow(row);
      setConfirmOpen(true);
      return;
    },
    [res],
  );

  const confirmDelete = useCallback(async () => {
    if (!res || !pendingDeleteRow) return;
    setLoading(true);
    setError(null);
    try {
      await SQLiteToolDeleteRow(res.table, keyOf(pendingDeleteRow));
      await refreshQuery(res.table, offset);
      setConfirmOpen(false);
      setPendingDeleteRow(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  }, [keyOf, offset, pendingDeleteRow, refreshQuery, res]);

  const createRow = useCallback(
    async (draft: Record<string, unknown>) => {
      if (!res) return;
      setLoading(true);
      setError(null);
      try {
        await SQLiteToolInsertRow(res.table, draft);
        setCreateOpen(false);
        await refreshQuery(res.table, 0);
        setOffset(0);
      } catch (e) {
        setError(e instanceof Error ? e.message : String(e));
      } finally {
        setLoading(false);
      }
    },
    [refreshQuery, res],
  );

  return (
    <div className="flex min-h-0 min-w-0 flex-1">
      <aside className="flex w-64 shrink-0 flex-col border-r border-[var(--vscode-border)] bg-[var(--vscode-sideBar-bg)]">
        <div className="px-3 py-2">
          <div className="mb-2 flex items-center justify-between gap-2">
            <div className="text-[11px] font-bold uppercase tracking-wide text-[var(--vscode-fg-muted)]">
              Database
            </div>
            <button
              type="button"
              className="rounded-lg border border-[var(--vscode-border)] px-2 py-1 text-[12px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-list-hover)] disabled:opacity-50"
              onClick={() => void openDB()}
              disabled={loading}
            >
              打开
            </button>
          </div>
          <select
            className="w-full rounded-lg border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-2 py-1.5 text-[12px] text-[var(--vscode-fg)]"
            value={dbPath || ""}
            onChange={(e) => void openDB(e.target.value)}
            disabled={loading || recentDBs.length === 0}
            title={dbPath || ""}
          >
            {recentDBs.length === 0 ? (
              <option value="">No recent DBs</option>
            ) : null}
            {recentDBs.map((p) => (
              <option key={p} value={p}>
                {p.split("/").slice(-2).join("/")}
              </option>
            ))}
          </select>
          <div className="mt-1 truncate text-[10px] text-[var(--vscode-fg-muted)]">
            {dbPath ? dbPath : "No database selected"}
          </div>
        </div>
        <div className="px-3 pb-1 text-[11px] font-bold uppercase tracking-wide text-[var(--vscode-fg-muted)]">
          Tables
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto px-1 pb-2">
          {tables.length === 0 ? (
            <div className="px-2 py-2 text-[12px] text-[var(--vscode-fg-muted)]">
              {dbPath ? "No tables." : "Open a SQLite database first."}
            </div>
          ) : (
            <ul className="flex flex-col gap-px">
              {tables.map((t) => (
                <li key={t}>
                  <button
                    type="button"
                    className={`group flex w-full cursor-pointer items-center gap-1 rounded px-2 py-1 text-left text-[13px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-list-hover)] ${
                      t === table
                        ? "bg-[var(--vscode-sidebar-item-selected)] hover:bg-[var(--vscode-sidebar-item-selected)]"
                        : ""
                    }`}
                    onClick={() => setTable(t)}
                  >
                    <span className="min-w-0 flex-1 truncate">{t}</span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </aside>

      <main className="flex min-h-0 min-w-0 flex-1 flex-col bg-[var(--vscode-editor-bg)]">
        <div className="flex shrink-0 items-center justify-between gap-2 border-b border-[var(--vscode-border)] px-4 py-2">
          <div className="min-w-0 truncate text-[13px] text-[var(--vscode-fg)]">
            {table ? (
              <>
                <span className="text-[var(--vscode-fg-muted)]">Table </span>
                <span className="font-mono text-[#b5cea8]">{table}</span>
                {res ? (
                  <span className="ml-2 text-[12px] text-[var(--vscode-fg-muted)]">
                    {res.total} rows
                  </span>
                ) : null}
              </>
            ) : (
              <span className="text-[var(--vscode-fg-muted)]">Select a table</span>
            )}
          </div>
          <div className="flex items-center gap-2">
            <select
              className="rounded-lg border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-2 py-1 text-[12px] text-[var(--vscode-fg)]"
              value={String(limit)}
              onChange={(e) => setLimit(Number(e.target.value) || 50)}
              disabled={!table || loading}
            >
              {[25, 50, 100, 200].map((n) => (
                <option key={n} value={n}>
                  {n} / page
                </option>
              ))}
            </select>
            <button
              type="button"
              className="rounded-lg border border-[var(--vscode-border)] px-3 py-1 text-[12px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-list-hover)] disabled:opacity-50"
              disabled={!table || loading || !res || res.offset <= 0}
              onClick={() => {
                const next = Math.max(0, offset - limit);
                setOffset(next);
                void refreshQuery(table, next);
              }}
            >
              Prev
            </button>
            <button
              type="button"
              className="rounded-lg border border-[var(--vscode-border)] px-3 py-1 text-[12px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-list-hover)] disabled:opacity-50"
              disabled={!table || loading || !res || res.offset + res.limit >= res.total}
              onClick={() => {
                const next = offset + limit;
                setOffset(next);
                void refreshQuery(table, next);
              }}
            >
              Next
            </button>
            <button
              type="button"
              className="rounded-lg bg-[#4a9eff] px-3 py-1 text-[12px] font-medium text-white hover:bg-[#3d8eef] disabled:opacity-50"
              disabled={!table || loading}
              onClick={() => setCreateOpen(true)}
            >
              新增
            </button>
          </div>
        </div>

        {error ? (
          <div className="shrink-0 border-b border-[#5a1d1d] bg-[#3c1c1c] px-4 py-2 text-[12px] text-[#f48771]">
            {error}
          </div>
        ) : null}

        <div className="allow-select min-h-0 flex-1 overflow-auto">
          {!table ? (
            <div className="flex h-full items-center justify-center text-[13px] text-[var(--vscode-fg-muted)]">
              Open a database and pick a table.
            </div>
          ) : loading && !res ? (
            <div className="flex h-full items-center justify-center text-[13px] text-[var(--vscode-fg-muted)]">
              Loading…
            </div>
          ) : res ? (
            <table className="w-full border-collapse text-[12px]">
              <thead className="sticky top-0 bg-[var(--vscode-sideBar-bg)]">
                <tr>
                  <th className="border-b border-[var(--vscode-border)] px-3 py-2 text-left text-[11px] uppercase tracking-wide text-[var(--vscode-fg-muted)]">
                    Actions
                  </th>
                  {cols.map((c) => (
                    <th
                      key={c.name}
                      className="border-b border-[var(--vscode-border)] px-3 py-2 text-left text-[11px] uppercase tracking-wide text-[var(--vscode-fg-muted)]"
                    >
                      {c.name}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rows.map((row, idx) => (
                  <tr
                    key={idx}
                    className="hover:bg-[var(--vscode-list-hover)]"
                  >
                    <td className="border-b border-[var(--vscode-border)] px-3 py-2 align-top">
                      <div className="flex gap-2">
                        <button
                          type="button"
                          className="text-[#3794ff] hover:underline disabled:opacity-50"
                          disabled={!canMutate || loading}
                          onClick={() => {
                            setEditRow({ ...row });
                            setEditOpen(true);
                          }}
                        >
                          Edit
                        </button>
                        <button
                          type="button"
                          className="text-[#f48771] hover:underline disabled:opacity-50"
                          disabled={!canMutate || loading}
                          onClick={() => void deleteRow(row)}
                        >
                          Delete
                        </button>
                      </div>
                    </td>
                    {cols.map((c) => (
                      <td
                        key={c.name}
                        className="border-b border-[var(--vscode-border)] px-3 py-2 align-top"
                      >
                        {row[c.name] == null ? (
                          <span className="text-[var(--vscode-fg-muted)]">
                            null
                          </span>
                        ) : (
                          String(row[c.name])
                        )}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <div className="flex h-full items-center justify-center text-[13px] text-[var(--vscode-fg-muted)]">
              No data.
            </div>
          )}
        </div>
      </main>

      {editOpen && editRow && res ? (
        <ModalShell
          title={`Edit ${res.table}`}
          onClose={() => {
            setEditOpen(false);
            setEditRow(null);
          }}
        >
          {!canMutate ? (
            <div className="text-[12px] text-[var(--vscode-fg-muted)]">
              This table has no primary key/rowid. Edit/Delete is disabled.
            </div>
          ) : (
            <>
              <div className="max-h-[55vh] overflow-auto pr-1">
                <div className="grid grid-cols-1 gap-2">
                  {cols.map((c) => {
                    const disabled = pkCols.includes(c.name);
                    return (
                      <label key={c.name} className="block">
                        <div className="mb-1 text-[11px] font-medium uppercase tracking-wide text-[var(--vscode-fg-muted)]">
                          {c.name}{" "}
                          <span className="font-mono normal-case text-[10px]">
                            {c.type || ""}
                            {c.pk ? " pk" : ""}
                          </span>
                        </div>
                        <input
                          className="w-full rounded-lg border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-3 py-2 text-[13px] text-[var(--vscode-fg)] outline-none disabled:opacity-60"
                          value={
                            editRow[c.name] == null ? "" : String(editRow[c.name])
                          }
                          disabled={disabled}
                          placeholder={editRow[c.name] == null ? "null" : ""}
                          onChange={(e) =>
                            setEditRow((prev) => ({
                              ...(prev ?? {}),
                              [c.name]: coerceInput(e.target.value),
                            }))
                          }
                        />
                      </label>
                    );
                  })}
                </div>
              </div>
              <div className="mt-4 flex justify-end gap-2">
                <button
                  type="button"
                  className="rounded-lg border border-[var(--vscode-border)] px-4 py-2 text-[13px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-list-hover)]"
                  onClick={() => {
                    setEditOpen(false);
                    setEditRow(null);
                  }}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="rounded-lg bg-[#4a9eff] px-4 py-2 text-[13px] font-medium text-white hover:bg-[#3d8eef] disabled:opacity-50"
                  disabled={loading}
                  onClick={() => void saveEdit()}
                >
                  Save
                </button>
              </div>
            </>
          )}
        </ModalShell>
      ) : null}

      {createOpen && res ? (
        <CreateRowModal
          columns={cols}
          onClose={() => setCreateOpen(false)}
          onCreate={(draft) => void createRow(draft)}
          loading={loading}
        />
      ) : null}

      {confirmOpen && pendingDeleteRow && res ? (
        <ConfirmModal
          title="Delete row"
          message={`Delete 1 row from "${res.table}"? This cannot be undone.`}
          confirmText="Delete"
          cancelText="Cancel"
          confirmTone="danger"
          loading={loading}
          onClose={() => {
            if (loading) return;
            setConfirmOpen(false);
            setPendingDeleteRow(null);
          }}
          onConfirm={() => void confirmDelete()}
        />
      ) : null}
    </div>
  );
}

function CreateRowModal({
  columns,
  onClose,
  onCreate,
  loading,
}: {
  columns: Column[];
  onClose: () => void;
  onCreate: (values: Record<string, unknown>) => void;
  loading: boolean;
}) {
  const [draft, setDraft] = useState<Record<string, unknown>>({});

  return (
    <ModalShell title="Insert row" onClose={onClose}>
      <div className="max-h-[55vh] overflow-auto pr-1">
        <div className="grid grid-cols-1 gap-2">
          {columns.map((c) => (
            <label key={c.name} className="block">
              <div className="mb-1 text-[11px] font-medium uppercase tracking-wide text-[var(--vscode-fg-muted)]">
                {c.name}{" "}
                <span className="font-mono normal-case text-[10px]">
                  {c.type || ""}
                  {c.pk ? " pk" : ""}
                </span>
              </div>
              <input
                className="w-full rounded-lg border border-[var(--vscode-border)] bg-[var(--vscode-input-bg)] px-3 py-2 text-[13px] text-[var(--vscode-fg)] outline-none"
                value={draft[c.name] == null ? "" : String(draft[c.name])}
                placeholder="(empty)"
                onChange={(e) =>
                  setDraft((prev) => ({
                    ...prev,
                    [c.name]: coerceInput(e.target.value),
                  }))
                }
              />
            </label>
          ))}
        </div>
      </div>
      <div className="mt-4 flex justify-end gap-2">
        <button
          type="button"
          className="rounded-lg border border-[var(--vscode-border)] px-4 py-2 text-[13px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-list-hover)]"
          onClick={onClose}
        >
          Cancel
        </button>
        <button
          type="button"
          className="rounded-lg bg-[#4a9eff] px-4 py-2 text-[13px] font-medium text-white hover:bg-[#3d8eef] disabled:opacity-50"
          disabled={loading}
          onClick={() => onCreate(draft)}
        >
          Create
        </button>
      </div>
    </ModalShell>
  );
}

