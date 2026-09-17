import { readFileSync, statSync, writeFileSync } from "node:fs";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import {
  TARGETS,
  expectedUpdaterAssetNames,
} from "../../.github/scripts/generate-updater-manifest.mjs";

export function expectedReleaseAssets(version) {
  return [
    ...expectedUpdaterAssetNames(version),
    `Oh-My-Rime_${version}_macOS_arm64.dmg`,
    `Oh-My-Rime_${version}_macOS_x64.dmg`,
    "cli-macos-arm64",
    "cli-macos-x64",
    "cli-linux-x64",
    "cli-windows-x64.exe",
    "latest.json",
  ];
}

export function mirrorManifest(manifest, version, repo, directory) {
  if (manifest.version !== version) throw new Error("更新清单版本与标签不一致");
  for (const name of expectedReleaseAssets(version)) {
    if (
      !statSync(join(directory, name)).isFile() ||
      statSync(join(directory, name)).size === 0
    ) {
      throw new Error(`产物缺失或为空：${name}`);
    }
  }
  const platforms = {};
  for (const target of TARGETS) {
    const name = `Oh-My-Rime_${version}_${target.os}_${target.arch}${target.suffix}`;
    const signature = readFileSync(join(directory, `${name}.sig`), "utf8");
    if (!signature.trim()) throw new Error(`签名为空：${name}`);
    for (const key of target.keys) {
      const entry = manifest.platforms?.[key];
      if (
        entry?.url !==
          `https://github.com/Mintimate/oh-my-rime-cli/releases/download/v${version}/${name}` ||
        entry.signature.trim() !== signature.trim()
      ) {
        throw new Error(`平台产物或签名不匹配：${key}`);
      }
      platforms[key] = {
        ...entry,
        url: `https://cnb.cool/${repo}/-/releases/download/v${version}/${encodeURIComponent(name)}`,
      };
    }
  }
  if (Object.keys(manifest.platforms).length !== Object.keys(platforms).length)
    throw new Error("存在未知更新平台");
  return { ...manifest, platforms };
}

if (
  process.argv[1] &&
  fileURLToPath(import.meta.url) === resolve(process.argv[1])
) {
  const [command, version, repo, directory, metadataPath] =
    process.argv.slice(2);
  if (command === "assets") {
    console.log(expectedReleaseAssets(version).join("\n"));
  } else if (command === "prepare") {
    const path = join(directory, "latest.json");
    const original = JSON.parse(readFileSync(path, "utf8"));
    const mirrored = mirrorManifest(original, version, repo, directory);
    writeFileSync(
      metadataPath,
      JSON.stringify(
        {
          tag_name: `v${version}`,
          name: `Oh My Rime v${version}`,
          body: original.notes ?? "",
          prerelease: version.includes("-"),
        },
        null,
        2,
      ),
    );
    writeFileSync(path, `${JSON.stringify(mirrored, null, 2)}\n`);
    console.log(
      `已准备 ${expectedReleaseAssets(version).length} 个 CNB 发布产物`,
    );
  } else
    throw new Error(
      "用法：mirror-manifest.mjs assets <version> 或 prepare <version> <cnb-repo> <assets-dir> <metadata-path>",
    );
}
