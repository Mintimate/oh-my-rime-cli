use chrono::Local;
use reqwest::blocking::Client;
use serde::Serialize;
use std::{
    fs,
    fs::File,
    io::{self, Read, Write},
    path::{Path, PathBuf},
    time::{Duration, Instant, SystemTime, UNIX_EPOCH},
};
use tempfile::NamedTempFile;
use zip::ZipArchive;

pub const OH_MY_RIME_REPO: &str =
    "https://cnb.cool/Mintimate/rime/oh-my-rime/-/releases/download/latest/oh-my-rime.zip";
pub const WANXIANG_GRAM: &str = "https://cnb.cool/Mintimate/rime/oh-my-rime/-/releases/download/latest/wanxiang-lts-zh-hans.gram";
const MAX_DOWNLOAD_BYTES: u64 = 2 * 1024 * 1024 * 1024;
const MAX_EXTRACTED_BYTES: u64 = 4 * 1024 * 1024 * 1024;
const BACKUP_KEEP_COUNT: usize = 3;

#[derive(Debug, Clone)]
pub enum Action {
    Main,
    Model,
    Dict,
    Custom { url: String },
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct UpdateProgress {
    pub phase: String,
    pub percentage: Option<f64>,
    pub message: String,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct TargetOption {
    pub label: String,
    pub path: String,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct SystemInfo {
    pub os: String,
    pub options: Vec<TargetOption>,
}

pub fn system_info() -> SystemInfo {
    let home = std::env::var("HOME").unwrap_or_else(|_| "~".to_string());
    #[cfg(target_os = "windows")]
    let options = vec![TargetOption {
        label: "Rime（Weasel）".into(),
        path: windows_rime_dir(),
    }];
    #[cfg(target_os = "macos")]
    let options = vec![
        TargetOption {
            label: "鼠须管".into(),
            path: format!("{home}/Library/Rime"),
        },
        TargetOption {
            label: "小企鹅".into(),
            path: format!("{home}/.local/share/fcitx5/rime"),
        },
    ];
    #[cfg(target_os = "linux")]
    let options = vec![
        TargetOption {
            label: "iBus".into(),
            path: format!("{home}/.config/ibus/rime"),
        },
        TargetOption {
            label: "Fcitx5".into(),
            path: format!("{home}/.local/share/fcitx5/rime"),
        },
        TargetOption {
            label: "Fcitx5 Flatpak".into(),
            path: format!("{home}/.var/app/org.fcitx.Fcitx5/data/fcitx5/rime"),
        },
    ];
    #[cfg(not(any(target_os = "windows", target_os = "macos", target_os = "linux")))]
    let options = vec![TargetOption {
        label: "自定义目录".into(),
        path: home.clone(),
    }];
    SystemInfo {
        os: std::env::consts::OS.to_string(),
        options,
    }
}

pub fn execute_update<F>(
    action: Action,
    target_dir: String,
    mut on_progress: F,
) -> Result<(), String>
where
    F: FnMut(UpdateProgress),
{
    let target_dir = expand_home(&target_dir);
    if target_dir.as_os_str().is_empty() {
        return Err("目标目录不能为空".into());
    }
    let (url, kind) = match action {
        Action::Main => (OH_MY_RIME_REPO.to_string(), UpdateKind::Main),
        Action::Model => (WANXIANG_GRAM.to_string(), UpdateKind::Model),
        Action::Dict => (OH_MY_RIME_REPO.to_string(), UpdateKind::Dict),
        Action::Custom { url } => {
            let kind = custom_kind(&url).ok_or("自定义链接必须以 .zip 或 .gram 结尾")?;
            (url, kind)
        }
    };

    emit(
        &mut on_progress,
        "download",
        None,
        format!("正在下载 {}", display_url(&url)),
    );
    let temp = download(&url, &mut on_progress)?;
    emit(
        &mut on_progress,
        "backup",
        None,
        "正在创建更新前备份".into(),
    );
    let backup = create_backup(&target_dir)?;
    let result = apply_update(&temp, &target_dir, kind, &mut on_progress);
    if let Err(error) = result {
        if let Some(backup_dir) = backup.as_ref() {
            emit(
                &mut on_progress,
                "rollback",
                None,
                "更新失败，正在恢复备份".into(),
            );
            restore_backup(&target_dir, backup_dir)
                .map_err(|restore| format!("{error}；备份恢复失败: {restore}"))?;
        } else if target_dir.exists() {
            fs::remove_dir_all(&target_dir)
                .map_err(|cleanup| format!("{error}；清理不完整目录失败: {cleanup}"))?;
        }
        return Err(error);
    }
    prune_backups(&target_dir)?;
    emit(
        &mut on_progress,
        "done",
        Some(100.0),
        "更新完成，请重新部署 Rime".into(),
    );
    Ok(())
}

#[derive(Debug, Clone, Copy)]
enum UpdateKind {
    Main,
    Model,
    Dict,
}

fn download<F>(url: &str, on_progress: &mut F) -> Result<NamedTempFile, String>
where
    F: FnMut(UpdateProgress),
{
    let client = Client::builder()
        .connect_timeout(Duration::from_secs(20))
        .timeout(Duration::from_secs(30 * 60))
        .build()
        .map_err(|e| format!("创建网络客户端失败: {e}"))?;
    let mut response = client
        .get(url)
        .header("User-Agent", "oh-my-rime/2")
        .send()
        .map_err(|e| format!("请求失败: {e}"))?;
    if !response.status().is_success() {
        return Err(format!("下载失败: HTTP {}", response.status()));
    }
    if response
        .headers()
        .get(reqwest::header::CONTENT_TYPE)
        .and_then(|v| v.to_str().ok())
        .is_some_and(|v| v.contains("text/html"))
    {
        return Err("下载内容是 HTML 页面，不是资源文件".into());
    }
    let total = response.content_length().unwrap_or(0);
    if total > MAX_DOWNLOAD_BYTES {
        return Err("下载文件超过安全大小限制".into());
    }
    let mut temp = NamedTempFile::new().map_err(|e| format!("创建临时文件失败: {e}"))?;
    let mut buffer = [0_u8; 64 * 1024];
    let mut downloaded = 0_u64;
    let mut last_report = Instant::now();
    loop {
        let read = response
            .read(&mut buffer)
            .map_err(|e| format!("读取下载内容失败: {e}"))?;
        if read == 0 {
            break;
        }
        downloaded += read as u64;
        if downloaded > MAX_DOWNLOAD_BYTES {
            return Err("下载文件超过安全大小限制".into());
        }
        temp.write_all(&buffer[..read])
            .map_err(|e| format!("写入临时文件失败: {e}"))?;
        if last_report.elapsed() >= Duration::from_millis(100) {
            let percentage = (total > 0).then_some(downloaded as f64 / total as f64 * 100.0);
            emit(
                on_progress,
                "download",
                percentage,
                format!(
                    "已下载 {} / {}",
                    format_bytes(downloaded),
                    if total > 0 {
                        format_bytes(total)
                    } else {
                        "未知大小".into()
                    }
                ),
            );
            last_report = Instant::now();
        }
    }
    temp.as_file_mut()
        .sync_all()
        .map_err(|e| format!("同步临时文件失败: {e}"))?;
    Ok(temp)
}

fn apply_update<F>(
    temp: &NamedTempFile,
    target: &Path,
    kind: UpdateKind,
    on_progress: &mut F,
) -> Result<(), String>
where
    F: FnMut(UpdateProgress),
{
    fs::create_dir_all(target).map_err(|e| format!("创建目标目录失败: {e}"))?;
    match kind {
        UpdateKind::Model => {
            let destination = target.join("wanxiang-lts-zh-hans.gram");
            fs::copy(temp.path(), &destination).map_err(|e| format!("写入模型失败: {e}"))?;
            emit(on_progress, "apply", Some(100.0), "模型文件已更新".into());
        }
        UpdateKind::Main | UpdateKind::Dict => {
            let file = File::open(temp.path()).map_err(|e| format!("打开 ZIP 失败: {e}"))?;
            let mut archive = ZipArchive::new(file).map_err(|e| format!("读取 ZIP 失败: {e}"))?;
            let mut extracted = 0_u64;
            for index in 0..archive.len() {
                let mut entry = archive
                    .by_index(index)
                    .map_err(|e| format!("读取 ZIP 条目失败: {e}"))?;
                let name = entry
                    .enclosed_name()
                    .ok_or_else(|| format!("ZIP 包含不安全路径: {}", entry.name()))?
                    .to_path_buf();
                if matches!(kind, UpdateKind::Dict) && !name.to_string_lossy().starts_with("dicts/")
                {
                    continue;
                }
                if name
                    .file_name()
                    .is_some_and(|n| n.to_string_lossy().ends_with(".custom.yaml"))
                {
                    continue;
                }
                extracted = extracted.saturating_add(entry.size());
                if extracted > MAX_EXTRACTED_BYTES {
                    return Err("解压内容超过安全大小限制".into());
                }
                let destination = target.join(name);
                if entry.is_dir() {
                    fs::create_dir_all(&destination).map_err(|e| format!("创建目录失败: {e}"))?;
                    continue;
                }
                if let Some(parent) = destination.parent() {
                    fs::create_dir_all(parent).map_err(|e| format!("创建目录失败: {e}"))?;
                }
                let mut output =
                    File::create(&destination).map_err(|e| format!("创建文件失败: {e}"))?;
                io::copy(&mut entry, &mut output).map_err(|e| format!("写入文件失败: {e}"))?;
                output
                    .sync_all()
                    .map_err(|e| format!("同步文件失败: {e}"))?;
                emit(
                    on_progress,
                    "apply",
                    None,
                    format!("更新 {}", destination.display()),
                );
            }
        }
    }
    Ok(())
}

fn create_backup(target: &Path) -> Result<Option<PathBuf>, String> {
    if !target.exists() {
        return Ok(None);
    }
    if !target.is_dir() {
        return Err(format!("目标路径不是目录: {}", target.display()));
    }
    let root = backup_root(target);
    fs::create_dir_all(&root).map_err(|e| format!("创建备份目录失败: {e}"))?;
    let stamp = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_err(|e| e.to_string())?
        .as_nanos();
    let backup = root.join(format!("{}-{stamp}", Local::now().format("%Y%m%d-%H%M%S")));
    copy_dir(target, &backup).map_err(|e| format!("创建备份失败: {e}"))?;
    Ok(Some(backup))
}

fn restore_backup(target: &Path, backup: &Path) -> Result<(), String> {
    if target.exists() {
        fs::remove_dir_all(target).map_err(|e| format!("删除失败目录失败: {e}"))?;
    }
    copy_dir(backup, target).map_err(|e| format!("恢复备份失败: {e}"))
}

fn prune_backups(target: &Path) -> Result<(), String> {
    let root = backup_root(target);
    let Ok(entries) = fs::read_dir(&root) else {
        return Ok(());
    };
    let mut dirs: Vec<PathBuf> = entries
        .filter_map(Result::ok)
        .map(|e| e.path())
        .filter(|p| p.is_dir())
        .collect();
    dirs.sort();
    if dirs.len() > BACKUP_KEEP_COUNT {
        let remove_count = dirs.len() - BACKUP_KEEP_COUNT;
        for path in dirs.into_iter().take(remove_count) {
            fs::remove_dir_all(path).map_err(|e| format!("清理旧备份失败: {e}"))?;
        }
    }
    Ok(())
}

fn copy_dir(source: &Path, destination: &Path) -> io::Result<()> {
    fs::create_dir_all(destination)?;
    for entry in fs::read_dir(source)? {
        let entry = entry?;
        let from = entry.path();
        let to = destination.join(entry.file_name());
        if entry.file_type()?.is_dir() {
            copy_dir(&from, &to)?;
        } else {
            fs::copy(from, to)?;
        }
    }
    Ok(())
}

#[cfg(target_os = "windows")]
fn windows_rime_dir() -> String {
    let fallback = std::env::var("APPDATA")
        .map(|app_data| {
            PathBuf::from(app_data)
                .join("Rime")
                .to_string_lossy()
                .into_owned()
        })
        .unwrap_or_else(|_| "%APPDATA%\\Rime".into());
    let hkcu = winreg::RegKey::predef(winreg::enums::HKEY_CURRENT_USER);
    hkcu.open_subkey("Software\\Rime\\Weasel")
        .and_then(|key| key.get_value::<String, _>("RimeUserDir"))
        .ok()
        .filter(|value| !value.trim().is_empty())
        .unwrap_or(fallback)
}

fn backup_root(target: &Path) -> PathBuf {
    target
        .parent()
        .unwrap_or_else(|| Path::new("."))
        .join(format!(
            "{}.backups",
            target.file_name().unwrap_or_default().to_string_lossy()
        ))
}
pub(crate) fn expand_home(path: &str) -> PathBuf {
    if path == "~" {
        return PathBuf::from(std::env::var("HOME").unwrap_or_else(|_| path.into()));
    }
    if let Some(rest) = path.strip_prefix("~/") {
        return PathBuf::from(std::env::var("HOME").unwrap_or_else(|_| "~".into())).join(rest);
    }
    PathBuf::from(path)
}
fn custom_kind(url: &str) -> Option<UpdateKind> {
    let path = url
        .split('?')
        .next()
        .unwrap_or(url)
        .split('#')
        .next()
        .unwrap_or(url)
        .to_ascii_lowercase();
    if path.ends_with(".zip") {
        Some(UpdateKind::Main)
    } else if path.ends_with(".gram") {
        Some(UpdateKind::Model)
    } else {
        None
    }
}
fn display_url(url: &str) -> String {
    url.split('?').next().unwrap_or(url).to_string()
}
fn format_bytes(bytes: u64) -> String {
    const UNITS: &[&str] = &["B", "KB", "MB", "GB"];
    let mut value = bytes as f64;
    let mut index = 0;
    while value >= 1024.0 && index < UNITS.len() - 1 {
        value /= 1024.0;
        index += 1;
    }
    format!("{value:.1} {}", UNITS[index])
}
fn emit<F: FnMut(UpdateProgress)>(
    callback: &mut F,
    phase: &str,
    percentage: Option<f64>,
    message: String,
) {
    callback(UpdateProgress {
        phase: phase.into(),
        percentage,
        message,
    });
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn custom_kind_accepts_query_string() {
        assert!(matches!(
            custom_kind("https://example.test/file.zip?download=1"),
            Some(UpdateKind::Main)
        ));
    }
    #[test]
    fn expand_home_handles_tilde() {
        assert!(expand_home("~/Rime").ends_with("Rime"));
    }

    #[test]
    fn main_update_preserves_custom_yaml() {
        use std::io::Write;
        use zip::{write::SimpleFileOptions, ZipWriter};
        let root = tempfile::tempdir().unwrap();
        let target = root.path().join("Rime");
        std::fs::create_dir_all(&target).unwrap();
        std::fs::write(target.join("default.custom.yaml"), b"user").unwrap();
        let mut archive = NamedTempFile::new().unwrap();
        {
            let mut writer = ZipWriter::new(archive.as_file_mut());
            writer
                .start_file("default.yaml", SimpleFileOptions::default())
                .unwrap();
            writer.write_all(b"new").unwrap();
            writer
                .start_file("default.custom.yaml", SimpleFileOptions::default())
                .unwrap();
            writer.write_all(b"upstream").unwrap();
            writer.finish().unwrap();
        }
        apply_update(&archive, &target, UpdateKind::Main, &mut |_| {}).unwrap();
        assert_eq!(std::fs::read(target.join("default.yaml")).unwrap(), b"new");
        assert_eq!(
            std::fs::read(target.join("default.custom.yaml")).unwrap(),
            b"user"
        );
    }
}
