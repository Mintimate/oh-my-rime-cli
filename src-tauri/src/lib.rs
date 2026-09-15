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
    action: String,
    target_dir: String,
    custom_url: Option<String>,
) -> UpdateResult {
    let action = match action.as_str() {
        "main" => Action::Main,
        "model" => Action::Model,
        "dict" => Action::Dict,
        "custom" => Action::Custom {
            url: custom_url.unwrap_or_default(),
        },
        _ => {
            return UpdateResult {
                success: false,
                error: Some("未知的更新类型".into()),
            }
        }
    };
    let result = tauri::async_runtime::spawn_blocking(move || {
        rime_core::execute_update(action, target_dir, |progress: UpdateProgress| {
            let _ = app.emit("update-progress", progress);
        })
    })
    .await;
    match result {
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
    }
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
fn open_directory_in_manager(path: String) -> Result<(), String> {
    #[cfg(target_os = "macos")]
    let command = ("open", vec![path.as_str()]);
    #[cfg(target_os = "linux")]
    let command = ("xdg-open", vec![path.as_str()]);
    #[cfg(target_os = "windows")]
    let command = ("explorer", vec![path.as_str()]);
    std::process::Command::new(command.0)
        .args(command.1)
        .spawn()
        .map(|_| ())
        .map_err(|e| e.to_string())
}

pub fn run() {
    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![
            get_system_info,
            execute_update,
            select_directory,
            open_directory_in_manager
        ])
        .run(tauri::generate_context!())
        .expect("error while running Oh My Rime");
}
