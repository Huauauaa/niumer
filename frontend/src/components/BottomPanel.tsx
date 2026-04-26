import { useLayoutEffect, useState } from "react";
import { TerminalPane } from "./TerminalPane";

type PanelTab = "terminal" | "problems" | "output";

type Props = {
  height: number;
  visible: boolean;
  onResizeStart: () => void;
  onToggle: () => void;
};

export function BottomPanel({
  height,
  visible,
  onResizeStart,
  onToggle,
}: Props) {
  const [tab, setTab] = useState<PanelTab>("terminal");
  /** After first visit to Terminal while panel is open, keep PTY when switching tabs. */
  const [terminalWarm, setTerminalWarm] = useState(false);

  useLayoutEffect(() => {
    if (!visible) {
      setTerminalWarm(false);
      return;
    }
    if (tab === "terminal") setTerminalWarm(true);
  }, [visible, tab]);

  if (!visible) {
    return (
      <button
        type="button"
        className="flex h-6 w-full shrink-0 items-center justify-center border-t border-[var(--vscode-border)] bg-[var(--vscode-panel-bg)] text-[11px] text-[var(--vscode-fg)] hover:bg-[var(--vscode-list-hover)]"
        onClick={onToggle}
      >
        ▲ Panel (show)
      </button>
    );
  }

  const tabs: { id: PanelTab; label: string }[] = [
    { id: "terminal", label: "TERMINAL" },
    { id: "problems", label: "PROBLEMS" },
    { id: "output", label: "OUTPUT" },
  ];

  const tabBlock = (active: boolean) =>
    active
      ? "absolute inset-0 flex min-h-0 flex-col overflow-hidden"
      : "pointer-events-none invisible absolute inset-0 flex min-h-0 flex-col overflow-hidden";

  return (
    <div
      className="flex shrink-0 flex-col border-t border-[var(--vscode-border)]"
      style={{ height, background: "var(--vscode-panel-bg)" }}
    >
      <div className="relative flex h-8 shrink-0 items-center gap-3 border-b border-[var(--vscode-border)] px-2 text-[11px] font-bold uppercase tracking-wide text-[var(--vscode-fg)]">
        <button
          type="button"
          aria-label="Resize panel"
          className="absolute -top-1 left-0 right-0 z-10 h-2 cursor-row-resize hover:bg-[#007fd4]/25"
          onMouseDown={(e) => {
            e.preventDefault();
            onResizeStart();
          }}
        />
        {tabs.map((t) => (
          <button
            key={t.id}
            type="button"
            className={`rounded px-2 py-1 ${
              tab === t.id
                ? "border-b-2 border-[#007fd4] text-[var(--vscode-panel-tab-active-fg)]"
                : "text-[var(--vscode-panel-tab-inactive-fg)] hover:text-[var(--vscode-panel-tab-active-fg)]"
            }`}
            onClick={() => setTab(t.id)}
          >
            {t.label}
          </button>
        ))}
        <div className="ml-auto flex items-center gap-1 pr-1">
          <button
            type="button"
            className="rounded px-2 py-0.5 hover:bg-[var(--vscode-menu-hover)]"
            onClick={onToggle}
            aria-label="Hide panel"
          >
            ▼
          </button>
        </div>
      </div>

      <div className="relative min-h-0 flex-1">
        {terminalWarm ? (
          <div className={tabBlock(tab === "terminal")}>
            <TerminalPane
              panelOpen={visible}
              shellAttention={tab === "terminal" && visible}
            />
          </div>
        ) : null}

        <div
          className={`${tabBlock(tab === "problems")} overflow-auto p-2 text-[12px] leading-relaxed text-[var(--vscode-fg)]`}
        >
          <div className="text-[var(--vscode-fg-muted)]">
            No problems have been detected.
          </div>
        </div>

        <div
          className={`${tabBlock(tab === "output")} overflow-auto p-2 text-[12px] leading-relaxed text-[var(--vscode-fg)]`}
        >
          <div className="text-[var(--vscode-fg-muted)]">
            Output channel is empty.
          </div>
        </div>
      </div>
    </div>
  );
}
