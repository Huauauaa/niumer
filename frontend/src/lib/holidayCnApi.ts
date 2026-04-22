import type { HolidayPeriod } from "./chinaHolidays";

/** NateScarlet/holiday-cn via jsDelivr (static JSON, no Cloudflare browser challenge). */
export const HOLIDAY_CN_CDN_BASE =
  "https://cdn.jsdelivr.net/gh/NateScarlet/holiday-cn@master";

export type HolidayCnDayRow = {
  name: string;
  date: string;
  isOffDay: boolean;
};

type HolidayCnDoc = {
  year: number;
  days: HolidayCnDayRow[];
};

function toYMD(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function nextCalendarDate(ymd: string): string {
  const [y, mo, da] = ymd.split("-").map(Number);
  const d = new Date(y, mo - 1, da);
  d.setDate(d.getDate() + 1);
  return toYMD(d);
}

export async function fetchHolidayCnYear(
  year: number,
): Promise<HolidayCnDayRow[]> {
  const url = `${HOLIDAY_CN_CDN_BASE}/${year}.json`;
  const res = await fetch(url, { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`获取 ${year} 年节假日数据失败（HTTP ${res.status}）`);
  }
  const doc = (await res.json()) as HolidayCnDoc;
  if (!Array.isArray(doc.days)) {
    throw new Error(`${year} 年数据缺少 days 字段`);
  }
  return doc.days;
}

/** Merge consecutive `isOffDay` rows with the same `name` into inclusive [start, end] periods. */
export function daysToHolidayPeriods(days: HolidayCnDayRow[]): HolidayPeriod[] {
  const sorted = [...days].sort((a, b) => a.date.localeCompare(b.date));
  const periods: HolidayPeriod[] = [];
  let cur: { name: string; start: string; end: string } | null = null;

  for (const row of sorted) {
    if (!row.isOffDay) {
      if (cur) {
        periods.push({ name: cur.name, start: cur.start, end: cur.end });
        cur = null;
      }
      continue;
    }
    if (!cur) {
      cur = { name: row.name, start: row.date, end: row.date };
      continue;
    }
    const contiguous =
      nextCalendarDate(cur.end) === row.date && row.name === cur.name;
    if (contiguous) {
      cur.end = row.date;
    } else {
      periods.push({ name: cur.name, start: cur.start, end: cur.end });
      cur = { name: row.name, start: row.date, end: row.date };
    }
  }
  if (cur) {
    periods.push({ name: cur.name, start: cur.start, end: cur.end });
  }
  return periods;
}

export type HolidayDataBundle = {
  periods: HolidayPeriod[];
  days: HolidayCnDayRow[];
};

/**
 * Load holiday-cn for `year` and `year + 1` (covers year-end into next January).
 * If the next calendar year file is missing, only the primary year is used.
 */
export async function fetchHolidayDataAround(
  ref: Date = new Date(),
): Promise<HolidayDataBundle> {
  const y = ref.getFullYear();
  const days: HolidayCnDayRow[] = [];
  days.push(...(await fetchHolidayCnYear(y)));
  try {
    days.push(...(await fetchHolidayCnYear(y + 1)));
  } catch {
    // e.g. remote year not published yet
  }
  const periods = daysToHolidayPeriods(days).sort(
    (a, b) => a.start.localeCompare(b.start) || a.end.localeCompare(b.end),
  );
  return { periods, days };
}
