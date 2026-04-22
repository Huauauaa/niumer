export type PullRequestState = "open" | "merged" | "closed";

export type PullRequestListItem = {
  id: number;
  number: number;
  url: string;
  title: string;
  author: string;
  sourceBranch: string;
  targetBranch: string;
  state: PullRequestState;
  createdAt: string;
  updatedAt: string;
};

export type PullRequestListResponse = {
  items: PullRequestListItem[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
};
