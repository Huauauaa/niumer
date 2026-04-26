type Props = {
  /** Cached work-hour tenant user id (local SQLite). */
  account?: string | null;
  onAccountClick?: () => void;
};

/** VS Code–style status bar: shows signed-in user id when known. */
export function StatusBar({ account, onAccountClick }: Props) {
  const trimmed = (account ?? "").trim();
  return (
    <footer
      className="flex h-[22px] shrink-0 items-center border-t border-[#007acc] px-2 text-[11px] text-white/95"
      style={{ background: "var(--vscode-statusBar-bg)" }}
    >
      {trimmed ? (
        <button
          type="button"
          className="max-w-[min(100%,28rem)] truncate rounded px-1 py-0.5 font-mono hover:bg-white/15 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-white/80"
          title="User info"
          aria-label={`User info, ${trimmed}`}
          onClick={onAccountClick}
        >
          {trimmed}
        </button>
      ) : (
        <span className="select-none font-mono text-white/55">—</span>
      )}
    </footer>
  );
}
