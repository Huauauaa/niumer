export type UITheme = "dark" | "light";

const STORAGE_KEY = "niumer-ui-theme";

export function getStoredTheme(): UITheme {
  try {
    const v = localStorage.getItem(STORAGE_KEY);
    if (v === "light") return "light";
  } catch {
    /* ignore */
  }
  return "dark";
}

export function applyTheme(theme: UITheme): void {
  document.documentElement.dataset.theme = theme;
  try {
    localStorage.setItem(STORAGE_KEY, theme);
  } catch {
    /* ignore */
  }
}

export function initThemeFromStorage(): void {
  applyTheme(getStoredTheme());
}
