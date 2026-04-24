import { useCallback, useRef, useState } from "react";
import {
  parseJsonOrJsLiteral,
  stringifyMinified,
  stringifyPretty,
} from "../utils/jsonFormat";

type Props = {
  value: string;
  onChange: (next: string) => void;
};

export function JsonFormatterView({ value, onChange }: Props) {
  const taRef = useRef<HTMLTextAreaElement>(null);
  const gutterRef = useRef<HTMLDivElement>(null);
  const [error, setError] = useState<string | null>(null);

  const lines = value.split("\n");
  const lineCount = Math.max(1, lines.length);

  const applyFormatted = useCallback(
    (pretty: boolean) => {
      try {
        const parsed = parseJsonOrJsLiteral(value);
        const next = pretty
          ? stringifyPretty(parsed)
          : stringifyMinified(parsed);
        onChange(next);
        setError(null);
      } catch (e) {
        setError(e instanceof Error ? e.message : String(e));
      }
    },
    [value, onChange],
  );

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
      <div className="flex shrink-0 items-center justify-end gap-4 border-b border-[var(--vscode-border)] bg-[var(--vscode-editor-bg)] px-4 py-1.5">
        <button
          type="button"
          className="text-[13px] text-[#3794ff] hover:underline"
          onClick={() => applyFormatted(true)}
        >
          Beautify
        </button>
        <button
          type="button"
          className="text-[13px] text-[var(--vscode-fg)] hover:text-[var(--vscode-tab-active-fg)] hover:underline"
          onClick={() => applyFormatted(false)}
        >
          Minify
        </button>
      </div>
      {error ? (
        <div className="allow-select shrink-0 border-b border-[#5a1d1d] bg-[#3c1c1c] px-4 py-1.5 text-[12px] text-[#f48771]">
          {error}
        </div>
      ) : null}
      <div className="allow-select flex min-h-0 flex-1 overflow-hidden font-mono text-[13px] leading-[22px]">
        <div
          ref={gutterRef}
          className="min-w-[3rem] shrink-0 select-none overflow-y-auto overflow-x-hidden border-r border-[var(--vscode-border)] bg-[var(--vscode-gutter-bg)] py-2 pl-3 pr-3 text-right text-[var(--vscode-fg-muted)]"
        >
          {Array.from({ length: lineCount }, (_, i) => (
            <div key={i}>{i + 1}</div>
          ))}
        </div>
        <textarea
          ref={taRef}
          className="min-h-0 min-w-0 flex-1 resize-none overflow-y-auto border-0 bg-[var(--vscode-editor-bg)] p-2 font-mono text-[13px] leading-[22px] text-[var(--vscode-editor-fg)] caret-[var(--vscode-caret)] outline-none focus:ring-0"
          spellCheck={false}
          value={value}
          onChange={(e) => {
            setError(null);
            onChange(e.target.value);
          }}
          onScroll={(e) => {
            if (gutterRef.current) {
              gutterRef.current.scrollTop = e.currentTarget.scrollTop;
            }
          }}
          placeholder='Paste JSON, or a JS object / array literal (e.g. { "a": 1 } or { a: 1 })…'
          aria-label="JSON or JavaScript literal input"
        />
      </div>
    </div>
  );
}
