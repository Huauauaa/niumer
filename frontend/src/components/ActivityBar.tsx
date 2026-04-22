import {
  IconBlog,
  IconPullRequest,
  IconReminder,
  IconTool,
  IconWorkhour,
} from "./icons";

export type ActivityId =
  | "blog"
  | "tool"
  | "pullRequest"
  | "workhour"
  | "reminder";

type Props = {
  active: ActivityId;
  onChange: (id: ActivityId) => void;
};

const entries: { id: ActivityId; label: string; Icon: typeof IconBlog }[] = [
  { id: "blog", label: "Blog", Icon: IconBlog },
  { id: "tool", label: "Tool", Icon: IconTool },
  { id: "pullRequest", label: "Pull Request", Icon: IconPullRequest },
  { id: "workhour", label: "Workhour", Icon: IconWorkhour },
  { id: "reminder", label: "提醒", Icon: IconReminder },
];

export function ActivityBar({ active, onChange }: Props) {
  return (
    <nav
      className="flex w-12 shrink-0 flex-col items-center gap-0.5 border-r border-[var(--vscode-border)] py-2"
      style={{ background: "var(--vscode-activityBar-bg)" }}
      aria-label="Activity bar"
    >
      {entries.map(({ id, label, Icon }) => {
        const isActive = active === id;
        return (
          <button
            key={id}
            type="button"
            title={label}
            aria-label={label}
            aria-current={isActive ? "page" : undefined}
            onClick={() => onChange(id)}
            className="relative flex h-12 w-12 items-center justify-center text-[#858585] hover:text-white focus:outline-none focus-visible:ring-1 focus-visible:ring-white/40"
          >
            {isActive && (
              <span
                className="absolute left-0 top-2 h-8 w-0.5 rounded-r bg-white"
                aria-hidden
              />
            )}
            <Icon className={isActive ? "text-white" : undefined} />
          </button>
        );
      })}
    </nav>
  );
}
