import { useCallback, useEffect, useMemo, useState } from "react";
import type { CustomReminder } from "../types/reminder";
import {
  diffCalendarDays,
  FALLBACK_CHINA_HOLIDAY_PERIODS,
  formatRangeCN,
  formatTodayChinese,
  getHolidayCountdown,
  parseYMD,
  startOfLocalDay,
  upcomingPeriods,
} from "../lib/chinaHolidays";
import {
  fetchHolidayDataAround,
  type HolidayCnDayRow,
} from "../lib/holidayCnApi";
import { getNextPaydayInfo, makeIsWorkday } from "../lib/payday";

const YMD_RE = /^\d{4}-\d{2}-\d{2}$/;

const cx = {
  sectionHead:
    "mb-2 text-[11px] uppercase tracking-wide text-[#858585]",
  refreshBtn:
    "shrink-0 rounded border border-[var(--vscode-border)] bg-[#3c3c3c] px-3 py-1.5 text-[12px] text-[#cccccc] hover:bg-[#454545] disabled:cursor-not-allowed disabled:opacity-50",
  panel: "mb-5 rounded border border-[var(--vscode-border)] bg-[#252526] px-3 py-3",
  body: "allow-select min-h-0 flex-1 overflow-auto px-4 py-4 text-[#cccccc]",
  prose: "text-[15px] leading-relaxed text-[#e1e4e8]",
  muted: "text-[12px] leading-relaxed text-[#858585]",
  accentNum: "tabular-nums text-[#e37933]",
} as const;

type Props = {
  /** 左侧「我的提醒」中的全部条目，主区一并展示倒计时 */
  customReminders?: CustomReminder[];
};

type PersonalRow =
  | {
      key: string;
      kind: "future";
      name: string;
      date: string;
      days: number;
    }
  | { key: string; kind: "today"; name: string; date: string }
  | {
      key: string;
      kind: "past";
      name: string;
      date: string;
      daysPast: number;
    };

function buildPersonalRows(items: CustomReminder[]): PersonalRow[] {
  const today = startOfLocalDay(new Date());
  const rows: PersonalRow[] = [];
  for (const r of items) {
    if (!r.date || !YMD_RE.test(r.date)) continue;
    const key = r.id.trim() || `${r.date}::${r.name}`;
    const target = startOfLocalDay(parseYMD(r.date));
    const n = diffCalendarDays(today, target);
    if (n > 0) {
      rows.push({
        key,
        kind: "future",
        name: r.name,
        date: r.date,
        days: n,
      });
    } else if (n === 0) {
      rows.push({ key, kind: "today", name: r.name, date: r.date });
    } else {
      rows.push({
        key,
        kind: "past",
        name: r.name,
        date: r.date,
        daysPast: -n,
      });
    }
  }
  rows.sort((a, b) => a.date.localeCompare(b.date));
  return rows;
}

function PersonalCountdownLine({ row }: { row: PersonalRow }) {
  if (row.kind === "future") {
    return (
      <p className={cx.prose}>
        距离「{row.name}」还有 <span className={cx.accentNum}>{row.days}</span>{" "}
        天。
        <span className="mt-1 block font-mono text-[11px] font-normal text-[#858585]">
          目标日 {row.date}
        </span>
      </p>
    );
  }
  if (row.kind === "today") {
    return (
      <p className={cx.prose}>
        「{row.name}」就是今天（{row.date}）。
      </p>
    );
  }
  return (
    <p className={cx.prose}>
      「{row.name}」的日期（{row.date}）已过，已过去{" "}
      <span className={cx.accentNum}>{row.daysPast}</span> 天。
    </p>
  );
}

export function HolidayReminderView({ customReminders = [] }: Props) {
  const [periods, setPeriods] = useState(FALLBACK_CHINA_HOLIDAY_PERIODS);
  const [calendarDays, setCalendarDays] = useState<HolidayCnDayRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [source, setSource] = useState<"api" | "fallback">("fallback");

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const bundle = await fetchHolidayDataAround(new Date());
      setPeriods(bundle.periods);
      setCalendarDays(bundle.days);
      setSource("api");
    } catch (e) {
      setPeriods(FALLBACK_CHINA_HOLIDAY_PERIODS);
      setCalendarDays([]);
      setSource("fallback");
      setError(
        e instanceof Error
          ? e.message
          : "无法从网络加载节假日数据，已使用内置备用表。",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const countdown = useMemo(
    () => getHolidayCountdown(periods, new Date()),
    [periods],
  );
  const table = useMemo(
    () => upcomingPeriods(periods, new Date(), 10),
    [periods],
  );

  const payday = useMemo(() => {
    const isWorkday = makeIsWorkday(calendarDays, periods);
    return getNextPaydayInfo(new Date(), isWorkday);
  }, [calendarDays, periods]);

  const headline = useMemo(() => {
    if (countdown.kind === "upcoming") {
      const d = countdown.daysUntil;
      return `今天是${countdown.todayLabel}，距离「${countdown.period.name}」假期还有 ${d} 天。`;
    }
    if (countdown.kind === "during") {
      return `今天是${countdown.todayLabel}，您正处于「${countdown.period.name}」假期中。距离「${countdown.next.period.name}」还有 ${countdown.next.daysUntil} 天。`;
    }
    return `今天是${countdown.todayLabel}。${countdown.message}`;
  }, [countdown]);

  const personalRows = useMemo(
    () => buildPersonalRows(customReminders),
    [customReminders],
  );

  const sourceCaption =
    source === "api"
      ? "数据：holiday-cn（jsDelivr CDN）"
      : "数据：内置备用表（联网后将自动尝试拉取最新）";

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {error ? (
        <div className="allow-select shrink-0 px-4 py-2 text-[12px] text-[#f48771]">
          {error}
        </div>
      ) : null}
      <div className={cx.body}>
        <div className="mb-4 flex flex-wrap items-center justify-between gap-2">
          <p className="min-w-0 flex-1 text-[11px] text-[#858585]">
            {sourceCaption}
          </p>
          <button
            type="button"
            disabled={loading}
            onClick={() => void load()}
            className={cx.refreshBtn}
          >
            {loading ? "加载中…" : "刷新"}
          </button>
        </div>

        <div className={cx.panel}>
          <div className="mb-1.5 text-[11px] font-bold uppercase tracking-wide text-[#858585]">
            我的倒计时
          </div>
          {personalRows.length === 0 ? (
            <p className={cx.muted}>
              在左侧「我的提醒」中添加条目，此处会列出全部条目的剩余或已过天数。
            </p>
          ) : (
            <ul className="flex flex-col gap-3">
              {personalRows.map((row) => (
                <li
                  key={row.key}
                  className="border-b border-[var(--vscode-border)] pb-3 last:border-b-0 last:pb-0"
                >
                  <PersonalCountdownLine row={row} />
                </li>
              ))}
            </ul>
          )}
        </div>

        <p className={`mb-6 ${cx.prose}`}>{headline}</p>

        <div className={`border-t border-[var(--vscode-border)] pt-5 ${cx.sectionHead}`}>
          发薪日
        </div>
        <p className={`mb-6 ${cx.prose}`}>
          下次发薪日：
          {formatTodayChinese(parseYMD(payday.payYmd))}
          ，距离还有 <span className={cx.accentNum}>{payday.daysUntil}</span>{" "}
          天。
        </p>

        <div className={cx.sectionHead}>后续假期（参考）</div>
        <div className="min-h-0 overflow-auto rounded border border-[var(--vscode-border)]">
          <table className="w-max min-w-full border-collapse text-left text-[12px]">
            <thead className="sticky top-0 z-[1] bg-[#252526] text-[#cccccc]">
              <tr>
                <th className="whitespace-nowrap border-b border-r border-[var(--vscode-border)] px-2 py-1.5 font-normal text-[#858585]">
                  节日
                </th>
                <th className="whitespace-nowrap border-b border-r border-[var(--vscode-border)] px-2 py-1.5 font-normal text-[#858585]">
                  放假区间
                </th>
                <th className="whitespace-nowrap border-b border-[var(--vscode-border)] px-2 py-1.5 font-normal text-[#858585]">
                  距开始（天）
                </th>
              </tr>
            </thead>
            <tbody className="text-[#d4d4d4]">
              {table.length === 0 ? (
                <tr>
                  <td colSpan={3} className="px-4 py-6 text-[#858585]">
                    暂无后续条目。
                  </td>
                </tr>
              ) : (
                table.map(({ period, daysUntil }) => (
                  <tr
                    key={`${period.name}-${period.start}`}
                    className="hover:bg-[#2a2d2e]"
                  >
                    <td className="border-b border-r border-[var(--vscode-border)] px-2 py-1.5 font-mono text-[11px]">
                      {period.name}
                    </td>
                    <td className="border-b border-r border-[var(--vscode-border)] px-2 py-1.5 font-mono text-[11px]">
                      {formatRangeCN(period)}
                    </td>
                    <td className="border-b border-[var(--vscode-border)] px-2 py-1.5 font-mono text-[11px] tabular-nums text-[#e37933]">
                      {daysUntil}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        <p className="mt-4 text-[11px] leading-relaxed text-[#858585]">
          在线数据来自开源项目{" "}
          <a
            className="text-[#3794ff] hover:underline"
            href="https://github.com/NateScarlet/holiday-cn"
            target="_blank"
            rel="noreferrer"
          >
            NateScarlet/holiday-cn
          </a>
          ，依据国务院通知维护；与官方文件不一致时以政府公布为准。
        </p>
      </div>
    </div>
  );
}
