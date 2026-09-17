#!/usr/bin/env node

import { readFileSync, writeFileSync } from "node:fs";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

export const TARGETS = [
  {
    keys: ["darwin-aarch64", "darwin-aarch64-app"],
    os: "macOS",
    arch: "arm64",
    suffix: ".app.tar.gz",
  },
  {
    keys: ["darwin-x86_64", "darwin-x86_64-app"],
    os: "macOS",
    arch: "x64",
    suffix: ".app.tar.gz",
  },
  {
    keys: ["windows-x86_64", "windows-x86_64-nsis"],
    os: "Windows",
    arch: "x64",
    suffix: "-setup.exe",
  },
  {
    keys: ["windows-x86_64-msi"],
    os: "Windows",
    arch: "x64",
    suffix: ".msi",
  },
  {
    keys: ["linux-x86_64", "linux-x86_64-appimage"],
    os: "Linux",
    arch: "x64",
    suffix: ".AppImage",
  },
  {
    keys: ["linux-x86_64-deb"],
    os: "Linux",
    arch: "x64",
    suffix: ".deb",
  },
  {
    keys: ["linux-x86_64-rpm"],
    os: "Linux",
    arch: "x64",
    suffix: ".rpm",
  },
];

function assetName(version, target, signature = false) {
  return `Oh-My-Rime_${version}_${target.os}_${target.arch}${target.suffix}${signature ? ".sig" : ""}`;
}

function findUniqueAsset(release, expectedName) {
  const matches = release.assets.filter(
    (asset) => asset.name === expectedName || asset.label === expectedName,
  );

  if (matches.length !== 1) {
    throw new Error(
      `Release 中应当恰好存在一个 ${expectedName}，实际找到 ${matches.length} 个`,
    );
  }

  const [asset] = matches;
  if (asset.state !== "uploaded" || !(asset.size > 0)) {
    throw new Error(`Release 资产 ${expectedName} 尚未完整上传`);
  }
  const expectedUrl = `https://github.com/Mintimate/oh-my-rime-cli/releases/download/${release.tag_name}/${expectedName}`;
  if (asset.browser_download_url !== expectedUrl) {
    throw new Error(`Release 资产 ${expectedName} 下载地址不匹配`);
  }
  return asset;
}

export function expectedUpdaterAssetNames(version) {
  return TARGETS.flatMap((target) => [
    assetName(version, target),
    assetName(version, target, true),
  ]);
}

export function generateUpdaterManifest({
  version,
  tag,
  release,
  signatureDirectory,
  pubDate = new Date().toISOString(),
}) {
  if (tag !== `v${version}`) {
    throw new Error(`版本 ${version} 与标签 ${tag} 不一致`);
  }
  if (release.tag_name !== tag) {
    throw new Error(`Release 标签 ${release.tag_name} 与预期 ${tag} 不一致`);
  }
  if (!Array.isArray(release.assets)) {
    throw new Error("Release 数据缺少 assets 数组");
  }

  const platforms = {};
  for (const target of TARGETS) {
    const bundleName = assetName(version, target);
    const signatureName = assetName(version, target, true);
    const bundle = findUniqueAsset(release, bundleName);
    findUniqueAsset(release, signatureName);

    const signature = readFileSync(
      join(signatureDirectory, signatureName),
      "utf8",
    );
    if (!signature.trim()) {
      throw new Error(`签名文件 ${signatureName} 为空`);
    }

    const entry = {
      signature,
      url: bundle.browser_download_url,
    };
    for (const key of target.keys) {
      platforms[key] = entry;
    }
  }

  return {
    version,
    notes: release.body ?? "",
    pub_date: pubDate,
    platforms,
  };
}

function main() {
  const [version, tag, releasePath, signatureDirectory, outputPath] =
    process.argv.slice(2);
  if (!version || !tag || !releasePath || !signatureDirectory || !outputPath) {
    throw new Error(
      "用法：generate-updater-manifest.mjs <version> <tag> <release-json> <signature-dir> <output>",
    );
  }

  const release = JSON.parse(readFileSync(releasePath, "utf8"));
  const manifest = generateUpdaterManifest({
    version,
    tag,
    release,
    signatureDirectory,
  });
  writeFileSync(outputPath, `${JSON.stringify(manifest, null, 2)}\n`);
}

if (
  process.argv[1] &&
  fileURLToPath(import.meta.url) === resolve(process.argv[1])
) {
  main();
}
