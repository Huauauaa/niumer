import type { AttendanceRecord } from "../types/workhour";

type Props = {
  records: AttendanceRecord[];
  loading: boolean;
  error: string | null;
  dbPath: string;
  onRefresh: () => void;
};

const COLUMNS: { key: keyof AttendanceRecord; label: string }[] = [
  { key: "id", label: "id" },
  { key: "attendanceDate", label: "考勤日期" },
  { key: "clockInTime", label: "打卡时间" },
  { key: "attendanceStatus", label: "状态" },
  { key: "workDay", label: "工作日" },
  { key: "clockInType", label: "类型" },
  { key: "hrId", label: "hrId" },
  { key: "clockInDate", label: "clockInDate" },
  { key: "dayId", label: "dayId" },
  { key: "clockingInSequenceNumber", label: "序号" },
  { key: "creationDate", label: "creationDate" },
  { key: "createdBy", label: "createdBy" },
  { key: "lastUpdateDate", label: "lastUpdateDate" },
  { key: "lastUpdatedBy", label: "lastUpdatedBy" },
  { key: "originalId", label: "originalId" },
  { key: "dataSource", label: "dataSource" },
  { key: "clockInReason", label: "clockInReason" },
  { key: "earlyClockInTime", label: "earlyClockInTime" },
  { key: "lateClockInTime", label: "lateClockInTime" },
  { key: "earlyClockInType", label: "earlyClockInType" },
  { key: "lateClockInType", label: "lateClockInType" },
  { key: "minuteNumber", label: "minuteNumber" },
  { key: "hourNumber", label: "hourNumber" },
  { key: "attendProcessId", label: "attendProcessId" },
  { key: "attendanceStatusCode", label: "attendanceStatusCode" },
  { key: "earlyClockInReason", label: "earlyClockInReason" },
  { key: "lateClockInReason", label: "lateClockInReason" },
  { key: "earlyClockTag", label: "earlyClockTag" },
  { key: "lateClockTag", label: "lateClockTag" },
];

function cellValue(r: AttendanceRecord, key: keyof AttendanceRecord): string {
  const v = r[key];
  if (v === null || v === undefined) return "";
  return String(v);
}

export function WorkHourView({ records, loading, error, dbPath, onRefresh }: Props) {
  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="flex shrink-0 items-center justify-between gap-3 border-b border-[var(--vscode-border)] px-4 py-2">
        <div className="min-w-0 flex-1">
          <div className="text-[13px] text-[#cccccc]">考勤记录</div>
          <div className="truncate text-[11px] text-[#858585]" title={dbPath}>
            {dbPath ? `数据库: ${dbPath}` : "数据库路径加载中…"}
          </div>
        </div>
        <button
          type="button"
          disabled={loading}
          onClick={() => onRefresh()}
          className="shrink-0 rounded border border-[var(--vscode-border)] bg-[#3c3c3c] px-3 py-1.5 text-[12px] text-[#cccccc] hover:bg-[#454545] disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? "同步中…" : "刷新"}
        </button>
      </div>
      {error ? (
        <div className="allow-select px-4 py-3 text-[13px] text-[#f48771]">{error}</div>
      ) : null}
      <div className="allow-select min-h-0 flex-1 overflow-auto">
        <table className="w-max min-w-full border-collapse text-left text-[12px]">
          <thead className="sticky top-0 z-[1] bg-[#252526] text-[#cccccc]">
            <tr>
              {COLUMNS.map((c) => (
                <th
                  key={c.key}
                  className="whitespace-nowrap border-b border-r border-[var(--vscode-border)] px-2 py-1.5 font-normal text-[#858585]"
                >
                  {c.label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="text-[#d4d4d4]">
            {records.length === 0 && !loading ? (
              <tr>
                <td colSpan={COLUMNS.length} className="px-4 py-6 text-[#858585]">
                  暂无记录。点击「刷新」由 Go 无头浏览器登录页取 Cookie 并拉取接口（需本机 Chrome/Chromium，见 README）。也可将已有{" "}
                  <code className="rounded bg-[#2d2d2d] px-1 py-0.5 font-mono text-[11px]">work_hour.db</code>{" "}
                  放到 Preference 配置的路径后仅浏览。
                </td>
              </tr>
            ) : null}
            {records.map((row, idx) => (
              <tr key={`${row.id}-${idx}`} className="hover:bg-[#2a2d2e]">
                {COLUMNS.map((c) => (
                  <td
                    key={c.key}
                    className="max-w-[14rem] truncate border-b border-r border-[var(--vscode-border)] px-2 py-1 font-mono text-[11px]"
                    title={cellValue(row, c.key)}
                  >
                    {cellValue(row, c.key)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
