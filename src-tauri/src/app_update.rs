use semver::Version;
use serde::{Deserialize, Serialize};
use std::{sync::Mutex, time::Duration};
use tauri::{AppHandle, Emitter, State};
use tauri_plugin_updater::{Update, UpdaterExt};

const RELEASES_API: &str =
    "https://api.github.com/repos/Mintimate/oh-my-rime-cli/releases?per_page=100";
const UPDATE_CHECK_TIMEOUT: Duration = Duration::from_secs(30);
const UPDATE_DOWNLOAD_TIMEOUT: Duration = Duration::from_secs(30 * 60);

#[derive(Debug, Deserialize)]
struct ReleaseAsset {
    name: String,
}

#[derive(Debug, Deserialize)]
struct Release {
    tag_name: String,
    draft: bool,
    prerelease: bool,
    assets: Vec<ReleaseAsset>,
}

// Old Go releases have no updater manifest. Never offer them to the Rust app.
fn select_release(releases: &[Release], include_prerelease: bool) -> Option<(&Release, Version)> {
    releases
        .iter()
        .filter(|release| {
            !release.draft
                && release
                    .assets
                    .iter()
                    .filter(|asset| asset.name == "latest.json")
                    .count()
                    == 1
        })
        .filter_map(|release| {
            let version = Version::parse(release.tag_name.strip_prefix('v')?).ok()?;
            if !include_prerelease && (release.prerelease || !version.pre.is_empty()) {
                return None;
            }
            Some((release, version))
        })
        .max_by(|a, b| a.1.cmp(&b.1))
}

pub struct AppUpdateState {
    pending: Mutex<Option<Update>>,
    pub task_gate: tokio::sync::Mutex<()>,
    operation_gate: tokio::sync::Mutex<()>,
}

impl Default for AppUpdateState {
    fn default() -> Self {
        Self {
            pending: Mutex::new(None),
            task_gate: tokio::sync::Mutex::new(()),
            operation_gate: tokio::sync::Mutex::new(()),
        }
    }
}

impl AppUpdateState {
    fn clear_pending(&self) -> Result<(), String> {
        *self
            .pending
            .lock()
            .map_err(|error| format!("更新状态锁已失效: {error}"))? = None;
        Ok(())
    }

    fn store_pending(&self, update: Update) -> Result<(), String> {
        *self
            .pending
            .lock()
            .map_err(|error| format!("更新状态锁已失效: {error}"))? = Some(update);
        Ok(())
    }

    fn take_pending(&self) -> Result<Update, String> {
        self.pending
            .lock()
            .map_err(|error| format!("更新状态锁已失效: {error}"))?
            .take()
            .ok_or_else(|| "没有待安装的 Oh My Rime 更新，请先检查更新".to_string())
    }
}

#[derive(Debug, Clone, Copy, Serialize)]
#[serde(rename_all = "camelCase")]
pub enum AppUpdateStatus {
    Unsupported,
    UpToDate,
    NoRelease,
    Available,
    Error,
}

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct AppUpdateCheckResult {
    pub status: AppUpdateStatus,
    pub current_version: String,
    pub version: Option<String>,
    pub body: Option<String>,
    pub date: Option<String>,
    pub reason: Option<String>,
}

#[derive(Debug, Clone, Serialize)]
#[serde(tag = "event", content = "data", rename_all = "camelCase")]
enum AppUpdateEvent {
    #[serde(rename_all = "camelCase")]
    Started {
        content_length: Option<u64>,
    },
    #[serde(rename_all = "camelCase")]
    Progress {
        chunk_length: usize,
    },
    Finished,
}

fn result(status: AppUpdateStatus, reason: Option<String>) -> AppUpdateCheckResult {
    AppUpdateCheckResult {
        status,
        current_version: env!("CARGO_PKG_VERSION").to_string(),
        version: None,
        body: None,
        date: None,
        reason,
    }
}

#[tauri::command]
pub fn get_app_version() -> String {
    env!("CARGO_PKG_VERSION").to_string()
}

#[tauri::command]
pub async fn check_app_update(
    app: AppHandle,
    state: State<'_, AppUpdateState>,
    include_prerelease: bool,
) -> Result<AppUpdateCheckResult, String> {
    let _guard = state.operation_gate.lock().await;

    if tauri::is_dev() {
        state.clear_pending()?;
        return Ok(result(AppUpdateStatus::Unsupported, None));
    }

    state.clear_pending()?;
    let releases = reqwest::Client::builder()
        .user_agent(concat!("oh-my-rime/", env!("CARGO_PKG_VERSION")))
        .timeout(UPDATE_CHECK_TIMEOUT)
        .build()
        .map_err(|error| error.to_string())?
        .get(RELEASES_API)
        .header("Accept", "application/vnd.github+json")
        .send()
        .await
        .map_err(|error| format!("无法连接 GitHub：{error}"))?
        .error_for_status()
        .map_err(|error| format!("GitHub 更新检查失败：{error}"))?
        .json::<Vec<Release>>()
        .await
        .map_err(|error| format!("无法读取发布列表：{error}"))?;
    let Some((release, version)) = select_release(&releases, include_prerelease) else {
        return Ok(result(AppUpdateStatus::NoRelease, None));
    };
    if version <= Version::parse(env!("CARGO_PKG_VERSION")).expect("valid package version") {
        return Ok(result(AppUpdateStatus::UpToDate, None));
    }
    let endpoint = format!(
        "https://github.com/Mintimate/oh-my-rime-cli/releases/download/{}/latest.json",
        release.tag_name
    )
    .parse()
    .map_err(|error| format!("无效的更新地址：{error}"))?;
    let updater = match app
        .updater_builder()
        .endpoints(vec![endpoint])
        .and_then(|builder| builder.timeout(UPDATE_CHECK_TIMEOUT).build())
    {
        Ok(updater) => updater,
        Err(error) => {
            state.clear_pending()?;
            return Ok(result(AppUpdateStatus::Error, Some(error.to_string())));
        }
    };

    let update = match updater.check().await {
        Ok(update) => update,
        Err(error) => {
            state.clear_pending()?;
            return Ok(result(AppUpdateStatus::Error, Some(error.to_string())));
        }
    };

    let Some(mut update) = update else {
        state.clear_pending()?;
        return Ok(result(AppUpdateStatus::UpToDate, None));
    };

    if update.version != version.to_string() {
        return Err("更新清单版本与发布标签不一致".into());
    }

    // GitHub Release assets can be much larger than the update manifest and may
    // be downloaded through a slower redirected CDN. Keep checks responsive,
    // but allow enough time for the signed application package itself.
    update.timeout = Some(UPDATE_DOWNLOAD_TIMEOUT);

    let check_result = AppUpdateCheckResult {
        status: AppUpdateStatus::Available,
        current_version: update.current_version.clone(),
        version: Some(update.version.clone()),
        body: update.body.clone(),
        date: update.date.as_ref().map(ToString::to_string),
        reason: None,
    };
    state.store_pending(update)?;
    Ok(check_result)
}

#[tauri::command]
pub async fn install_app_update(
    app: AppHandle,
    state: State<'_, AppUpdateState>,
) -> Result<bool, String> {
    let _guard = state.operation_gate.lock().await;
    let _task = state
        .task_gate
        .try_lock()
        .map_err(|_| "正在更新 Rime 配置，请等待任务完成后再安装应用更新".to_string())?;
    let update = state.take_pending()?;
    let app_for_started = app.clone();
    let app_for_progress = app.clone();
    let app_for_finished = app.clone();
    let mut started = false;

    let download = update
        .download(
            move |chunk_length, content_length| {
                if !started {
                    let _ = app_for_started.emit(
                        "app-update-event",
                        AppUpdateEvent::Started { content_length },
                    );
                    started = true;
                }
                let _ = app_for_progress.emit(
                    "app-update-event",
                    AppUpdateEvent::Progress { chunk_length },
                );
            },
            move || {
                let _ = app_for_finished.emit("app-update-event", AppUpdateEvent::Finished);
            },
        )
        .await;

    let bytes = match download {
        Ok(bytes) => bytes,
        Err(error) => {
            state.store_pending(update)?;
            return Err(error.to_string());
        }
    };

    if let Err(error) = update.install(bytes) {
        state.store_pending(update)?;
        return Err(error.to_string());
    }
    app.request_restart();
    Ok(true)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn release(tag: &str, prerelease: bool, draft: bool, manifest: bool) -> Release {
        Release {
            tag_name: tag.into(),
            prerelease,
            draft,
            assets: if manifest {
                vec![ReleaseAsset {
                    name: "latest.json".into(),
                }]
            } else {
                vec![]
            },
        }
    }

    #[test]
    fn ignores_old_go_releases_and_incomplete_drafts() {
        let releases = vec![
            release("v3.0.0", false, false, false),
            release("v4.0.0", false, true, true),
        ];
        assert!(select_release(&releases, true).is_none());
    }

    #[test]
    fn stable_channel_excludes_all_prereleases() {
        let releases = vec![
            release("v2.2.0-test.1", true, false, true),
            release("v2.3.0-test.1", false, false, true),
            release("v2.4.0", true, false, true),
            release("v2.1.1", false, false, true),
        ];
        assert_eq!(
            select_release(&releases, false).unwrap().0.tag_name,
            "v2.1.1"
        );
    }

    #[test]
    fn preview_channel_uses_semver_not_publication_order() {
        let releases = vec![
            release("v2.2.0-test.2", true, false, true),
            release("v2.1.9", false, false, true),
            release("v2.2.0-test.10", true, false, true),
        ];
        assert_eq!(
            select_release(&releases, true).unwrap().0.tag_name,
            "v2.2.0-test.10"
        );
    }

    #[test]
    fn stable_release_supersedes_its_preview() {
        let releases = vec![
            release("v2.2.0-test.10", true, false, true),
            release("v2.2.0", false, false, true),
        ];
        assert_eq!(
            select_release(&releases, true).unwrap().0.tag_name,
            "v2.2.0"
        );
    }

    #[test]
    fn rejects_invalid_tags_and_duplicate_manifests() {
        let mut duplicate = release("v2.2.0", false, false, true);
        duplicate.assets.push(ReleaseAsset {
            name: "latest.json".into(),
        });
        let releases = vec![release("../../other", false, false, true), duplicate];
        assert!(select_release(&releases, true).is_none());
    }
}
