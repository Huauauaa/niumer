/** One public holiday period (inclusive dates, local calendar). */
export type HolidayPeriod = {
  /** Short label shown in UI */
  name: string;
  /** YYYY-MM-DD */
  start: string;
  end: string;
};

/**
 * Offline fallback when holiday-cn (jsDelivr) cannot be loaded.
 * Prefer loading from holiday-cn at runtime (`fetchHolidayDataAround` in `holidayCnApi`).
 */
export const FALLBACK_CHINA_HOLIDAY_PERIODS: HolidayPeriod[] = [
  { name: "元旦", start: "2026-01-01", end: "2026-01-03" },
  { name: "春节", start: "2026-02-15", end: "2026-02-23" },
  { name: "清明节", start: "2026-04-04", end: "2026-04-06" },
  { name: "劳动节", start: "2026-05-01", end: "2026-05-05" },
  { name: "端午节", start: "2026-06-19", end: "2026-06-21" },
  { name: "中秋节", start: "2026-09-25", end: "2026-09-27" },
  { name: "国庆节", start: "2026-10-01", end: "2026-10-07" },
  { name: "元旦", start: "2027-01-01", end: "2027-01-03" },
  { name: "春节", start: "2027-02-05", end: "2027-02-11" },
  { name: "清明节", start: "2027-04-04", end: "2027-04-06" },
  { name: "劳动节", start: "2027-05-01", end: "2027-05-02" },
  { name: "端午节", start: "2027-06-07", end: "2027-06-09" },
  { name: "中秋节", start: "2027-09-15", end: "2027-09-17" },
  { name: "国庆节", start: "2027-10-01", end: "2027-10-07" },
];

export function parseYMD(s: string): Date {
  const [y, m, d] = s.split("-").map(Number);
  return new Date(y, m - 1, d);
}

/** Local calendar date at midnight (no time-of-day). */
export function startOfLocalDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate());
}

export function diffCalendarDays(from: Date, to: Date): number {
  const a = startOfLocalDay(from).getTime();
  const b = startOfLocalDay(to).getTime();
  return Math.round((b - a) / 86400000);
}

export function formatTodayChinese(d: Date = new Date()): string {
  const w = ["日", "一", "二", "三", "四", "五", "六"][d.getDay()];
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日 星期${w}`;
}

export function formatRangeCN(p: HolidayPeriod): string {
  const a = parseYMD(p.start);
  const b = parseYMD(p.end);
  const f = (x: Date) => `${x.getMonth() + 1}月${x.getDate()}日`;
  if (p.start === p.end) return f(a);
  return `${f(a)} — ${f(b)}`;
}

export type HolidayCountdown =
  | {
      kind: "upcoming";
      todayLabel: string;
      period: HolidayPeriod;
      daysUntil: number;
    }
  | {
      kind: "during";
      todayLabel: string;
      period: HolidayPeriod;
      next: { period: HolidayPeriod; daysUntil: number };
    }
  | { kind: "none"; todayLabel: string; message: string };

/** Next holiday context relative to `now` (local). */
export function getHolidayCountdown(
  periods: HolidayPeriod[],
  now: Date = new Date(),
): HolidayCountdown {
  const today = startOfLocalDay(now);
  const todayLabel = formatTodayChinese(now);
  const sorted = [...periods].sort(
    (a, b) => parseYMD(a.start).getTime() - parseYMD(b.start).getTime(),
  );

  for (let i = 0; i < sorted.length; i++) {
    const p = sorted[i]!;
    const ps = startOfLocalDay(parseYMD(p.start));
    const pe = startOfLocalDay(parseYMD(p.end));
    if (today < ps) {
      return {
        kind: "upcoming",
        todayLabel,
        period: p,
        daysUntil: diffCalendarDays(today, ps),
      };
    }
    if (today >= ps && today <= pe) {
      const next = sorted
        .slice(i + 1)
        .find((q) => startOfLocalDay(parseYMD(q.start)) > pe);
      if (!next) {
        return {
          kind: "none",
          todayLabel,
          message: "当前假期之后暂无更多放假安排数据。",
        };
      }
      const ns = startOfLocalDay(parseYMD(next.start));
      return {
        kind: "during",
        todayLabel,
        period: p,
        next: {
          period: next,
          daysUntil: diffCalendarDays(today, ns),
        },
      };
    }
  }
  return {
    kind: "none",
    todayLabel,
    message: "当前日期之后暂无放假安排数据，请检查网络或更新数据源。",
  };
}

/** Upcoming periods whose start is on or after `today` (for table). */
export function upcomingPeriods(
  periods: HolidayPeriod[],
  now: Date = new Date(),
  limit = 8,
): { period: HolidayPeriod; daysUntil: number }[] {
  const today = startOfLocalDay(now);
  const sorted = [...periods].sort(
    (a, b) => parseYMD(a.start).getTime() - parseYMD(b.start).getTime(),
  );
  const out: { period: HolidayPeriod; daysUntil: number }[] = [];
  for (const p of sorted) {
    const ps = startOfLocalDay(parseYMD(p.start));
    if (ps < today) continue;
    out.push({ period: p, daysUntil: diffCalendarDays(today, ps) });
    if (out.length >= limit) break;
  }
  return out;
}
