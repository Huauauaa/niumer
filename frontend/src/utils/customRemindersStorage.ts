import type { CustomReminder } from "../types/reminder";

/** Legacy browser-only storage; migrated once into `reminder.db`. */
export const LEGACY_CUSTOM_REMINDERS_KEY = "niumer.v1.customReminders";

/** After first successful read from SQLite, set so empty DB is never confused with "needs legacy import". */
export const REMINDER_SQLITE_MIGRATION_DONE_KEY =
  "niumer.v1.customReminders.sqliteMigrationDone";

export function hasReminderSqliteMigrationDone(): boolean {
  try {
    return localStorage.getItem(REMINDER_SQLITE_MIGRATION_DONE_KEY) === "1";
  } catch {
    return true;
  }
}

export function markReminderSqliteMigrationDone(): void {
  try {
    localStorage.setItem(REMINDER_SQLITE_MIGRATION_DONE_KEY, "1");
  } catch {
    /* empty */
  }
}

function isValidItem(x: unknown): x is CustomReminder {
  if (!x || typeof x !== "object") return false;
  const o = x as Record<string, unknown>;
  return (
    typeof o.id === "string" &&
    typeof o.name === "string" &&
    typeof o.date === "string" &&
    /^\d{4}-\d{2}-\d{2}$/.test(o.date)
  );
}

export function loadCustomReminders(): CustomReminder[] {
  try {
    const raw = localStorage.getItem(LEGACY_CUSTOM_REMINDERS_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed)) return [];
    return parsed.filter(isValidItem);
  } catch {
    return [];
  }
}

export function clearLegacyCustomReminders(): void {
  try {
    localStorage.removeItem(LEGACY_CUSTOM_REMINDERS_KEY);
  } catch {
    /* empty */
  }
}
