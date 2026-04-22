import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

type Props = {
  markdown: string;
};

/** Renders blog Markdown with VS Code–like preview chrome (see `.vscode-markdown-body` in index.css). */
export function MarkdownPreview({ markdown }: Props) {
  return (
    <div className="vscode-markdown-preview allow-select min-h-0 min-w-0 flex-1 overflow-auto">
      <div className="vscode-markdown-body">
        <ReactMarkdown remarkPlugins={[remarkGfm]}>{markdown || "\u00a0"}</ReactMarkdown>
      </div>
    </div>
  );
}
