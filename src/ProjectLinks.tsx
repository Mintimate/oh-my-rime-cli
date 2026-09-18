import { isTauri } from "@tauri-apps/api/core";
import { openUrl } from "@tauri-apps/plugin-opener";
import { ExternalLink, Github, Globe } from "lucide-react";
import { useState, type MouseEvent } from "react";

const links = [
  { label: "官方网站", url: "https://www.mintimate.cc", icon: Globe },
  {
    label: "GitHub 项目",
    url: "https://github.com/Mintimate/oh-my-rime-cli",
    icon: Github,
  },
];

export function ProjectLinks() {
  const [error, setError] = useState<string | null>(null);

  async function openLink(event: MouseEvent<HTMLAnchorElement>, url: string) {
    if (!isTauri()) return;
    event.preventDefault();
    setError(null);
    try {
      await openUrl(url);
    } catch {
      setError("打开链接失败，请重试或复制链接到浏览器。");
    }
  }

  return (
    <nav className="project-links" aria-label="官方网站与项目地址">
      {links.map(({ label, url, icon: Icon }) => (
        <a
          key={url}
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          title={url}
          onClick={(event) => void openLink(event, url)}
        >
          <Icon size={16} aria-hidden="true" />
          <span>{label}</span>
          <ExternalLink size={13} aria-hidden="true" />
        </a>
      ))}
      {error && (
        <p className="project-links-error" role="alert">
          {error}
        </p>
      )}
    </nav>
  );
}
