mod app_update;
pub mod rime_core;

use rime_core::{Action, SystemInfo, UpdateProgress};
use serde::Serialize;
use tauri::{AppHandle, Emitter};

#[derive(Debug, Clone, Serialize)]
#[serde(rename_all = "camelCase")]
struct UpdateResult {
    success: bool,
    error: Option<String>,
}

#[tauri::command]
fn get_system_info() -> SystemInfo {
    rime_core::system_info()
}

#[tauri::command]
async fn execute_update(
    app: AppHandle,
    state: tauri::State<'_, app_update::AppUpdateState>,
    action: String,
    target_dir: String,
    custom_url: Option<String>,
) -> Result<UpdateResult, String> {
    let _task = state
        .task_gate
        .try_lock()
        .map_err(|_| "已有更新任务正在执行，请稍后再试".to_string())?;
    let action = match action.as_str() {
        "main" => Action::Main,
        "model" => Action::Model,
        "dict" => Action::Dict,
        "custom" => Action::Custom {
            url: custom_url.unwrap_or_default(),
        },
        _ => {
            return Ok(UpdateResult {
                success: false,
                error: Some("未知的更新类型".into()),
            })
        }
    };
    let result = tauri::async_runtime::spawn_blocking(move || {
        rime_core::execute_update(action, target_dir, |progress: UpdateProgress| {
            let _ = app.emit("update-progress", progress);
        })
    })
    .await;
    Ok(match result {
        Ok(Ok(())) => UpdateResult {
            success: true,
            error: None,
        },
        Ok(Err(error)) => UpdateResult {
            success: false,
            error: Some(error),
        },
        Err(error) => UpdateResult {
            success: false,
            error: Some(format!("更新任务异常: {error}")),
        },
    })
}

#[tauri::command]
async fn select_directory() -> Option<String> {
    tauri::async_runtime::spawn_blocking(|| {
        rfd::FileDialog::new()
            .set_title("选择 Rime 配置目录")
            .pick_folder()
            .map(|path| path.to_string_lossy().into_owned())
    })
    .await
    .ok()
    .flatten()
}

#[tauri::command]
async fn open_directory_in_manager(path: String) -> Result<(), String> {
    tauri::async_runtime::spawn_blocking(move || {
        if path.trim().is_empty() {
            return Err("请先选择或填写目标目录".into());
        }
        let path = rime_core::expand_home(&path)
            .canonicalize()
            .map_err(|error| format!("目录不存在或无法访问：{error}"))?;
        if !path.is_dir() {
            return Err("所选路径不是目录".into());
        }
        tauri_plugin_opener::open_path(&path, None::<&str>)
            .map_err(|error| format!("无法启动文件管理器：{error}"))
    })
    .await
    .map_err(|error| format!("打开目录任务异常：{error}"))?
}

pub fn run() {
    tauri::Builder::default()
        .plugin(
            tauri_plugin_opener::Builder::new()
                .open_js_links_on_click(false)
                .build(),
        )
        .plugin(tauri_plugin_updater::Builder::new().build())
        .manage(app_update::AppUpdateState::default())
        .invoke_handler(tauri::generate_handler![
            get_system_info,
            execute_update,
            select_directory,
            open_directory_in_manager,
            app_update::get_app_version,
            app_update::check_app_update,
            app_update::install_app_update
        ])
        .run(tauri::generate_context!())
        .expect("error while running Oh My Rime");
}
