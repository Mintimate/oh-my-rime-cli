# oh-my-rime

A local desktop manager for updating and maintaining Rime configurations across macOS, Linux, and Windows.

The desktop application is built with **Tauri 2 + React + TypeScript + Rust**. The GUI and the standalone Rust CLI share the same update core, including platform detection, streaming downloads, backups, safe extraction, and rollback.

## Features

- Update the Mintimate Rime scheme
- Update the Wanxiang gram model
- Update the Wanxiang dictionary package
- Install a custom `.zip` or `.gram` resource URL
- Preserve user `*.custom.yaml` files
- Create a backup before each update
- Restore the previous configuration after a failed update
- macOS, Linux, and Windows target-directory presets
- Local-only operation without telemetry or a backend service

## Requirements

- Node.js 20+ and npm
- Rust stable and Cargo
- Tauri 2 platform build dependencies
- Go is no longer required for the desktop application

## Development

```bash
npm install
npm run dev
npm run build:web
npm run build
cargo fmt --manifest-path src-tauri/Cargo.toml -- --check
cargo test --manifest-path src-tauri/Cargo.toml
```

The Rust update core is located at `src-tauri/src/rime_core.rs`.

Run the standalone CLI with:

```bash
cargo run --manifest-path src-tauri/Cargo.toml --features cli --bin oh-my-rime-cli
```

## Build outputs

On macOS, you can validate the application bundle without creating a DMG:

```bash
npx tauri build --bundles app
```

DMG, Windows installers, and Linux packages are produced by the platform-specific release workflow.

The desktop installer contains the GUI executable only. The standalone CLI is built separately with the `cli` Cargo feature and published under platform-specific names.

macOS builds use ad-hoc signing by default, both locally and in CI, without requiring a signing key or environment variable. They are not Apple-notarized. After copying the app to Applications and attempting to open it, confirm its source and use **System Settings → Privacy & Security → Open Anyway** if Apple cannot verify it. Enter your Mac login password or use Touch ID when prompted. The exact behavior depends on your macOS version and security policy. See [Apple's instructions](https://support.apple.com/en-us/102445). If macOS reports the app as damaged, redownload it and verify its signature; the alert alone does not identify whether the cause is download corruption or a signing problem.

Tauri updater keys verify update artifacts and do not replace macOS code signing or Apple notarization. See the [GUI build guide](GUI_README.md#构建) for signature verification commands.

After an update, use the reload/redeploy action provided by your Rime input method so the new configuration takes effect.

## License

MIT License.
