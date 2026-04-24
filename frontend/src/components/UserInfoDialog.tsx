import { useCallback, useEffect, useState } from "react";
import { GetWorkHourUserProfile } from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";
import { parseUserInfoDataSummary } from "../utils/userInfoData";

type Props = {
  open: boolean;
  onClose: () => void;
};

export function UserInfoDialog({ open, onClose }: Props) {
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const [profile, setProfile] = useState<main.WorkHourUserProfileView | null>(
    null,
  );

  const load = useCallback(async () => {
    setLoading(true);
    setErr(null);
    try {
      setProfile(await GetWorkHourUserProfile());
    } catch (e) {
      setProfile(null);
      setErr(e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!open) return;
    void load();
  }, [open, load]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onClose]);

  if (!open) return null;

  const hasUserCache = Boolean(
    profile &&
      ((profile.hrId ?? 0) > 0 ||
        (profile.userAccount || "").trim() !== ""),
  );
  const { department, supplier } = profile
    ? parseUserInfoDataSummary(profile.userInfoJson)
    : { department: "—", supplier: "—" };

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50"
      role="presentation"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        className="max-h-[min(100%-2rem,90vh)] w-[min(100%-2rem,480px)] overflow-y-auto rounded border border-[var(--vscode-border)] bg-[#252526] p-4 shadow-xl"
        role="dialog"
        aria-labelledby="userinfo-title"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <h2
          id="userinfo-title"
          className="m-0 mb-3 text-[13px] font-semibold text-[#cccccc]"
        >
          User info
        </h2>
        {loading && (
          <p className="mb-3 text-[12px] text-[#858585]">正在加载…</p>
        )}
        {err && (
          <p className="mb-3 text-[12px] text-[#f48771]">{err}</p>
        )}
        {hasUserCache && !loading && profile && (
          <dl className="mb-3 grid max-w-lg grid-cols-[7.5rem_1fr] gap-x-2 gap-y-1.5 text-[12px] text-[#cccccc]">
            <dt className="text-[#858585]">Account</dt>
            <dd className="m-0 font-mono">{profile.userAccount || "—"}</dd>
            <dt className="text-[#858585]">Department</dt>
            <dd className="m-0 break-words text-[#cccccc]">{department}</dd>
            <dt className="text-[#858585]">Supplier</dt>
            <dd className="m-0 break-words text-[#cccccc]">{supplier}</dd>
          </dl>
        )}
        {!loading && !hasUserCache && !err && (
          <p className="mb-3 text-[12px] text-[#858585]">暂无数据</p>
        )}
        <div className="mt-4 flex justify-end">
          <button
            type="button"
            className="rounded px-3 py-1.5 text-[12px] text-[#cccccc] hover:bg-white/10"
            onClick={onClose}
          >
            关闭
          </button>
        </div>
      </div>
    </div>
  );
}
