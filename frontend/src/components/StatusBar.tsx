/** VS Code–style status bar shell; content intentionally empty for now. */
export function StatusBar() {
  return (
    <footer
      className="h-[22px] shrink-0 border-t border-[#007acc]"
      style={{ background: "var(--vscode-statusBar-bg)" }}
      aria-hidden
    />
  );
}
