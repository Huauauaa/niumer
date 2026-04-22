import { useState } from "react";

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

  if (!visible) {
    return (
      <button
        type="button"
        className="flex h-6 w-full shrink-0 items-center justify-center border-t border-[var(--vscode-border)] bg-[var(--vscode-panel-bg)] text-[11px] text-[#cccccc] hover:bg-[#2a2d2e]"
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

  return (
    <div
      className="flex shrink-0 flex-col border-t border-[var(--vscode-border)]"
      style={{ height, background: "var(--vscode-panel-bg)" }}
    >
      <div className="relative flex h-8 shrink-0 items-center gap-3 border-b border-[var(--vscode-border)] px-2 text-[11px] font-bold uppercase tracking-wide text-[#cccccc]">
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
                ? "border-b-2 border-[#007fd4] text-white"
                : "text-[#969696] hover:text-white"
            }`}
            onClick={() => setTab(t.id)}
          >
            {t.label}
          </button>
        ))}
        <div className="ml-auto flex items-center gap-1 pr-1">
          <button
            type="button"
            className="rounded px-2 py-0.5 hover:bg-white/10"
            onClick={onToggle}
            aria-label="Hide panel"
          >
            ▼
          </button>
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-auto p-2 font-mono text-[12px] leading-relaxed text-[#cccccc]">
        {tab === "terminal" && (
          <div>
            <div className="text-[#6a9955]">$ npm run build</div>
            <div className="text-[#d4d4d4]">
              &gt; niumer-frontend@0.0.0 build
            </div>
            <div className="text-[#d4d4d4]">&gt; vite build</div>
            <div className="text-[#569cd6]">
              vite v5.x building for production...
            </div>
            <div className="text-[#6a9955]">✓ built in 420ms</div>
            <div className="mt-2 flex items-center gap-1">
              <span className="text-[#cccccc]">$</span>
              <span className="animate-pulse">▌</span>
            </div>
          </div>
        )}
        {tab === "problems" && (
          <div className="text-[#858585]">No problems have been detected.</div>
        )}
        {tab === "output" && (
          <div className="text-[#858585]">Output channel is empty.</div>
        )}
      </div>
    </div>
  );
}
