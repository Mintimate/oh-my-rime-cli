import { invoke } from "@tauri-apps/api/core";
import { listen, type UnlistenFn } from "@tauri-apps/api/event";
import { useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import {
  BookOpen,
  CheckCircle2,
  ChevronRight,
  CircleHelp,
  Copy,
  FolderOpen,
  History,
  LoaderCircle,
  Monitor,
  Moon,
  PackageCheck,
  Settings2,
  Sun,
  Terminal,
  UploadCloud,
  XCircle,
} from "lucide-react";
import "./styles.css";

type TargetOption = { label: string; path: string };
type SystemInfo = { os: string; options: TargetOption[] };
type Progress = { phase: string; percentage: number | null; message: string };
type Result = { success: boolean; error: string | null };
type Action = "main" | "model" | "dict" | "custom";
type Theme = "light" | "system" | "dark";

const appIcon = new URL("../src-tauri/icons/128x128.png", import.meta.url).href;
const themes = [
  { value: "light", label: "亮色", icon: Sun },
  { value: "system", label: "跟随系统", icon: Monitor },
  { value: "dark", label: "暗色", icon: Moon },
] as const;

const actions: {
  id: Action;
  icon: typeof PackageCheck;
  title: string;
  description: string;
  tone: string;
}[] = [
  {
    id: "main",
    icon: PackageCheck,
    title: "薄荷方案",
    description: "同步最新基础方案，保留用户自定义配置。",
    tone: "mint",
  },
  {
    id: "model",
    icon: UploadCloud,
    title: "万象模型",
    description: "更新 wanxiang-lts-zh-hans.gram 模型文件。",
    tone: "violet",
  },
  {
    id: "dict",
    icon: BookOpen,
    title: "万象词库",
    description: "只更新压缩包中的 dicts 词库目录。",
    tone: "amber",
  },
  {
    id: "custom",
    icon: Terminal,
    title: "自定义资源",
    description: "粘贴 zip 或 gram 直链进行更新。",
    tone: "blue",
  },
];

function App() {
  const [activeTab, setActiveTab] = useState<"update" | "logs" | "help">(
    "update",
  );
  const [theme, setTheme] = useState<Theme>(() => {
    const saved = localStorage.getItem("oh-my-rime-theme");
    return saved === "light" || saved === "dark" ? saved : "system";
  });
  const [system, setSystem] = useState<SystemInfo | null>(null);
  const [targetDir, setTargetDir] = useState("");
  const [customUrl, setCustomUrl] = useState("");
  const [running, setRunning] = useState(false);
  const [progress, setProgress] = useState<Progress>({
    phase: "ready",
    percentage: 0,
    message: "准备就绪",
  });
  const [logs, setLogs] = useState<string[]>([]);

  useEffect(() => {
    const preference = window.matchMedia("(prefers-color-scheme: dark)");
    const applyTheme = () => {
      document.documentElement.dataset.theme =
        theme === "system" ? (preference.matches ? "dark" : "light") : theme;
    };
    applyTheme();
    localStorage.setItem("oh-my-rime-theme", theme);
    if (theme === "system") {
      preference.addEventListener("change", applyTheme);
      return () => preference.removeEventListener("change", applyTheme);
    }
  }, [theme]);

  useEffect(() => {
    void invoke<SystemInfo>("get_system_info")
      .then((value) => {
        setSystem(value);
        setTargetDir(value.options[0]?.path ?? "");
      })
      .catch((error) => addLog(`读取系统信息失败：${String(error)}`));
    let stop: UnlistenFn | undefined;
    void listen<Progress>("update-progress", (event) => {
      setProgress(event.payload);
      addLog(`[${event.payload.phase}] ${event.payload.message}`);
    }).then((unlisten) => {
      stop = unlisten;
    });
    return () => stop?.();
  }, []);

  const currentPercent = Math.max(0, Math.min(100, progress.percentage ?? 0));
  const selectedLabel = useMemo(
    () =>
      system?.options.find((option) => option.path === targetDir)?.label ??
      "自定义目录",
    [system, targetDir],
  );

  function addLog(message: string) {
    setLogs((items) => [
      ...items.slice(-199),
      `${new Date().toLocaleTimeString()}  ${message}`,
    ]);
  }

  async function chooseDirectory() {
    try {
      const selected = await invoke<string | null>("select_directory");
      if (selected) setTargetDir(selected);
    } catch (error) {
      addLog(`打开目录选择器失败：${String(error)}`);
    }
  }

  async function runAction(action: Action) {
    if (running) return;
    if (!targetDir.trim()) {
      setActiveTab("update");
      addLog("请先选择或填写 Rime 配置目录");
      return;
    }
    if (action === "custom" && !customUrl.trim()) {
      addLog("请填写 zip 或 gram 资源链接");
      return;
    }
    setRunning(true);
    setActiveTab("update");
    setProgress({ phase: "starting", percentage: 0, message: "正在准备更新" });
    addLog(
      `开始更新：${actions.find((item) => item.id === action)?.title ?? action}`,
    );
    try {
      const result = await invoke<Result>("execute_update", {
        action,
        targetDir,
        customUrl: action === "custom" ? customUrl.trim() : null,
      });
      if (result.success) {
        setProgress({
          phase: "done",
          percentage: 100,
          message: "更新完成，请重新部署 Rime",
        });
        addLog("更新完成，请重新部署 Rime");
      } else {
        setProgress({
          phase: "error",
          percentage: currentPercent,
          message: result.error ?? "更新失败",
        });
        addLog(`更新失败：${result.error ?? "未知错误"}`);
      }
    } catch (error) {
      setProgress({
        phase: "error",
        percentage: currentPercent,
        message: "更新任务异常",
      });
      addLog(`更新任务异常：${String(error)}`);
    } finally {
      setRunning(false);
    }
  }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <img className="brand-mark" src={appIcon} alt="" />
          <div>
            <strong>Oh My Rime</strong>
            <span>配置管理器</span>
          </div>
        </div>
        <nav className="nav-list">
          <button
            className={activeTab === "update" ? "nav-item active" : "nav-item"}
            onClick={() => setActiveTab("update")}
          >
            <PackageCheck size={18} />
            方案更新
          </button>
          <button
            className={activeTab === "logs" ? "nav-item active" : "nav-item"}
            onClick={() => setActiveTab("logs")}
          >
            <History size={18} />
            运行日志
          </button>
          <button
            className={activeTab === "help" ? "nav-item active" : "nav-item"}
            onClick={() => setActiveTab("help")}
          >
            <CircleHelp size={18} />
            使用说明
          </button>
        </nav>
        <div className="sidebar-bottom">
          <div className="theme-switch" role="group" aria-label="外观">
            {themes.map(({ value, label, icon: Icon }) => (
              <button
                key={value}
                className={theme === value ? "selected" : ""}
                aria-pressed={theme === value}
                onClick={() => setTheme(value)}
              >
                <Icon size={15} />
                {label}
              </button>
            ))}
          </div>
          <span className="version">v2.1.0 · OMR</span>
        </div>
      </aside>
      <main className="main-panel">
        <header className="topbar">
          <div>
            <p className="eyebrow">RIME WORKSPACE</p>
            <h1>
              {activeTab === "update"
                ? "让输入法保持最新"
                : activeTab === "logs"
                  ? "运行日志"
                  : "安全地管理你的 Rime"}
            </h1>
          </div>
        </header>
        {activeTab === "update" && (
          <>
            <section className="hero">
              <div>
                <span className="hero-kicker">SAFE UPDATE PIPELINE</span>
                <h2>先备份，再更新。</h2>
                <p>
                  所有资源先下载到临时文件，更新失败会尝试恢复更新前的配置。
                </p>
              </div>
              <div className="hero-icon">
                <Settings2 size={30} />
              </div>
            </section>
            <section className="workspace-grid">
              <div className="resource-column">
                <div className="section-heading">
                  <div>
                    <h3>选择更新内容</h3>
                    <p>选择一项操作开始任务</p>
                  </div>
                  <span className="badge">4 项功能</span>
                </div>
                <div className="action-grid">
                  {actions.map(
                    ({ id, icon: Icon, title, description, tone }) => (
                      <button
                        key={id}
                        className={`action-card ${tone}`}
                        onClick={() =>
                          id === "custom"
                            ? document.getElementById("custom-url")?.focus()
                            : void runAction(id)
                        }
                        disabled={running}
                      >
                        <span className="action-icon">
                          <Icon size={21} />
                        </span>
                        <span className="action-copy">
                          <strong>{title}</strong>
                          <small>{description}</small>
                        </span>
                        <ChevronRight size={18} className="action-arrow" />
                      </button>
                    ),
                  )}
                </div>
                <div className="custom-row">
                  <label htmlFor="custom-url">自定义资源链接</label>
                  <div className="input-with-action">
                    <input
                      id="custom-url"
                      value={customUrl}
                      onChange={(event) => setCustomUrl(event.target.value)}
                      placeholder="https://example.com/resource.zip"
                    />
                    <button
                      onClick={() => void runAction("custom")}
                      disabled={running || !customUrl.trim()}
                    >
                      更新
                    </button>
                  </div>
                </div>
              </div>
              <div className="side-column">
                <div className="panel-card">
                  <div className="section-heading compact">
                    <div>
                      <h3>目标目录</h3>
                      <p>将写入所选 Rime 配置目录</p>
                    </div>
                    <button
                      className="icon-button"
                      onClick={() => void chooseDirectory()}
                      title="选择目录"
                    >
                      <FolderOpen size={19} />
                    </button>
                  </div>
                  <div className="target-options">
                    {system?.options.map((option) => (
                      <button
                        key={option.path}
                        className={
                          targetDir === option.path
                            ? "target-option selected"
                            : "target-option"
                        }
                        onClick={() => setTargetDir(option.path)}
                      >
                        <span>
                          <strong>{option.label}</strong>
                          <small>{option.path}</small>
                        </span>
                        {targetDir === option.path && (
                          <CheckCircle2 size={18} />
                        )}
                      </button>
                    ))}
                  </div>
                  <label className="field-label" htmlFor="target-dir">
                    自定义路径
                  </label>
                  <input
                    id="target-dir"
                    className="full-input"
                    value={targetDir}
                    onChange={(event) => setTargetDir(event.target.value)}
                    placeholder="/path/to/rime"
                  />
                  <p className="selected-path">当前：{selectedLabel}</p>
                </div>
                <div className="panel-card progress-card">
                  <div className="section-heading compact">
                    <div>
                      <h3>任务进度</h3>
                      <p>{progress.message}</p>
                    </div>
                    {running ? (
                      <LoaderCircle className="spin" size={20} />
                    ) : progress.phase === "error" ? (
                      <XCircle className="danger" size={20} />
                    ) : (
                      <CheckCircle2 className="success" size={20} />
                    )}
                  </div>
                  <div className="progress-track">
                    <span style={{ width: `${currentPercent}%` }} />
                  </div>
                  <div className="progress-meta">
                    <span>{progress.phase}</span>
                    <strong>{currentPercent.toFixed(0)}%</strong>
                  </div>
                </div>
              </div>
            </section>
          </>
        )}
        {activeTab === "logs" && (
          <section className="panel-card log-panel">
            <div className="log-toolbar">
              <div>
                <h3>本次任务日志</h3>
                <p>日志仅保存在当前应用会话中</p>
              </div>
              <button
                className="ghost-button"
                onClick={() => navigator.clipboard.writeText(logs.join("\n"))}
              >
                <Copy size={16} />
                复制
              </button>
            </div>
            <pre>{logs.length ? logs.join("\n") : "还没有运行日志"}</pre>
          </section>
        )}
        {activeTab === "help" && (
          <section className="help-grid">
            <div className="panel-card">
              <BookOpen size={22} />
              <h3>更新流程</h3>
              <p>
                选择资源后，程序会先下载并检查文件，再创建当前目录备份，最后执行更新。出现错误时会尝试恢复备份。
              </p>
            </div>
            <div className="panel-card">
              <Settings2 size={22} />
              <h3>部署生效</h3>
              <p>
                更新完成后，请使用对应输入法的重新部署功能，让新的方案、词库或模型加载生效。
              </p>
            </div>
            <div className="panel-card">
              <Terminal size={22} />
              <h3>本地优先</h3>
              <p>
                工具不上传你的 Rime
                配置，也不包含遥测服务。资源下载地址和更新内容可以在日志中查看。
              </p>
            </div>
          </section>
        )}
      </main>
    </div>
  );
}

const rootElement = document.getElementById("root");
if (!rootElement) {
  throw new Error("找不到应用挂载节点 #root");
}

createRoot(rootElement).render(<App />);
