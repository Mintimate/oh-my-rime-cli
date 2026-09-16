# Oh My Rime 桌面端使用说明

当前桌面端使用 Tauri 2 + React + TypeScript + Rust 构建。

## 功能

- 更新薄荷方案
- 更新万象模型
- 更新万象词库
- 使用自定义 `.zip` 或 `.gram` 链接
- 选择 macOS、Linux、Windows 的 Rime 配置目录
- 更新前自动创建备份
- 更新失败时尝试恢复备份
- 实时进度与运行日志
- 明暗主题

## 开发

```bash
npm install
npm run dev
```

## 构建

```bash
npm run build
```

macOS 上只验证应用 Bundle：

```bash
npx tauri build --bundles app
```

GUI 安装包只包含桌面程序；独立 CLI 通过 `cli` Cargo feature 单独构建。macOS 本地打包如需与 CI 使用相同的 ad-hoc 签名：

```bash
APPLE_SIGNING_IDENTITY=- npm run build -- --bundles app,dmg
```

ad-hoc 签名不包含 Apple 公证。首次打开下载的应用时，如提示“Apple 无法验证”，请确认来源后到“系统设置 → 隐私与安全 → 仍要打开”确认。后续应用自动更新的产物签名需要单独配置，不能代替 macOS 签名和公证。

## 更新流程

1. 选择 Rime 目标目录。
2. 选择方案、模型、词库或自定义资源。
3. 程序先下载到临时文件。
4. 下载完成后创建目标目录备份。
5. 执行安全解压或模型替换。
6. 失败时尝试恢复备份。
7. 成功后重新部署 Rime 配置。

## CLI

GUI 和 CLI 共用同一个 Rust 更新核心：

```bash
cargo run --manifest-path src-tauri/Cargo.toml --features cli --bin oh-my-rime-cli
```
