import { useCallback, useEffect, useState } from "react";
import { GetAboutInfo } from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";

type Props = {
  open: boolean;
  onClose: () => void;
};

function fallbackAbout(): main.AboutView {
  return main.AboutView.createFrom({
    appName: "niumer",
    version: "0.0.0",
    commit: "",
    buildTime: "",
    goVersion: "",
    wailsVersion: "v2.12.0",
    osArch: typeof navigator !== "undefined" ? navigator.platform : "",
  });
}

function formatAboutLines(a: main.AboutView): string {
  const lines = [
    `Version: ${a.version}`,
    `Commit: ${a.commit || "—"}`,
    `Date: ${a.buildTime || "—"}`,
    `Go: ${a.goVersion || "—"}`,
    `Wails: ${a.wailsVersion || "—"}`,
    `OS: ${a.osArch || "—"}`,
  ];
  return `${a.appName}\n\n${lines.join("\n")}`;
}

export function AboutDialog({ open, onClose }: Props) {
  const [info, setInfo] = useState<main.AboutView | null>(null);
  const [copyDone, setCopyDone] = useState(false);

  const load = useCallback(async () => {
    try {
      const raw = await GetAboutInfo();
      setInfo(main.AboutView.createFrom(raw));
    } catch {
      setInfo(fallbackAbout());
    }
  }, []);

  useEffect(() => {
    if (!open) return;
    setCopyDone(false);
    void load();
  }, [open, load]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onClose]);

  const handleCopy = async () => {
    if (!info) return;
    try {
      await navigator.clipboard.writeText(formatAboutLines(info));
      setCopyDone(true);
      window.setTimeout(() => setCopyDone(false), 2000);
    } catch {
      setCopyDone(false);
    }
  };

  if (!open) return null;

  const a = info ?? fallbackAbout();

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center bg-black/55 p-4"
      role="presentation"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        className="w-full max-w-[420px] rounded-2xl border border-[var(--vscode-border)] bg-[var(--vscode-dialog-bg)] px-10 py-10 text-center shadow-2xl"
        style={{ boxShadow: "0 24px 80px rgba(0,0,0,0.35)" }}
        role="dialog"
        aria-labelledby="about-title"
        aria-describedby="about-details"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <div className="mx-auto mb-5 flex h-20 w-20 items-center justify-center rounded-2xl bg-gradient-to-br from-[#4b6cb7] to-[#182848] shadow-lg">
          <span className="text-[28px] font-semibold tracking-tight text-white">
            N
          </span>
        </div>
        <h2
          id="about-title"
          className="m-0 mb-6 text-[22px] font-semibold tracking-tight text-[var(--vscode-fg-heading)]"
        >
          {a.appName}
        </h2>
        <div
          id="about-details"
          className="space-y-1.5 text-left text-[13px] leading-relaxed text-[var(--vscode-fg)]"
        >
          <p>
            <span className="text-[var(--vscode-fg-muted)]">Version: </span>
            {a.version}
          </p>
          <p className="break-all">
            <span className="text-[var(--vscode-fg-muted)]">Commit: </span>
            {a.commit || "—"}
          </p>
          <p>
            <span className="text-[var(--vscode-fg-muted)]">Date: </span>
            {a.buildTime || "—"}
          </p>
          <p>
            <span className="text-[var(--vscode-fg-muted)]">Go: </span>
            {a.goVersion || "—"}
          </p>
          <p>
            <span className="text-[var(--vscode-fg-muted)]">Wails: </span>
            {a.wailsVersion || "—"}
          </p>
          <p>
            <span className="text-[var(--vscode-fg-muted)]">OS: </span>
            {a.osArch || "—"}
          </p>
        </div>
        <div className="mt-10 flex justify-center gap-3">
          <button
            type="button"
            className="min-w-[100px] rounded-lg border border-[#007fd4]/50 bg-[var(--vscode-input-bg)] px-5 py-2 text-[13px] font-medium text-[var(--vscode-fg)] shadow-sm hover:bg-[var(--vscode-button-hover)] focus:outline-none focus-visible:ring-2 focus-visible:ring-[#007fd4]"
            onClick={onClose}
          >
            OK
          </button>
          <button
            type="button"
            className="min-w-[100px] rounded-lg bg-[#0e639c] px-5 py-2 text-[13px] font-medium text-white shadow-sm hover:bg-[#1177bb] focus:outline-none focus-visible:ring-2 focus-visible:ring-[#4fc1ff]"
            onClick={() => void handleCopy()}
          >
            {copyDone ? "Copied" : "Copy"}
          </button>
        </div>
      </div>
    </div>
  );
}
