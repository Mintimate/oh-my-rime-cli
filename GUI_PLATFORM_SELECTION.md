# Rime 平台目录选择

Tauri 桌面端提供平台目录预设，也支持手动填写和原生目录选择器。

## Windows

优先读取：

```text
HKEY_CURRENT_USER\Software\Rime\Weasel\RimeUserDir
```

注册表读取失败时回退到：

```text
%APPDATA%\Rime
```

## macOS

- 鼠须管：`~/Library/Rime`
- 小企鹅：`~/.local/share/fcitx5/rime`

## Linux

- iBus：`~/.config/ibus/rime`
- Fcitx5：`~/.local/share/fcitx5/rime`
- Fcitx5 Flatpak：`~/.var/app/org.fcitx.Fcitx5/data/fcitx5/rime`

每次更新都会先备份目标目录。自定义配置文件 `*.custom.yaml` 默认不会被上游资源覆盖。
