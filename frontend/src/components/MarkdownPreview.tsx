import { useEffect, useId, useMemo, useState } from "react";
import mermaid from "mermaid";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

type Props = {
  markdown: string;
};

function MermaidBlock({ code }: { code: string }) {
  const reactId = useId();
  const id = useMemo(() => `mermaid-${reactId.replace(/[:]/g, "")}`, [reactId]);
  const [svg, setSvg] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setSvg(null);
    setErr(null);

    // Initialize once per mount (idempotent).
    try {
      mermaid.initialize({
        startOnLoad: false,
        securityLevel: "strict",
        theme: "dark",
      });
    } catch {
      // ignore
    }

    void (async () => {
      try {
        const { svg } = await mermaid.render(id, code);
        if (!cancelled) setSvg(svg);
      } catch (e) {
        if (!cancelled) setErr(e instanceof Error ? e.message : String(e));
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [code, id]);

  if (err) {
    return (
      <pre className="mermaid-error">
        <code>{err}</code>
      </pre>
    );
  }
  if (!svg) {
    return (
      <pre className="mermaid-loading">
        <code>Rendering mermaid…</code>
      </pre>
    );
  }
  return <div className="mermaid" dangerouslySetInnerHTML={{ __html: svg }} />;
}

/** Renders blog Markdown with VS Code–like preview chrome (see `.vscode-markdown-body` in index.css). */
export function MarkdownPreview({ markdown }: Props) {
  return (
    <div className="vscode-markdown-preview allow-select min-h-0 min-w-0 flex-1 overflow-auto">
      <div className="vscode-markdown-body">
        <ReactMarkdown
          remarkPlugins={[remarkGfm]}
          components={{
            code({ className, children, ...props }) {
              const m = /language-(\w+)/.exec(className || "");
              const lang = (m?.[1] || "").toLowerCase();
              const raw = String(children ?? "");
              if (lang === "mermaid") {
                return <MermaidBlock code={raw.replace(/\n$/, "")} />;
              }
              return (
                <code className={className} {...props}>
                  {children}
                </code>
              );
            },
          }}
        >
          {markdown || "\u00a0"}
        </ReactMarkdown>
      </div>
    </div>
  );
}
