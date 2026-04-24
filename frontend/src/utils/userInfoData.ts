/** From `/user-info` `data` JSON (same shape as SQLite `userInfoJson`). */
export function parseUserInfoDataSummary(userInfoJson: string | undefined) {
  let department = "—";
  let supplier = "—";
  if (!userInfoJson?.trim()) {
    return { department, supplier };
  }
  try {
    const o = JSON.parse(userInfoJson) as {
      departmentDTO?: {
        departmentChineseName?: string;
        departmentEnglishName?: string;
      };
      supplier?: string;
    };
    if (o.departmentDTO) {
      const zh = (o.departmentDTO.departmentChineseName || "").trim();
      const en = (o.departmentDTO.departmentEnglishName || "").trim();
      department = zh || en || "—";
    }
    const s = o.supplier != null ? String(o.supplier).trim() : "";
    supplier = s || "—";
  } catch {
    /* keep — */
  }
  return { department, supplier };
}
