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
- 亮色、跟随系统、暗色主题
- 应用更新检查、下载进度与签名校验

## 开发

```bash
npm install
npm run dev
```

## 构建

```bash
npm run build -- --config '{"bundle":{"createUpdaterArtifacts":false}}'
```

macOS 上只验证应用 Bundle：

```bash
npx tauri build --bundles app --config '{"bundle":{"createUpdaterArtifacts":false}}'
```

GUI 安装包只包含桌面程序；独立 CLI 通过 `cli` Cargo feature 单独构建。macOS 已在 `src-tauri/tauri.conf.json` 中默认配置 ad-hoc 签名（`bundle.macOS.signingIdentity: "-"`），本地打包和 CI 使用相同方式，macOS ad-hoc 签名无需证书；生成自动更新产物时还需提供独立的更新私钥：

```bash
TAURI_SIGNING_PRIVATE_KEY="/path/to/updater.key" npm run build -- --bundles app,dmg
```

ad-hoc 签名不包含 Apple 公证。首次打开下载的应用时，如提示“Apple 无法验证”，请确认来源后到“系统设置 → 隐私与安全 → 仍要打开”确认，按系统提示输入 Mac 登录密码或使用触控 ID。具体提示取决于 macOS 版本和安全策略。

构建后可检查应用签名；签名后不要再修改 Bundle 内容：

```bash
codesign --verify --deep --strict --verbose=2 "src-tauri/target/release/bundle/macos/Oh My Rime.app"
codesign --display --verbose=4 "src-tauri/target/release/bundle/macos/Oh My Rime.app"
```

验证应成功，默认签名信息应包含 `Signature=adhoc`。交叉编译时需在 `target/` 后加入对应的 target triple。应用自动更新所用的 `TAURI_SIGNING_PRIVATE_KEY` 需要单独配置，与这里的 macOS 签名不同，不能代替 Developer ID 签名和 Apple 公证。

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

## 应用自动更新

桌面端启动后自动检查新版本，左下角的“检查应用更新”可手动检查。发现更新后可以查看版本说明，点击“下载并安装”；应用会验证签名，安装成功后重新启动。Rime 配置更新进行中不能安装应用更新，失败时可重试。

- 默认接收正式版；测试版安装包首次运行默认接收测试版。可以在更新窗口切换“接收测试版”，选择会保存在本机。
- 更新检查读取 GitHub 最近 100 个 Release，只选择带 `latest.json` 的发布，并按 SemVer 比较版本。旧 Go 版 `v3.0.0` 没有该清单，不会作为 Rust 桌面端更新。
- 网络错误或 GitHub 限流不会阻止使用应用，手动检查时会显示错误。开发模式不安装更新。
- macOS 建议从 DMG 拖到“应用程序”后运行，不要直接在只读 DMG 中安装更新。Linux 使用对应的 AppImage/deb/rpm 更新产物；系统级安装可能需要系统授权。
- 独立 CLI 不自动更新。没有更新功能的旧桌面版本需要先手动安装一次新版本。

## 发布签名更新

1. 使用 `npm run bump 4.0.1-test.1` 同步 npm、Tauri 和 Cargo 版本，然后提交并推送同名 `v4.0.1-test.1` 标签。正式版同理。CI 拒绝标签和应用版本不一致，测试版不能只改标签而保持应用版本不变。
2. GitHub Actions 需要仓库 Secret `TAURI_SIGNING_PRIVATE_KEY`（私钥文件内容）；如果私钥加密，还需 `TAURI_SIGNING_PRIVATE_KEY_PASSWORD`。公钥在 `src-tauri/tauri.conf.json` 的 `plugins.updater.pubkey` 中。
3. CI 构建各平台安装包及 `.sig`，统一生成 `latest.json`，使用嵌入的公钥验证上传产物，全部通过后才公开 Release。任何签名缺失或验证失败均保留草稿，不能进入自动更新渠道。

私钥不可提交到仓库，必须安全备份并在后续版本继续使用；随意更换公钥会使已安装版本无法验证新更新。macOS ad-hoc 签名与 updater 签名是两套独立机制。

本地只检查 UI 和应用 Bundle 时使用上面的 `createUpdaterArtifacts: false` 覆盖配置；正式发布始终启用更新产物和签名。

## CNB 产物镜像

- GitHub 为源码及安装包构建来源。合并到 `main` 后，`同步到 CNB` 工作流把源码同步到 `Mintimate/rime/oh-my-rime-cli`。
- `Release` 工作流成功公开版本后，再同步该版本标签，触发 CNB 的 `.cnb.yml`。
- CNB 拉取 21 个产物（安装包、4 个 CLI、7 个签名及更新清单），再次验签，改写清单下载地址，上传齐全后才公开 CNB Release。
- GitHub 仓库需要 `CNB_GIT_PASSWORD` Actions Secret。CNB 构建使用平台提供的 `CNB_TOKEN` 上传附件，不需要复制 updater 私钥。
- 同步采用普通推送，遇到 CNB 分支分叉或同名标签冲突会失败，不会强制覆盖历史。
- 附件上传后逐个通过下载接口验证 SHA-256，包括 `latest.json`；不依赖 Release 汇总接口的附件列表。短暂不可见会有限重试，校验失败仍保留草稿。
- 如果镜像脚本在发布后修复，应在 CNB 选择 **main → 手动运行（web_trigger）**：使用 main 的新脚本恢复 `tauri.conf.json` 中当前版本的镜像，不修改公开标签。直接重跑旧标签仍会使用旧脚本。
- 应用内更新目前从 GitHub 检查。CNB 镜像更新清单保留原签名，仅替换下载 URL。
