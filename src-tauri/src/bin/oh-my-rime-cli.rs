#[path = "../rime_core.rs"]
mod rime_core;

use rime_core::{Action, UpdateProgress};
use std::io::{self, Write};

fn main() {
    println!("Oh My Rime CLI 2.1");
    println!("输入 1 更新薄荷方案，2 更新万象模型，3 更新万象词库，4 自定义，q 退出");
    loop {
        print!("\n请选择操作: ");
        let _ = io::stdout().flush();
        let mut choice = String::new();
        if io::stdin().read_line(&mut choice).is_err() {
            return;
        }
        let choice = choice.trim();
        if choice == "q" {
            return;
        }
        let action = match choice {
            "1" => Action::Main,
            "2" => Action::Model,
            "3" => Action::Dict,
            "4" => {
                print!("请输入 zip 或 gram URL: ");
                let _ = io::stdout().flush();
                let mut url = String::new();
                if io::stdin().read_line(&mut url).is_err() {
                    continue;
                }
                Action::Custom {
                    url: url.trim().to_string(),
                }
            }
            _ => {
                println!("无效选项");
                continue;
            }
        };
        let target = choose_target();
        println!("目标目录: {target}");
        let result = rime_core::execute_update(action, target, |progress: UpdateProgress| {
            println!("[{}] {}", progress.phase, progress.message);
        });
        match result {
            Ok(_) => println!("更新完成，请重新部署 Rime。"),
            Err(error) => eprintln!("更新失败: {error}"),
        }
    }
}

fn choose_target() -> String {
    let info = rime_core::system_info();
    if info.options.len() == 1 {
        return info.options[0].path.clone();
    }
    for (index, option) in info.options.iter().enumerate() {
        println!("{}: {} ({})", index + 1, option.label, option.path);
    }
    print!("选择目录，默认 1: ");
    let _ = io::stdout().flush();
    let mut input = String::new();
    let _ = io::stdin().read_line(&mut input);
    input
        .trim()
        .parse::<usize>()
        .ok()
        .and_then(|index| info.options.get(index.saturating_sub(1)))
        .map(|option| option.path.clone())
        .unwrap_or_else(|| info.options[0].path.clone())
}
