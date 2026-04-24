import { useEffect, useRef, useState } from "react";

type Props = {
  onOpenPreference: () => void;
  onOpenUserInfo: () => void;
};

export function MenuBar({ onOpenPreference, onOpenUserInfo }: Props) {
  const [prefOpen, setPrefOpen] = useState(false);
  const wrapRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const onDoc = (e: MouseEvent) => {
      if (wrapRef.current && !wrapRef.current.contains(e.target as Node)) {
        setPrefOpen(false);
      }
    };
    document.addEventListener("mousedown", onDoc);
    return () => document.removeEventListener("mousedown", onDoc);
  }, []);

  return (
    <header
      className="flex h-[30px] shrink-0 items-center gap-1 border-b border-[var(--vscode-border)] px-2 text-[13px] text-[#cccccc]"
      style={{ background: "var(--vscode-menuBar-bg)" }}
    >
      <div className="relative" ref={wrapRef}>
        <button
          type="button"
          className={`rounded px-2 py-0.5 hover:bg-white/10 focus:outline-none focus-visible:ring-1 focus-visible:ring-white/30 ${
            prefOpen ? "bg-white/10" : ""
          }`}
          aria-expanded={prefOpen}
          aria-haspopup="menu"
          onClick={() => setPrefOpen((o) => !o)}
        >
          Preference
        </button>
        {prefOpen && (
          <div
            className="absolute left-0 top-full z-50 min-w-[160px] border border-[var(--vscode-border)] bg-[#252526] py-0.5 shadow-lg"
            role="menu"
          >
            <button
              type="button"
              role="menuitem"
              data-i18n="preference.preference"
              className="block w-full px-3 py-1.5 text-left text-[13px] text-[#cccccc] hover:bg-[#094771]"
              onClick={() => {
                setPrefOpen(false);
                onOpenPreference();
              }}
            >
              Preference
            </button>
          </div>
        )}
      </div>
      <button
        type="button"
        className="rounded px-2 py-0.5 text-[#cccccc] hover:bg-white/10 focus:outline-none focus-visible:ring-1 focus-visible:ring-white/30"
        onClick={onOpenUserInfo}
      >
        User info
      </button>
    </header>
  );
}
