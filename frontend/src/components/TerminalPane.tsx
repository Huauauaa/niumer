import { FitAddon } from "@xterm/addon-fit";
import { Terminal } from "@xterm/xterm";
import "@azurity/pure-nerd-font/pure-nerd-font.css";
import "@xterm/xterm/css/xterm.css";
import { useEffect, useRef } from "react";
import {
  TerminalResize,
  TerminalStart,
  TerminalStop,
  TerminalWrite,
} from "../../wailsjs/go/main/App";
import { EventsOn } from "../../wailsjs/runtime/runtime";

type Props = {
  /** Bottom panel is expanded. */
  panelOpen: boolean;
  /** Terminal tab is active (keyboard goes to shell). */
  shellAttention: boolean;
};

function hasWailsApp(): boolean {
  return (
    typeof window !== "undefined" &&
    !!(window as unknown as { go?: { main?: { App?: unknown } } }).go?.main
      ?.App
  );
}

export function TerminalPane({ panelOpen, shellAttention }: Props) {
  const hostRef = useRef<HTMLDivElement>(null);
  const termRef = useRef<Terminal | null>(null);
  const fitRef = useRef<FitAddon | null>(null);
  const attentionRef = useRef(shellAttention);
  attentionRef.current = shellAttention;

  useEffect(() => {
    if (!panelOpen) return;
    const host = hostRef.current;
    if (!host) return;

    /** Symbols font first so Powerline / Nerd glyphs resolve; ASCII falls back to Menlo. */
    const fontFamily =
      'PureNerdFont, "Symbols Nerd Font Mono", Menlo, Monaco, "Courier New", monospace';

    const term = new Terminal({
      cursorBlink: true,
      fontSize: 12,
      lineHeight: 1.0,
      fontFamily,
      theme: {
        background: "#181818",
        foreground: "#cccccc",
        cursor: "#aeafad",
        black: "#181818",
        brightBlack: "#333333",
        red: "#f48771",
        brightRed: "#f48771",
        green: "#6a9955",
        brightGreen: "#6a9955",
        yellow: "#dcdcaa",
        brightYellow: "#dcdcaa",
        blue: "#569cd6",
        brightBlue: "#569cd6",
        magenta: "#c586c0",
        brightMagenta: "#c586c0",
        cyan: "#4ec9b0",
        brightCyan: "#4ec9b0",
        white: "#cccccc",
        brightWhite: "#ffffff",
      },
    });
    const fit = new FitAddon();
    term.loadAddon(fit);
    term.open(host);
    termRef.current = term;
    fitRef.current = fit;

    const unsubOut = EventsOn("terminal:output", (...args: unknown[]) => {
      const b64 = args[0];
      const t = termRef.current;
      if (typeof b64 !== "string" || !t) return;
      try {
        const raw = atob(b64);
        const u8 = new Uint8Array(raw.length);
        for (let i = 0; i < raw.length; i++) u8[i] = raw.charCodeAt(i);
        t.write(u8);
      } catch {
        /* ignore invalid chunk */
      }
    });

    const unsubErr = EventsOn("terminal:error", (...args: unknown[]) => {
      const t = termRef.current;
      const msg = typeof args[0] === "string" ? args[0] : String(args[0] ?? "");
      if (t && msg) {
        t.write(`\r\n\x1b[31m${msg}\x1b[0m\r\n`);
      }
    });

    term.onData((data) => {
      if (!attentionRef.current) return;
      if (!hasWailsApp()) return;
      void TerminalWrite(data);
    });

    const fitAndResize = () => {
      const t = termRef.current;
      const f = fitRef.current;
      if (!t || !f || !hostRef.current) return;
      try {
        f.fit();
      } catch {
        /* zero-sized during layout */
      }
      if (hasWailsApp() && t.cols > 0 && t.rows > 0) {
        void TerminalResize(t.cols, t.rows);
      }
    };

    const ro = new ResizeObserver(() => {
      fitAndResize();
    });
    ro.observe(host);

    let disposed = false;
    void TerminalStart("")
      .then(() => {
        if (disposed) return;
        requestAnimationFrame(() => {
          requestAnimationFrame(() => fitAndResize());
        });
      })
      .catch((e: unknown) => {
        const msg = e instanceof Error ? e.message : String(e);
        term.writeln(`\x1b[31mCould not start shell:\x1b[0m ${msg}`);
      });

    return () => {
      disposed = true;
      unsubOut();
      unsubErr();
      ro.disconnect();
      term.dispose();
      termRef.current = null;
      fitRef.current = null;
      if (hasWailsApp()) {
        void TerminalStop();
      }
    };
  }, [panelOpen]);

  return (
    <div className="min-h-0 flex-1 overflow-hidden px-1 pb-1 pt-0">
      <div ref={hostRef} className="h-full min-h-[120px] w-full" />
    </div>
  );
}
