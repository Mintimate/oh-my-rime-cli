# Oh My Rime · OMR

[简体中文](README.md) · [Desktop and release guide](GUI_README.md) · [GitHub downloads](https://github.com/Mintimate/oh-my-rime-cli/releases) · [CNB mirror](https://cnb.cool/Mintimate/rime/oh-my-rime-cli/-/releases)

A desktop application and standalone CLI for maintaining [Oh-my-rime](https://github.com/Mintimate/oh-my-rime) configurations on macOS, Windows, and Linux.

> Starting with v4.0.0, OMR uses Tauri 2, React, TypeScript, and Rust. Older Go/Wails versions need one manual installation before they can use the application updater.

## Preview

Native-resolution macOS screenshots of the desktop application (2240 × 1520, lossless PNG). The preview follows the document's color scheme; the app offers Light, System, and Dark modes.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="assets/img/usingAppDark.png">
  <source media="(prefers-color-scheme: light)" srcset="assets/img/usingApp.png">
  <img alt="OMR desktop: configuration updates, target directories, progress and theme controls" src="assets/img/usingApp.png">
</picture>

[Light theme](assets/img/usingApp.png) · [Dark theme](assets/img/usingAppDark.png)

## Features

- Update the Mintimate scheme, Wanxiang model and dictionary, or a custom `.zip` or `.gram` URL.
- Detect platform-specific Rime directories or select a directory manually.
- Download to a temporary directory, back up before changing files, and attempt rollback on failure. Scheme updates preserve `*.custom.yaml` files.
- View progress and copy logs from the current session.
- Choose Light, System, or Dark appearance with saved preferences.
- Check for application updates, read release notes, download, verify signatures, install and restart. Prerelease updates are optional.
- Share the Rust update core between the GUI and CLI, without uploading user configurations or collecting telemetry.

## Download and install

Choose your platform and architecture on the [Releases page](https://github.com/Mintimate/oh-my-rime-cli/releases). Rust packages use the following names; test releases are marked **Pre-release**.

| Platform | Desktop installer | Standalone CLI |
| --- | --- | --- |
| macOS Apple Silicon | `Oh-My-Rime_<version>_macOS_arm64.dmg` | `cli-macos-arm64` |
| macOS Intel | `Oh-My-Rime_<version>_macOS_x64.dmg` | `cli-macos-x64` |
| Windows x64 | `.msi` or `-setup.exe` | `cli-windows-x64.exe` |
| Linux x64 | `.AppImage`, `.deb` or `.rpm` | `cli-linux-x64` |

The `.app.tar.gz`, `.sig` and `latest.json` files serve the application updater. Use the DMG for manual macOS installation.

### First launch on macOS

macOS bundles use ad-hoc signing and are not Apple-notarized.

1. Copy `Oh My Rime.app` into Applications and attempt to open it.
2. If Apple cannot verify the developer, confirm the download source and use **System Settings → Privacy & Security → Open Anyway**. Enter your Mac login password or use Touch ID when prompted.

The exact behavior depends on your macOS version and security policy. If the app is reported as damaged, redownload it and verify its signature. See [Apple's instructions](https://support.apple.com/en-us/102445) and the [signature verification guide](GUI_README.md#构建).

Windows installers are currently unsigned. Download them from this project's Releases page and verify their source before running them.

## Update Rime configurations

1. Open **方案更新** and select or enter a Rime configuration directory.
2. Choose a scheme, model or dictionary update, or enter a custom resource URL.
3. Follow the progress and open **运行日志** for details.
4. After success, use your input method's reload/redeploy action to apply the changes.

Directory detection supports Weasel on Windows (registry, then `%APPDATA%\Rime`), Squirrel and Fcitx5 on macOS, and iBus, Fcitx5 and Fcitx5 Flatpak on Linux.

## Update the OMR application

The app checks for updates at startup. Click **检查应用更新** in the sidebar to check manually. When a new version is available, read its notes and click **下载并安装** to download, verify and install it. The app restarts after installation. Installation is blocked while a Rime configuration update is running.

![Application updater with current version, prerelease preference and check result](assets/img/appUpdate.png)

This screenshot shows the real state with no available update; it does not simulate a release or download progress.

- Stable installations default to stable releases; prerelease installations default to including prereleases on first use. Toggle **接收测试版** to change this preference.
- Only releases with updater manifests are considered, excluding older Go packages.
- Network failures do not block other functionality; retry the check later.
- Older versions without updater support need one manual installation. The standalone CLI must still be updated manually.
- On macOS, run the app from Applications, not from a read-only DMG.

## Development

Requirements: Node.js 20.19+ (or 22.12+), stable Rust, and the platform dependencies for Tauri 2. Frontend code is in `src/`; the shared update core is in `src-tauri/src/rime_core.rs`.

```bash
npm ci
npm run dev
```

Build locally without an updater signing key:

```bash
npm run build -- --config '{"bundle":{"createUpdaterArtifacts":false}}'
```

Run the CLI:

```bash
cargo run --manifest-path src-tauri/Cargo.toml --features cli --bin oh-my-rime-cli
```

Validation:

```bash
npm run typecheck
npm run build:web
npm run test:release
cargo fmt --manifest-path src-tauri/Cargo.toml -- --check
cargo test --manifest-path src-tauri/Cargo.toml --features cli
```

## Publishing

Run `npm run bump <version>` to synchronize the npm, Tauri and Cargo versions. Commit the changes and push a matching `v<version>` tag, such as `4.0.1-test.1` / `v4.0.1-test.1`. Changing only the Git tag is not sufficient for application updates.

GitHub Actions builds the installers and CLI, signs updater packages, generates `latest.json`, and verifies the uploaded packages before publishing. Configure the `TAURI_SIGNING_PRIVATE_KEY` repository Secret; encrypted keys also require their password. Never commit the private key, and retain the same signing key for future updates.

After a successful GitHub release, Actions mirrors its tag to CNB. The `.cnb.yml` pipeline downloads and verifies the same installers, CLI files and signatures before publishing them on CNB. In-app updates currently use GitHub; CNB provides mirror downloads.

See the [signed release guide](GUI_README.md#发布签名更新) for details.

## Contributing and license

MIT License. Issues and pull requests are welcome.

## Support

- [Mintimate's Blog](https://www.mintimate.cn)
- [Support on Afdian](https://afdian.net/a/mintimate)
- [Bilibili](https://space.bilibili.com/355567627)
- [YouTube](https://www.youtube.com/channel/UCI7LLdUGNzkcKOE7grAqCoA)
