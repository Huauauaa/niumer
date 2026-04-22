import type { CustomReminder } from "../types/reminder";

function toPlainKeys(raw: unknown): Record<string, unknown> {
  if (raw == null || typeof raw !== "object") {
    return {};
  }
  try {
    const s = JSON.stringify(raw);
    if (s && s !== "{}" && s !== "null") {
      return JSON.parse(s) as Record<string, unknown>;
    }
  } catch {
    /* fall through */
  }
  const out: Record<string, unknown> = {};
  const o = raw as Record<string, unknown>;
  for (const k of Object.keys(o)) {
    out[k] = o[k];
  }
  return out;
}

/**
 * Wails may return class instances / proxies; normalize keys so DELETE/update
 * always see the same id/name/date as stored.
 */
export function normalizeCustomReminderRow(raw: unknown): CustomReminder {
  const o = toPlainKeys(raw);
  const pick = (...keys: string[]): string => {
    for (const k of keys) {
      const v = o[k];
      if (v != null && String(v).trim() !== "") return String(v).trim();
    }
    for (const k of Object.keys(o)) {
      if (/^id$/i.test(k)) {
        const v = o[k];
        if (v != null && String(v).trim() !== "") return String(v).trim();
      }
    }
    return "";
  };
  const pickName = (...keys: string[]): string => {
    for (const k of keys) {
      const v = o[k];
      if (v != null && String(v).trim() !== "") return String(v).trim();
    }
    for (const k of Object.keys(o)) {
      if (/^name$/i.test(k)) {
        const v = o[k];
        if (v != null && String(v).trim() !== "") return String(v).trim();
      }
    }
    return "";
  };
  const pickDate = (...keys: string[]): string => {
    for (const k of keys) {
      const v = o[k];
      if (v != null && String(v).trim() !== "") return String(v).trim();
    }
    for (const k of Object.keys(o)) {
      if (/^date$/i.test(k)) {
        const v = o[k];
        if (v != null && String(v).trim() !== "") return String(v).trim();
      }
    }
    return "";
  };
  return {
    id: pick("id", "ID", "Id"),
    name: pickName("name", "Name"),
    date: pickDate("date", "Date"),
  };
}
