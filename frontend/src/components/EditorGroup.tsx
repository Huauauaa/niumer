import { useEffect, useRef, useState } from "react";
import type { AttendanceRecord } from "../types/workhour";
import { JsonFormatterView } from "./JsonFormatterView";
import { MarkdownPreview } from "./MarkdownPreview";
import { WorkHourView } from "./WorkHourView";

export type EditorTab = { id: string; title: string; dirty?: boolean };

type Props = {
  tabs: EditorTab[];
  activeId: string;
  onSelect: (id: string) => void;
  onClose: (id: string) => void;
  /** When true, main area is the blog editor (textarea + line numbers). */
  blogEditor: boolean;
  editorContent: string;
  onEditorContentChange: (value: string) => void;
  /** Cmd/Ctrl+S when blog editor is active */
  onSaveBlog?: () => void | Promise<void>;
  breadcrumbLabel?: string;
  /** Work hour: DB-backed attendance table */
  workHourView?: boolean;
  workHourRecords?: AttendanceRecord[];
  workHourLoading?: boolean;
  workHourError?: string | null;
  onRefreshWorkHour?: () => void;
  /** Tool: JSON formatter (Postman-style body editor). */
  jsonFormatterView?: boolean;
  jsonFormatterContent?: string;
  onJsonFormatterContentChange?: (value: string) => void;
};

const sampleLines = [
  'import { useState } from "react";',
  "",
  "export function Welcome() {",
  "  const [open, setOpen] = useState(true);",
  "  return (",
  '    <section className="p-4">',
  '      <h1 className="text-lg">Welcome to niumer</h1>',
  '      <p className="text-sm text-[#858585]">',
  "        Layout mirrors VS Code: activity bar, sidebar, panel, status bar.",
  "      </p>",
  "    </section>",
  "  );",
  "}",
];

export function EditorGroup({
  tabs,
  activeId,
  onSelect,
  onClose,
  blogEditor,
  editorContent,
  onEditorContentChange,
  onSaveBlog,
  breadcrumbLabel,
  workHourView,
  workHourRecords = [],
  workHourLoading = false,
  workHourError = null,
  onRefreshWorkHour,
  jsonFormatterView = false,
  jsonFormatterContent = "",
  onJsonFormatterContentChange,
}: Props) {
  const active = tabs.find((t) => t.id === activeId) ?? tabs[0];
  const taRef = useRef<HTMLTextAreaElement>(null);
  const gutterRef = useRef<HTMLDivElement>(null);
  const [blogPreviewOpen, setBlogPreviewOpen] = useState(true);

  useEffect(() => {
    if (!blogEditor || !onSaveBlog) return;
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "s") {
        e.preventDefault();
        void onSaveBlog();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [blogEditor, onSaveBlog]);

  const lines = blogEditor ? editorContent.split("\n") : sampleLines;
  const lineCount = Math.max(1, lines.length);

  const showBlogEmpty = blogEditor && tabs.length === 0;

  return (
    <div
      className="flex min-h-0 min-w-0 flex-1 flex-col"
      style={{ background: "var(--vscode-editor-bg)" }}
    >
      <div
        className="flex h-9 shrink-0 items-end gap-px overflow-x-auto border-b border-[var(--vscode-border)] bg-[#252526]"
        role="tablist"
      >
        {tabs.map((tab) => {
          const isActive = tab.id === activeId;
          return (
            <div
              key={tab.id}
              role="tab"
              aria-selected={isActive}
              className={`group flex h-9 min-w-0 max-w-[200px] shrink-0 items-center gap-1 border-r border-[var(--vscode-border)] px-2 text-[13px] ${
                isActive
                  ? "border-t-2 border-t-[#007fd4] bg-[var(--vscode-editor-bg)] text-[#ffffff]"
                  : "cursor-pointer bg-[var(--vscode-tab-inactive)] text-[#969696] hover:bg-[#2a2d2e]"
              }`}
              onClick={() => onSelect(tab.id)}
            >
              <span className="truncate">
                {tab.title}
                {tab.dirty ? " \u2022" : ""}
              </span>
              <button
                type="button"
                className="ml-0.5 rounded p-0.5 opacity-0 hover:bg-white/10 group-hover:opacity-100"
                aria-label={`Close ${tab.title}`}
                onClick={(e) => {
                  e.stopPropagation();
                  onClose(tab.id);
                }}
              >
                ×
              </button>
            </div>
          );
        })}
      </div>

      <div className="flex min-h-0 flex-1 flex-col">
        <div className="flex h-6 shrink-0 items-center gap-2 border-b border-[var(--vscode-border)] px-4 text-[12px] text-[#cccccc] opacity-80">
          <span className="min-w-0 flex-1 truncate">
            {blogEditor ? (
              <>
                blog ›{" "}
                <span className="text-[#e37933]">
                  {breadcrumbLabel ?? active?.title ?? "untitled"}
                </span>
              </>
            ) : workHourView ? (
              <>
                workhour › <span className="text-[#e37933]">attendance</span>
              </>
            ) : jsonFormatterView ? (
              <>
                tool › <span className="text-[#e37933]">JSON formatter</span>
              </>
            ) : (
              <>
                niumer ›{" "}
                <span className="text-[#e37933]">
                  {active?.title ?? "Home"}
                </span>
              </>
            )}
          </span>
          {blogEditor && !showBlogEmpty ? (
            <button
              type="button"
              className="shrink-0 rounded px-1.5 py-0.5 text-[11px] text-[#cccccc] opacity-90 hover:bg-white/10 hover:opacity-100"
              title={
                blogPreviewOpen ? "隐藏 Markdown 预览" : "打开 Markdown 预览"
              }
              onClick={() => setBlogPreviewOpen((v) => !v)}
            >
              {blogPreviewOpen ? "隐藏预览" : "打开预览"}
            </button>
          ) : null}
        </div>

        {workHourView ? (
          <WorkHourView
            records={workHourRecords}
            loading={workHourLoading}
            error={workHourError}
            onRefresh={onRefreshWorkHour ?? (() => {})}
          />
        ) : showBlogEmpty ? (
          <div className="allow-select flex flex-1 items-center justify-center px-6 text-center text-[13px] text-[#858585]">
            No open document. Create a new file or open one from the Blog
            explorer.
          </div>
        ) : jsonFormatterView && onJsonFormatterContentChange ? (
          <JsonFormatterView
            value={jsonFormatterContent}
            onChange={onJsonFormatterContentChange}
          />
        ) : blogEditor ? (
          <div className="flex min-h-0 flex-1 overflow-hidden">
            <div className="allow-select flex min-h-0 min-w-0 flex-1 overflow-hidden font-mono text-[13px] leading-6">
              <div
                ref={gutterRef}
                className="min-w-[3rem] shrink-0 select-none overflow-y-auto overflow-x-hidden border-r border-[var(--vscode-border)] bg-[#1e1e1e] py-2 pl-3 pr-3 text-right text-[#858585]"
              >
                {Array.from({ length: lineCount }, (_, i) => (
                  <div key={i}>{i + 1}</div>
                ))}
              </div>
              <textarea
                ref={taRef}
                className="min-h-0 min-w-0 flex-1 resize-none overflow-y-auto border-0 bg-[#1e1e1e] p-2 font-mono text-[13px] leading-6 text-[#d4d4d4] caret-[#aeafad] outline-none focus:ring-0"
                spellCheck={false}
                value={editorContent}
                onChange={(e) => onEditorContentChange(e.target.value)}
                onScroll={(e) => {
                  if (gutterRef.current)
                    gutterRef.current.scrollTop = e.currentTarget.scrollTop;
                }}
                placeholder="Start typing…"
              />
            </div>
            {blogPreviewOpen ? (
              <>
                <div
                  className="w-px shrink-0 bg-[var(--vscode-border)]"
                  aria-hidden
                />
                <MarkdownPreview markdown={editorContent} />
              </>
            ) : null}
          </div>
        ) : (
          <div className="allow-select flex min-h-0 flex-1 overflow-auto font-mono text-[13px] leading-6">
            <div className="sticky left-0 shrink-0 select-none border-r border-[var(--vscode-border)] bg-[#1e1e1e] py-2 pl-4 pr-3 text-right text-[#858585]">
              {sampleLines.map((_, i) => (
                <div key={i}>{i + 1}</div>
              ))}
            </div>
            <pre className="m-0 flex-1 py-2 pl-4 text-[#d4d4d4]">
              <code>
                {sampleLines.map((line, i) => (
                  <div key={i}>{line || " "}</div>
                ))}
              </code>
            </pre>
          </div>
        )}
      </div>
    </div>
  );
}
