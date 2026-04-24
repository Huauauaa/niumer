type Props = {
  onOpenPreference: () => void;
  onOpenUserInfo: () => void;
};

export function MenuBar({ onOpenPreference, onOpenUserInfo }: Props) {
  return (
    <header
      className="flex h-[30px] shrink-0 items-center gap-1 border-b border-[var(--vscode-border)] px-2 text-[13px] text-[var(--vscode-fg)]"
      style={{ background: "var(--vscode-menuBar-bg)" }}
    >
      <button
        type="button"
        className="rounded px-2 py-0.5 text-[var(--vscode-fg)] hover:bg-[var(--vscode-menu-hover)] focus:outline-none focus-visible:ring-1 focus-visible:ring-[var(--vscode-focus-ring)]"
        onClick={onOpenPreference}
      >
        Preference
      </button>
      <button
        type="button"
        className="rounded px-2 py-0.5 text-[var(--vscode-fg)] hover:bg-[var(--vscode-menu-hover)] focus:outline-none focus-visible:ring-1 focus-visible:ring-[var(--vscode-focus-ring)]"
        onClick={onOpenUserInfo}
      >
        Profile
      </button>
    </header>
  );
}
