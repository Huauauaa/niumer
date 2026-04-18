/** One Markdown file in the blog working directory. `fileName` is the on-disk name (e.g. Welcome.md). */
export type BlogDocument = {
  fileName: string;
  title: string;
  content: string;
  dirty?: boolean;
};

/** Ensures `.md` suffix; trims whitespace. */
export function normalizeMarkdownFileName(input: string): string {
  const t = input.trim();
  if (!t) return "";
  return /\.md$/i.test(t) ? t : `${t}.md`;
}

/** Picks Untitled-1.md, Untitled-2.md, … not present in `existing` (case-insensitive). */
export function nextUntitledFileName(existing: string[]): string {
  const lower = new Set(existing.map((s) => s.toLowerCase()));
  let n = 1;
  let name = `Untitled-${n}.md`;
  while (lower.has(name.toLowerCase())) {
    n += 1;
    name = `Untitled-${n}.md`;
  }
  return name;
}
