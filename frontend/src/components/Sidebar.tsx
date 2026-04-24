import { useEffect, useRef, useState } from "react";
import { normalizeMarkdownFileName, type BlogDocument } from "../types/blog";
import type { ActivityId } from "./ActivityBar";
import { ReminderCustomSidebar } from "./ReminderCustomSidebar";
import type { CustomReminder } from "../types/reminder";
import type {
  PullRequestListItem,
  PullRequestState,
} from "../types/pullRequest";

type Props = {
  activity: ActivityId;
  width: number;
  onResizeStart: () => void;
  /** Blog explorer (only when activity === blog) */
  blogDocuments: BlogDocument[];
  blogSelectedId: string;
  onBlogOpen: (id: string) => void;
  onBlogNew: () => void;
  onBlogDelete: (id: string) => void;
  onBlogRename: (oldName: string, newName: string) => void | Promise<void>;
  /** 各条记录 effectiveWorkHours 之和（小时） */
  workHourTotalEffectiveHours?: number;
  /** 加班（小时）：按日历日汇总 max(0, 当日工时合计 − 8) 后求和 */
  workHourOvertimeHours?: number;
  /** 来自 /user-info 的 data.shiftInformationDTO.shiftNameZh（刷新考勤后更新） */
  workHourShiftNameZh?: string;
  /** Pull Request list (activity === pullRequest); Go RefreshPullRequestList (configs pull_request_list_url) */
  pullRequestItems?: PullRequestListItem[];
  pullRequestPage?: number;
  pullRequestTotalPages?: number;
  pullRequestLoading?: boolean;
  pullRequestError?: string | null;
  pullRequestSelectedNumber?: number | null;
  onPullRequestSelect?: (item: PullRequestListItem) => void;
  onPullRequestPageChange?: (page: number) => void;
  /** Reminder activity: custom dated items */
  customReminders?: CustomReminder[];
  selectedReminderId?: string | null;
  onReminderSelect?: (id: string | null) => void;
  onReminderAdd?: (name: string, date: string) => Promise<void>;
  onReminderUpdate?: (id: string, name: string, date: string) => Promise<void>;
  onReminderDelete?: (row: CustomReminder) => Promise<void>;
};

const rowClass =
  "group flex w-full cursor-pointer items-center gap-1 rounded px-2 py-1 text-left text-[13px] text-[#cccccc] hover:bg-[var(--vscode-list-hover)]";
const selectedRow = "bg-[#37373d] hover:bg-[#37373d]";

function prStateDotClass(state: PullRequestState): string {
  if (state === "merged") return "text-[#6a9955]";
  if (state === "open") return "text-[#4fc1ff]";
  return "text-[#858585]";
}

function BlogExplorer({
  documents,
  selectedId,
  onOpen,
  onNew,
  onDelete,
  onRename,
}: {
  documents: BlogDocument[];
  selectedId: string;
  onOpen: (id: string) => void;
  onNew: () => void;
  onDelete: (id: string) => void;
  onRename: (oldName: string, newName: string) => void | Promise<void>;
}) {
  const [renaming, setRenaming] = useState<string | null>(null);
  const [draft, setDraft] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const draftRef = useRef("");
  const inflightRef = useRef(false);
  draftRef.current = draft;

  useEffect(() => {
    if (renaming && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [renaming]);

  const commitRename = async (oldName: string) => {
    if (inflightRef.current) return;
    const next = normalizeMarkdownFileName(draftRef.current);
    if (!next) {
      window.alert("File name cannot be empty.");
      return;
    }
    if (next === oldName) {
      setRenaming(null);
      return;
    }
    inflightRef.current = true;
    try {
      await onRename(oldName, next);
      setRenaming(null);
    } finally {
      inflightRef.current = false;
    }
  };

  const cancelRename = () => {
    setRenaming(null);
  };

  return (
    <div
      className="flex min-h-0 flex-1 flex-col outline-none"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === "F2" && selectedId && !renaming) {
          e.preventDefault();
          setRenaming(selectedId);
          setDraft(selectedId);
        }
      }}
    >
      <div className="flex shrink-0 items-center justify-between px-2 py-2">
        <span className="pl-1 text-[11px] font-bold uppercase tracking-wide text-[#bbbbbb]">
          Explorer
        </span>
        <div className="flex items-center gap-0.5">
          <button
            type="button"
            title="New File"
            aria-label="New File"
            className="rounded p-1 text-[#cccccc] hover:bg-white/10"
            onClick={onNew}
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="currentColor"
              aria-hidden
            >
              <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 14h-3v3h-2v-3H8v-2h3v-3h2v3h3v2zm-3-7V3.5L18.5 9H13z" />
            </svg>
          </button>
        </div>
      </div>
      <div className="px-3 pb-1 text-[11px] font-bold uppercase tracking-wide text-[#858585]">
        Blog
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto px-1 pb-2">
        <ul className="flex flex-col gap-px">
          {documents.length === 0 ? (
            <li className="px-2 py-2 text-[12px] text-[#858585]">
              No files yet. Click New File.
            </li>
          ) : (
            documents.map((doc) => {
              const isSel = doc.fileName === selectedId;
              const isEditing = renaming === doc.fileName;
              return (
                <li key={doc.fileName}>
                  <div
                    className={`${rowClass} ${isSel ? selectedRow : ""}`}
                    onClick={() => !isEditing && onOpen(doc.fileName)}
                    onKeyDown={(e) => {
                      if (isEditing) return;
                      if (e.key === "Enter" || e.key === " ") {
                        e.preventDefault();
                        onOpen(doc.fileName);
                      }
                    }}
                    role="button"
                    tabIndex={isEditing ? -1 : 0}
                  >
                    <span className="shrink-0 text-[#519aba]" aria-hidden>
                      M
                    </span>
                    {isEditing ? (
                      <div
                        className="flex min-w-0 flex-1 items-center gap-0.5"
                        onClick={(e) => e.stopPropagation()}
                      >
                        <input
                          ref={inputRef}
                          type="text"
                          className="allow-select min-w-0 flex-1 rounded border border-[#007fd4] bg-[#3c3c3c] px-1 py-0.5 font-mono text-[12px] text-[#cccccc] outline-none"
                          value={draft}
                          onChange={(e) => setDraft(e.target.value)}
                          onKeyDown={(e) => {
                            if (e.key === "Enter") {
                              e.preventDefault();
                              void commitRename(doc.fileName);
                            }
                            if (e.key === "Escape") {
                              e.preventDefault();
                              cancelRename();
                            }
                          }}
                        />
                        <button
                          type="button"
                          title="Apply rename"
                          className="shrink-0 rounded px-1 py-0.5 text-[11px] text-[#cccccc] hover:bg-white/15"
                          onClick={() => void commitRename(doc.fileName)}
                        >
                          ✓
                        </button>
                      </div>
                    ) : (
                      <span
                        className="min-w-0 flex-1 truncate"
                        onDoubleClick={(e) => {
                          e.stopPropagation();
                          setRenaming(doc.fileName);
                          setDraft(doc.fileName);
                        }}
                      >
                        {doc.title}
                      </span>
                    )}
                    {!isEditing && doc.dirty && (
                      <span className="shrink-0 text-[#e37933]">●</span>
                    )}
                    {!isEditing && (
                      <>
                        <button
                          type="button"
                          title="Rename (or F2 when Explorer is focused)"
                          className="shrink-0 rounded p-0.5 opacity-0 hover:bg-white/15 group-hover:opacity-100"
                          aria-label={`Rename ${doc.title}`}
                          onClick={(e) => {
                            e.stopPropagation();
                            setRenaming(doc.fileName);
                            setDraft(doc.fileName);
                          }}
                        >
                          <svg
                            width="14"
                            height="14"
                            viewBox="0 0 24 24"
                            fill="currentColor"
                            className="text-[#858585] hover:text-[#cccccc]"
                            aria-hidden
                          >
                            <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z" />
                          </svg>
                        </button>
                        <button
                          type="button"
                          title={`Delete ${doc.title}`}
                          className="shrink-0 rounded p-0.5 opacity-0 hover:bg-white/15 group-hover:opacity-100"
                          aria-label={`Delete ${doc.title}`}
                          onClick={(e) => {
                            e.stopPropagation();
                            onDelete(doc.fileName);
                          }}
                        >
                          <svg
                            width="14"
                            height="14"
                            viewBox="0 0 24 24"
                            fill="currentColor"
                            className="text-[#858585] hover:text-[#cccccc]"
                          >
                            <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z" />
                          </svg>
                        </button>
                      </>
                    )}
                  </div>
                </li>
              );
            })
          )}
        </ul>
      </div>
    </div>
  );
}

function SidebarBody({
  activity,
  blogDocuments,
  blogSelectedId,
  onBlogOpen,
  onBlogNew,
  onBlogDelete,
  onBlogRename,
  workHourTotalEffectiveHours,
  workHourOvertimeHours,
  workHourShiftNameZh = "",
  pullRequestItems = [],
  pullRequestPage = 1,
  pullRequestTotalPages = 1,
  pullRequestLoading = false,
  pullRequestError = null,
  pullRequestSelectedNumber = null,
  onPullRequestSelect,
  onPullRequestPageChange,
  customReminders = [],
  selectedReminderId = null,
  onReminderSelect,
  onReminderAdd,
  onReminderUpdate,
  onReminderDelete,
}: Omit<Props, "width" | "onResizeStart">) {
  if (activity === "blog") {
    return (
      <BlogExplorer
        documents={blogDocuments}
        selectedId={blogSelectedId}
        onOpen={onBlogOpen}
        onNew={onBlogNew}
        onDelete={onBlogDelete}
        onRename={onBlogRename}
      />
    );
  }

  if (activity === "tool") {
    return (
      <div className="flex flex-col">
        <div className="px-3 py-2 text-[11px] font-bold uppercase tracking-wide text-[#bbbbbb]">
          Tool
        </div>
        <div className="flex flex-col gap-0.5 px-1 pb-2">
          <button type="button" className={rowClass}>
            <span className="text-[#569cd6]">{}</span>
            <span>JSON formatter</span>
          </button>
        </div>
      </div>
    );
  }

  if (activity === "pullRequest") {
    return (
      <div className="flex min-h-0 flex-1 flex-col">
        <div className="shrink-0 px-3 py-2 text-[11px] font-bold uppercase tracking-wide text-[#bbbbbb]">
          Pull Request
        </div>
        <div className="allow-select min-h-0 flex-1 overflow-y-auto px-1 pb-1">
          {pullRequestLoading ? (
            <div className="px-2 py-2 text-[12px] text-[#858585]">Loading…</div>
          ) : pullRequestError ? (
            <div className="px-2 py-2 text-[12px] leading-snug text-[#f48771]">
              {pullRequestError}
              <div className="mt-1 text-[11px] text-[#858585]">
                Start mockserver:{" "}
                <code className="font-mono">go run ./cmd/mockserver</code>
              </div>
            </div>
          ) : pullRequestItems.length === 0 ? (
            <div className="px-2 py-2 text-[12px] text-[#858585]">
              No items.
            </div>
          ) : (
            <ul className="flex flex-col gap-0.5">
              {pullRequestItems.map((pr) => {
                const sel = pr.number === pullRequestSelectedNumber;
                return (
                  <li key={pr.id}>
                    <button
                      type="button"
                      className={`${rowClass} flex-col items-stretch gap-0.5 py-1.5 ${sel ? selectedRow : ""}`}
                      onClick={() => onPullRequestSelect?.(pr)}
                    >
                      <div className="flex min-w-0 items-center gap-1">
                        <span
                          className={`shrink-0 text-[10px] leading-none ${prStateDotClass(pr.state)}`}
                          aria-hidden
                        >
                          ●
                        </span>
                        <span className="min-w-0 truncate font-medium">
                          #{pr.number} {pr.title}
                        </span>
                      </div>
                      <div className="truncate pl-3 text-[11px] text-[#858585]">
                        {pr.author} ·{" "}
                        <span className="font-mono text-[#b5cea8]">
                          {pr.sourceBranch}
                        </span>{" "}
                        →{" "}
                        <span className="font-mono text-[#b5cea8]">
                          {pr.targetBranch}
                        </span>
                      </div>
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
        <div className="shrink-0 border-t border-[var(--vscode-border)] px-2 py-2">
          <div className="flex items-center justify-between gap-1 text-[11px] text-[#858585]">
            <button
              type="button"
              className="rounded px-1.5 py-0.5 hover:bg-white/10 disabled:opacity-40"
              disabled={pullRequestLoading || pullRequestPage <= 1}
              onClick={() => onPullRequestPageChange?.(pullRequestPage - 1)}
            >
              Prev
            </button>
            <span className="tabular-nums">
              {pullRequestPage} / {pullRequestTotalPages}
            </span>
            <button
              type="button"
              className="rounded px-1.5 py-0.5 hover:bg-white/10 disabled:opacity-40"
              disabled={
                pullRequestLoading || pullRequestPage >= pullRequestTotalPages
              }
              onClick={() => onPullRequestPageChange?.(pullRequestPage + 1)}
            >
              Next
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (activity === "reminder") {
    return (
      <div className="flex min-h-0 flex-1 flex-col">
        <div className="shrink-0 px-3 py-2 text-[11px] font-bold uppercase tracking-wide text-[#bbbbbb]">
          提醒
        </div>
        {onReminderAdd &&
        onReminderUpdate &&
        onReminderDelete &&
        onReminderSelect ? (
          <ReminderCustomSidebar
            items={customReminders}
            selectedId={selectedReminderId}
            onSelect={onReminderSelect}
            onAdd={onReminderAdd}
            onUpdate={onReminderUpdate}
            onDelete={onReminderDelete}
          />
        ) : null}
      </div>
    );
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="px-3 py-2 text-[11px] font-bold uppercase tracking-wide text-[#bbbbbb]">
        Workhour
      </div>
      <div className="allow-select min-h-0 flex-1 space-y-3 overflow-y-auto px-3 pb-3 text-[13px] text-[#cccccc]">
        <div className="flex flex-wrap items-end gap-x-8 gap-y-2">
          <div>
            <div className="mb-1 text-[11px] uppercase text-[#858585]">
              总工时
            </div>
            <div className="font-mono text-[18px] font-semibold tabular-nums text-[#e37933]">
              {workHourTotalEffectiveHours === undefined
                ? "—"
                : `${workHourTotalEffectiveHours.toFixed(2)} h`}
            </div>
          </div>
          <div>
            <div className="mb-1 text-[11px] uppercase text-[#858585]">
              加班
            </div>
            <div className="font-mono text-[18px] font-semibold tabular-nums text-[#e37933]">
              {workHourOvertimeHours === undefined
                ? "—"
                : `${workHourOvertimeHours.toFixed(2)} h`}
            </div>
          </div>
        </div>
        <div>
          <div className="mb-1.5 text-[11px] uppercase text-[#858585]">
            作息
          </div>
          {workHourShiftNameZh.trim() !== "" ? (
            <p className="allow-select break-words font-mono text-[11px] leading-relaxed text-[#858585]">
              {workHourShiftNameZh}
            </p>
          ) : (
            <p className="text-[12px] leading-relaxed text-[#858585]">
              在主区域刷新考勤后，将显示「user-info」接口中的班次说明（shiftNameZh）。
            </p>
          )}
        </div>
      </div>
    </div>
  );
}

export function Sidebar(props: Props) {
  const { activity, width, onResizeStart, ...rest } = props;
  return (
    <aside
      className="relative flex min-h-0 min-w-0 shrink-0 flex-col border-r border-[var(--vscode-border)]"
      style={{ width, background: "var(--vscode-sideBar-bg)" }}
    >
      <SidebarBody {...rest} activity={activity} />
      <button
        type="button"
        aria-label="Resize sidebar"
        className="absolute -right-1 top-0 z-10 h-full w-2 cursor-col-resize hover:bg-[#007fd4]/30"
        onMouseDown={(e) => {
          e.preventDefault();
          onResizeStart();
        }}
      />
    </aside>
  );
}
