import type { HolidayPeriod } from "./chinaHolidays";
import { diffCalendarDays, parseYMD, startOfLocalDay } from "./chinaHolidays";
import type { HolidayCnDayRow } from "./holidayCnApi";

function toYmdLocal(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function ymdForCalendarDay(
  anchorYear: number,
  monthIndex0: number,
  day: number,
): string {
  const d = new Date(anchorYear, monthIndex0, day);
  return toYmdLocal(d);
}

/** Each date in inclusive [start,end] of periods is treated as non-workday. */
function collectOffDaysFromPeriods(periods: HolidayPeriod[]): Set<string> {
  const off = new Set<string>();
  for (const p of periods) {
    const end = startOfLocalDay(parseYMD(p.end));
    for (
      let d = startOfLocalDay(parseYMD(p.start));
      d <= end;
      d.setDate(d.getDate() + 1)
    ) {
      off.add(toYmdLocal(d));
    }
  }
  return off;
}

/**
 * Workday resolver: prefers holiday-cn rows (`!isOffDay`); unknown dates fall
 * back to Mon–Fri. When `days` is empty, uses weekend + statutory off ranges
 * from `fallbackPeriods`.
 */
export function makeIsWorkday(
  days: HolidayCnDayRow[],
  fallbackPeriods: HolidayPeriod[],
): (ymd: string) => boolean {
  if (days.length > 0) {
    const map = new Map<string, boolean>();
    for (const row of days) {
      map.set(row.date, !row.isOffDay);
    }
    return (ymd: string) => {
      const v = map.get(ymd);
      if (v !== undefined) return v;
      const d = parseYMD(ymd);
      const w = d.getDay();
      return w !== 0 && w !== 6;
    };
  }
  const off = collectOffDaysFromPeriods(fallbackPeriods);
  return (ymd: string) => {
    if (off.has(ymd)) return false;
    const d = parseYMD(ymd);
    const w = d.getDay();
    return w !== 0 && w !== 6;
  };
}

export type PayDateResolution = {
  payYmd: string;
  nominalYmd: string;
  advanced: boolean;
};

/** Last workday on or before the 15th of `(anchorYear, monthIndex0)`; walks back across month if needed. */
export function resolvePayDateOnOrBeforeFifteenth(
  anchorYear: number,
  monthIndex0: number,
  isWorkday: (ymd: string) => boolean,
): PayDateResolution {
  const nominalYmd = ymdForCalendarDay(anchorYear, monthIndex0, 15);
  const d = parseYMD(nominalYmd);
  let guard = 0;
  while (!isWorkday(toYmdLocal(d)) && guard < 62) {
    d.setDate(d.getDate() - 1);
    guard++;
  }
  const payYmd = toYmdLocal(d);
  return {
    payYmd,
    nominalYmd,
    advanced: payYmd !== nominalYmd,
  };
}

export type NextPaydayInfo = {
  payYmd: string;
  anchorYear: number;
  anchorMonth: number;
  advanced: boolean;
  daysUntil: number;
};

export function getNextPaydayInfo(
  now: Date,
  isWorkday: (ymd: string) => boolean,
): NextPaydayInfo {
  const today = startOfLocalDay(now);
  let y = today.getFullYear();
  let m0 = today.getMonth();

  for (let i = 0; i < 36; i++) {
    const { payYmd, advanced } = resolvePayDateOnOrBeforeFifteenth(
      y,
      m0,
      isWorkday,
    );
    const pay = startOfLocalDay(parseYMD(payYmd));
    if (pay.getTime() >= today.getTime()) {
      return {
        payYmd,
        anchorYear: y,
        anchorMonth: m0 + 1,
        advanced,
        daysUntil: diffCalendarDays(today, pay),
      };
    }
    m0++;
    if (m0 > 11) {
      m0 = 0;
      y++;
    }
  }

  throw new Error("getNextPaydayInfo: no pay date within 36 months");
}
