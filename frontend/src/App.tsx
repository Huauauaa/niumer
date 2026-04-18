import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { ActivityId } from "./components/ActivityBar";
import { ActivityBar } from "./components/ActivityBar";
import { BottomPanel } from "./components/BottomPanel";
import type { EditorTab } from "./components/EditorGroup";
import { EditorGroup } from "./components/EditorGroup";
import { MenuBar } from "./components/MenuBar";
import { PreferencesDialog } from "./components/PreferencesDialog";
import { Sidebar } from "./components/Sidebar";
import { StatusBar } from "./components/StatusBar";
import type { BlogDocument } from "./types/blog";
import { nextUntitledFileName } from "./types/blog";
import {
  DeleteBlogFile,
  EnsureWelcomeBlogFile,
  GetBlogWorkDir,
  GetWorkHourDBPath,
  GetWorkHourRecords,
  RefreshWorkHourData,
  ListBlogMarkdownFiles,
  ReadBlogFile,
  RenameBlogFile,
  WriteBlogFile,
} from "../wailsjs/go/main/App";
import type { AttendanceRecord } from "./types/workhour";

const OTHER_TAB_ID = "_home";

const DEV_FALLBACK_WELCOME = `# Welcome

Running without Wails backend: files are not saved to disk. Use \`wails dev\` from the project root.
`;

export default function App() {
  const [activity, setActivity] = useState<ActivityId>("blog");
  const [sidebarWidth, setSidebarWidth] = useState(260);
  const [panelHeight, setPanelHeight] = useState(200);
  const [panelVisible, setPanelVisible] = useState(true);

  const [blogDocs, setBlogDocs] = useState<BlogDocument[]>([]);
  const [blogOpenTabs, setBlogOpenTabs] = useState<string[]>([]);
  const [blogActiveId, setBlogActiveId] = useState("");

  const blogDocsRef = useRef<BlogDocument[]>([]);
  blogDocsRef.current = blogDocs;

  const [otherActiveId, setOtherActiveId] = useState(OTHER_TAB_ID);

  const [prefsOpen, setPrefsOpen] = useState(false);
  const [blogWorkDir, setBlogWorkDir] = useState("");

  const [workHourRecords, setWorkHourRecords] = useState<AttendanceRecord[]>([]);
  const [workHourLoading, setWorkHourLoading] = useState(false);
  const [workHourError, setWorkHourError] = useState<string | null>(null);
  const [workHourDbPath, setWorkHourDbPath] = useState("");

  const dragRef = useRef<"sidebar" | "panel" | null>(null);

  const refreshBlogFromDisk = useCallback(async () => {
    try {
      await EnsureWelcomeBlogFile();
      const files = await ListBlogMarkdownFiles();
      const docs: BlogDocument[] = [];
      for (const f of files) {
        const content = await ReadBlogFile(f);
        docs.push({ fileName: f, title: f, content, dirty: false });
      }
      docs.sort((a, b) => a.fileName.localeCompare(b.fileName));
      setBlogDocs(docs);
      const first = docs[0]?.fileName ?? "";
      setBlogOpenTabs((tabs) => {
        const valid = tabs.filter((t) => docs.some((d) => d.fileName === t));
        if (valid.length > 0) return valid;
        return first ? [first] : [];
      });
      setBlogActiveId((active) =>
        active && docs.some((d) => d.fileName === active) ? active : first,
      );
    } catch {
      setBlogDocs([
        {
          fileName: "Welcome.md",
          title: "Welcome.md",
          content: DEV_FALLBACK_WELCOME,
          dirty: false,
        },
      ]);
      setBlogOpenTabs(["Welcome.md"]);
      setBlogActiveId("Welcome.md");
    }
  }, []);

  /** 仅从本地数据库读取（进入 Workhour 页时调用） */
  const loadWorkHour = useCallback(async () => {
    setWorkHourLoading(true);
    setWorkHourError(null);
    try {
      const [rows, dbPath] = await Promise.all([
        GetWorkHourRecords(),
        GetWorkHourDBPath().catch(() => ""),
      ]);
      setWorkHourRecords(rows as AttendanceRecord[]);
      setWorkHourDbPath(dbPath);
    } catch (e) {
      setWorkHourError(e instanceof Error ? e.message : String(e));
      setWorkHourRecords([]);
    } finally {
      setWorkHourLoading(false);
    }
  }, []);

  /** 爬取 → 入库 → 再查询展示 */
  const refreshWorkHour = useCallback(async () => {
    setWorkHourLoading(true);
    setWorkHourError(null);
    try {
      const rows = await RefreshWorkHourData();
      setWorkHourRecords(rows as AttendanceRecord[]);
      const dbPath = await GetWorkHourDBPath().catch(() => "");
      setWorkHourDbPath(dbPath);
    } catch (e) {
      setWorkHourError(e instanceof Error ? e.message : String(e));
    } finally {
      setWorkHourLoading(false);
    }
  }, []);

  useEffect(() => {
    void GetBlogWorkDir()
      .then(setBlogWorkDir)
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (activity !== "workhour") return;
    void loadWorkHour();
  }, [activity, loadWorkHour]);

  useEffect(() => {
    void refreshBlogFromDisk();
  }, [refreshBlogFromDisk]);

  const onSidebarResizeStart = useCallback(() => {
    dragRef.current = "sidebar";
    document.body.style.cursor = "col-resize";
  }, []);

  const onPanelResizeStart = useCallback(() => {
    dragRef.current = "panel";
    document.body.style.cursor = "row-resize";
  }, []);

  useEffect(() => {
    const onMove = (e: MouseEvent) => {
      if (dragRef.current === "sidebar") {
        const next = Math.min(480, Math.max(180, e.clientX - 48));
        setSidebarWidth(next);
      } else if (dragRef.current === "panel") {
        const h = window.innerHeight - e.clientY;
        setPanelHeight(Math.min(520, Math.max(120, h)));
      }
    };
    const onUp = () => {
      dragRef.current = null;
      document.body.style.cursor = "";
    };
    window.addEventListener("mousemove", onMove);
    window.addEventListener("mouseup", onUp);
    return () => {
      window.removeEventListener("mousemove", onMove);
      window.removeEventListener("mouseup", onUp);
    };
  }, []);

  const openBlogDoc = useCallback((fileName: string) => {
    setBlogOpenTabs((prev) => (prev.includes(fileName) ? prev : [...prev, fileName]));
    setBlogActiveId(fileName);
  }, []);

  const createBlogDoc = useCallback(async () => {
    try {
      const disk = await ListBlogMarkdownFiles();
      const prev = blogDocsRef.current;
      const names = new Set([...prev.map((d) => d.fileName), ...disk]);
      const fileName = nextUntitledFileName([...names]);
      await WriteBlogFile(fileName, "");
      setBlogDocs((p) => {
        const row: BlogDocument = { fileName, title: fileName, content: "", dirty: false };
        const next = [...p.filter((d) => d.fileName !== fileName), row];
        return next.sort((a, b) => a.fileName.localeCompare(b.fileName));
      });
      setBlogOpenTabs((t) => (t.includes(fileName) ? t : [...t, fileName]));
      setBlogActiveId(fileName);
    } catch (e) {
      window.alert(e instanceof Error ? e.message : String(e));
    }
  }, []);

  const renameBlogDoc = useCallback(async (oldName: string, newName: string) => {
    if (!newName || newName === oldName) return;
    const exists = blogDocsRef.current.some(
      (d) => d.fileName !== oldName && d.fileName.toLowerCase() === newName.toLowerCase(),
    );
    if (exists) {
      window.alert("A file with that name already exists.");
      return;
    }
    try {
      await RenameBlogFile(oldName, newName);
    } catch (e) {
      window.alert(e instanceof Error ? e.message : String(e));
      return;
    }
    setBlogDocs((prev) => {
      const next = prev.map((d) =>
        d.fileName === oldName ? { ...d, fileName: newName, title: newName, dirty: false } : d,
      );
      return next.sort((a, b) => a.fileName.localeCompare(b.fileName));
    });
    setBlogOpenTabs((tabs) => tabs.map((t) => (t === oldName ? newName : t)));
    setBlogActiveId((active) => (active === oldName ? newName : active));
  }, []);

  const deleteBlogDoc = useCallback(
    async (fileName: string) => {
      const doc = blogDocsRef.current.find((d) => d.fileName === fileName);
      if (!doc) return;
      if (!window.confirm(`Delete "${doc.title}"?`)) return;
      try {
        await DeleteBlogFile(fileName);
      } catch (e) {
        window.alert(e instanceof Error ? e.message : String(e));
        return;
      }
      setBlogDocs((prev) => prev.filter((d) => d.fileName !== fileName));
      setBlogOpenTabs((prev) => {
        const next = prev.filter((t) => t !== fileName);
        setBlogActiveId((active) => {
          if (active !== fileName) return active;
          const idx = prev.indexOf(fileName);
          return next[Math.max(0, idx - 1)] ?? next[0] ?? "";
        });
        return next;
      });
    },
    [],
  );

  const saveActiveBlog = useCallback(async () => {
    const id = blogActiveId;
    if (!id) return;
    const doc = blogDocsRef.current.find((d) => d.fileName === id);
    if (!doc) return;
    try {
      await WriteBlogFile(doc.fileName, doc.content);
      setBlogDocs((prev) =>
        prev.map((d) => (d.fileName === doc.fileName ? { ...d, dirty: false } : d)),
      );
    } catch (e) {
      window.alert(e instanceof Error ? e.message : String(e));
    }
  }, [blogActiveId]);

  const updateBlogContent = useCallback(
    (value: string) => {
      if (!blogActiveId) return;
      setBlogDocs((prev) =>
        prev.map((d) =>
          d.fileName === blogActiveId ? { ...d, content: value, dirty: true } : d,
        ),
      );
    },
    [blogActiveId],
  );

  useEffect(() => {
    if (activity !== "blog") return;
    if (blogOpenTabs.length === 0) {
      if (blogActiveId !== "") setBlogActiveId("");
      return;
    }
    if (blogActiveId === "" || !blogOpenTabs.includes(blogActiveId)) {
      setBlogActiveId(blogOpenTabs[0] ?? "");
    }
  }, [activity, blogOpenTabs, blogActiveId]);

  const closeBlogTab = useCallback((fileName: string) => {
    setBlogOpenTabs((prev) => {
      const next = prev.filter((t) => t !== fileName);
      setBlogActiveId((active) => {
        if (active !== fileName) return active;
        const idx = prev.indexOf(fileName);
        return next[Math.max(0, idx - 1)] ?? next[0] ?? "";
      });
      return next;
    });
  }, []);

  const blogTabs: EditorTab[] = useMemo(() => {
    const out: EditorTab[] = [];
    for (const fileName of blogOpenTabs) {
      const d = blogDocs.find((x) => x.fileName === fileName);
      if (d) out.push({ id: d.fileName, title: d.title, dirty: d.dirty });
    }
    return out;
  }, [blogOpenTabs, blogDocs]);

  const activeBlogDoc = blogDocs.find((d) => d.fileName === blogActiveId);

  const handleEditorClose = (id: string) => {
    if (activity === "blog") {
      closeBlogTab(id);
      return;
    }
    if (id === OTHER_TAB_ID) return;
  };

  const editorTabs: EditorTab[] =
    activity === "blog"
      ? blogTabs
      : activity === "workhour"
        ? [{ id: OTHER_TAB_ID, title: "Workhour", dirty: false }]
        : [{ id: OTHER_TAB_ID, title: "Home", dirty: false }];

  const editorActiveId = activity === "blog" ? blogActiveId : otherActiveId;

  const setEditorActiveId = (id: string) => {
    if (activity === "blog") setBlogActiveId(id);
    else setOtherActiveId(id);
  };

  return (
    <div className="flex h-full flex-col bg-[#1e1e1e]">
      <MenuBar onOpenPreference={() => setPrefsOpen(true)} />
      <PreferencesDialog
        open={prefsOpen}
        onClose={() => setPrefsOpen(false)}
        onSaved={() => {
          void GetBlogWorkDir()
            .then(setBlogWorkDir)
            .catch(() => {});
          void GetWorkHourDBPath()
            .then(setWorkHourDbPath)
            .catch(() => {});
          void refreshBlogFromDisk();
          void loadWorkHour();
        }}
      />
      <div className="flex min-h-0 flex-1 flex-col">
        <div className="flex min-h-0 flex-1">
          <ActivityBar active={activity} onChange={setActivity} />
          <Sidebar
            activity={activity}
            width={sidebarWidth}
            onResizeStart={onSidebarResizeStart}
            blogDocuments={blogDocs}
            blogSelectedId={blogActiveId}
            onBlogOpen={openBlogDoc}
            onBlogNew={createBlogDoc}
            onBlogDelete={deleteBlogDoc}
            onBlogRename={renameBlogDoc}
            workHourRecordCount={activity === "workhour" ? workHourRecords.length : undefined}
            workHourDbPath={activity === "workhour" ? workHourDbPath : undefined}
          />
          <div className="flex min-w-0 flex-1 flex-col">
            <EditorGroup
              tabs={editorTabs}
              activeId={editorActiveId}
              onSelect={setEditorActiveId}
              onClose={handleEditorClose}
              blogEditor={activity === "blog"}
              editorContent={activity === "blog" ? (activeBlogDoc?.content ?? "") : ""}
              onEditorContentChange={updateBlogContent}
              onSaveBlog={activity === "blog" ? saveActiveBlog : undefined}
              breadcrumbLabel={activeBlogDoc?.title}
              workHourView={activity === "workhour"}
              workHourRecords={workHourRecords}
              workHourLoading={workHourLoading}
              workHourError={workHourError}
              workHourDbPath={workHourDbPath}
              onRefreshWorkHour={refreshWorkHour}
            />
            <BottomPanel
              height={panelHeight}
              visible={panelVisible}
              onResizeStart={onPanelResizeStart}
              onToggle={() => setPanelVisible((v) => !v)}
            />
          </div>
        </div>
        <StatusBar blogWorkDir={blogWorkDir} />
      </div>
    </div>
  );
}
