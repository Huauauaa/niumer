import type { PullRequestListResponse } from "../types/pullRequest";

/** Mock server base; override with `VITE_PULL_REQUEST_API_BASE` in `.env` / Vite. */
export function pullRequestApiBase(): string {
  const b = import.meta.env.VITE_PULL_REQUEST_API_BASE as string | undefined;
  return (b && b.trim()) || "http://127.0.0.1:17890";
}

export async function fetchPullRequests(
  page: number,
  pageSize: number,
): Promise<PullRequestListResponse> {
  const base = pullRequestApiBase().replace(/\/$/, "");
  const u = new URL(`${base}/pull-request`);
  u.searchParams.set("page", String(page));
  u.searchParams.set("page_size", String(pageSize));
  const res = await fetch(u.toString(), { method: "GET" });
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }
  return (await res.json()) as PullRequestListResponse;
}
