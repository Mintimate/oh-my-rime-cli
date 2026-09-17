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

GUI 安装包只包含桌面程序；独立 CLI 通过 `cli` Cargo feature 单独构建。macOS 已在 `src-tauri/tauri.conf.json` 中默认配置 ad-hoc 签名（`bundle.macOS.signingIdentity: "-"`），本地打包和 CI 使用相同方式，无需提供签名密钥：

```bash
npm run build -- --bundles app,dmg
```

ad-hoc 签名不包含 Apple 公证。首次打开下载的应用时，如提示“Apple 无法验证”，请确认来源后到“系统设置 → 隐私与安全 → 仍要打开”确认，按系统提示输入 Mac 登录密码或使用触控 ID。具体提示取决于 macOS 版本和安全策略。

构建后可检查应用签名；签名后不要再修改 Bundle 内容：

```bash
codesign --verify --deep --strict --verbose=2 "src-tauri/target/release/bundle/macos/Oh My Rime.app"
codesign --display --verbose=4 "src-tauri/target/release/bundle/macos/Oh My Rime.app"
```

验证应成功，默认签名信息应包含 `Signature=adhoc`。交叉编译时需在 `target/` 后加入对应的 target triple。后续应用自动更新所用的 `TAURI_SIGNING_PRIVATE_KEY` 需要单独配置，与这里的 macOS 签名不同，不能代替 Developer ID 签名和 Apple 公证。

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
