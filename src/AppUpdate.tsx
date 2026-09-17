import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";
import { Download, RefreshCw, X } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";

type CheckResult = {
  status: "unsupported" | "upToDate" | "noRelease" | "available" | "error";
  currentVersion: string;
  version: string | null;
  body: string | null;
  reason: string | null;
};
type UpdateEvent =
  | { event: "started"; data: { contentLength: number | null } }
  | { event: "progress"; data: { chunkLength: number } }
  | { event: "finished" };

export function AppUpdate({
  currentVersion,
  rimeRunning,
  onInstallingChange,
}: {
  currentVersion: string;
  rimeRunning: boolean;
  onInstallingChange: (value: boolean) => void;
}) {
  const [includePrerelease, setIncludePrerelease] = useState(() => {
    const saved = localStorage.getItem("omr-update-prerelease");
    return saved === null ? currentVersion.includes("-") : saved === "true";
  });
  const [result, setResult] = useState<CheckResult | null>(null);
  const [checking, setChecking] = useState(false);
  const [installing, setInstalling] = useState(false);
  const [error, setError] = useState("");
  const [progress, setProgress] = useState({
    downloaded: 0,
    total: 0,
    finished: false,
  });
  const [open, setOpen] = useState(false);
  const dialog = useRef<HTMLDialogElement>(null);
  const busy = useRef(false);

  const checkUpdate = useCallback(
    async (silent = false) => {
      if (busy.current) return;
      busy.current = true;
      setChecking(true);
      setError("");
      setResult(null);
      if (!silent) setOpen(true);
      try {
        const next = await invoke<CheckResult>("check_app_update", {
          includePrerelease,
        });
        setResult(next);
        if (next.status === "error")
          setError(next.reason ?? "检查更新失败，请稍后重试");
      } catch (err) {
        setError(String(err));
      } finally {
        busy.current = false;
        setChecking(false);
      }
    },
    [includePrerelease],
  );

  useEffect(() => {
    localStorage.setItem("omr-update-prerelease", String(includePrerelease));
    const timer = window.setTimeout(() => void checkUpdate(true), 1500);
    return () => window.clearTimeout(timer);
  }, [includePrerelease, checkUpdate]);

  useEffect(() => {
    let disposed = false;
    let stop: (() => void) | undefined;
    void listen<UpdateEvent>("app-update-event", ({ payload }) => {
      if (payload.event === "started") {
        setProgress({
          downloaded: 0,
          total: payload.data.contentLength ?? 0,
          finished: false,
        });
      } else if (payload.event === "progress") {
        setProgress((value) => ({
          ...value,
          downloaded: value.downloaded + payload.data.chunkLength,
        }));
      } else {
        setProgress((value) => ({ ...value, finished: true }));
      }
    })
      .then((unlisten) => {
        if (disposed) unlisten();
        else stop = unlisten;
      })
      .catch((err) => setError(`无法读取更新进度：${String(err)}`));
    return () => {
      disposed = true;
      stop?.();
    };
  }, []);

  useEffect(() => {
    if (open) dialog.current?.showModal();
    else dialog.current?.close();
  }, [open]);

  async function install() {
    if (busy.current || rimeRunning || result?.status !== "available") return;
    busy.current = true;
    setInstalling(true);
    onInstallingChange(true);
    setError("");
    setProgress({ downloaded: 0, total: 0, finished: false });
    try {
      await invoke("install_app_update");
      // Keep the dialog locked while the installed application restarts.
    } catch (err) {
      setError(String(err));
      busy.current = false;
      setInstalling(false);
      onInstallingChange(false);
    }
  }

  const available = result?.status === "available";
  const percentage =
    progress.total > 0
      ? Math.min(100, (progress.downloaded / progress.total) * 100)
      : undefined;

  return (
    <>
      <button
        className={`app-update-link${available ? " available" : ""}`}
        onClick={() => (available ? setOpen(true) : void checkUpdate())}
        disabled={checking || installing}
      >
        <RefreshCw size={13} className={checking ? "spin" : undefined} />
        {checking
          ? "正在检查更新"
          : available
            ? `发现新版本 v${result.version}`
            : "检查应用更新"}
      </button>
      <dialog
        ref={dialog}
        className="app-update-dialog"
        aria-labelledby="app-update-title"
        onCancel={(event) => {
          if (installing) event.preventDefault();
          else setOpen(false);
        }}
      >
        <div className="app-update-heading">
          <div>
            <h2 id="app-update-title">应用更新</h2>
            <p>当前版本 v{currentVersion} · OMR</p>
          </div>
          <button
            className="icon-button"
            aria-label="关闭更新窗口"
            disabled={installing}
            onClick={() => setOpen(false)}
          >
            <X size={18} />
          </button>
        </div>
        <label className="update-channel">
          <input
            type="checkbox"
            checked={includePrerelease}
            disabled={checking || installing}
            onChange={(event) => {
              setIncludePrerelease(event.target.checked);
              setResult(null);
              setError("");
            }}
          />
          接收测试版
        </label>
        <div className="app-update-message" role="status">
          {checking && <p>正在检查新版本…</p>}
          {result?.status === "upToDate" && <p>当前已是最新版本。</p>}
          {result?.status === "noRelease" && (
            <p>当前渠道还没有可用的应用更新。</p>
          )}
          {result?.status === "unsupported" && (
            <p>开发模式不安装应用更新，请使用安装包检查。</p>
          )}
          {available && (
            <>
              <h3>新版本 v{result.version}</h3>
              <pre className="app-release-notes">
                {result.body || "此版本暂无更新说明。"}
              </pre>
            </>
          )}
          {installing && (
            <div className="app-download-progress">
              <progress max={100} value={percentage} />
              <p>
                {progress.finished
                  ? "正在验证并安装，完成后将重新启动…"
                  : `正在下载更新${percentage === undefined ? "…" : ` ${percentage.toFixed(0)}%`}`}
              </p>
            </div>
          )}
        </div>
        {error && (
          <p className="app-update-error" role="alert">
            {error}
          </p>
        )}
        {available && !installing && (
          <p className="app-update-hint">
            安装完成后会重新启动应用。
            {rimeRunning
              ? "请先等待 Rime 更新任务完成。"
              : "请保存其他未完成的操作。"}
          </p>
        )}
        <div className="app-update-actions">
          <button
            className="ghost-button"
            disabled={checking || installing}
            onClick={() => void checkUpdate()}
          >
            <RefreshCw size={14} />
            重新检查
          </button>
          {available && (
            <button
              className="install-update-button"
              disabled={checking || installing || rimeRunning}
              onClick={() => void install()}
            >
              <Download size={14} />
              {installing ? "正在安装…" : "下载并安装"}
            </button>
          )}
        </div>
      </dialog>
    </>
  );
}
