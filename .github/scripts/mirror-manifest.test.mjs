import assert from "node:assert/strict";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import { generateUpdaterManifest } from "./generate-updater-manifest.mjs";
import {
  expectedReleaseAssets,
  mirrorManifest,
} from "../../.cnb/scripts/mirror-manifest.mjs";

function fixture(t) {
  const directory = mkdtempSync(join(tmpdir(), "omr-mirror-"));
  t.after(() => rmSync(directory, { recursive: true, force: true }));
  const version = "4.0.0";
  const names = expectedReleaseAssets(version);
  for (const name of names)
    writeFileSync(join(directory, name), `content:${name}`);
  const release = {
    tag_name: `v${version}`,
    body: "Release notes",
    assets: names.map((name) => ({
      name,
      size: 1,
      state: "uploaded",
      browser_download_url: `https://github.com/Mintimate/oh-my-rime-cli/releases/download/v${version}/${name}`,
    })),
  };
  const manifest = generateUpdaterManifest({
    version,
    tag: `v${version}`,
    release,
    signatureDirectory: directory,
  });
  return { directory, version, manifest };
}

test("镜像全部 21 个产物且只替换下载地址", (t) => {
  const { directory, version, manifest } = fixture(t);
  const mirrored = mirrorManifest(
    manifest,
    version,
    "Mintimate/rime/oh-my-rime-cli",
    directory,
  );
  assert.equal(expectedReleaseAssets(version).length, 21);
  assert.equal(Object.keys(mirrored.platforms).length, 11);
  assert.equal(mirrored.notes, manifest.notes);
  for (const [key, entry] of Object.entries(mirrored.platforms)) {
    assert.equal(entry.signature, manifest.platforms[key].signature);
    assert.ok(
      entry.url.startsWith(
        "https://cnb.cool/Mintimate/rime/oh-my-rime-cli/-/releases/download/v4.0.0/",
      ),
    );
  }
});

test("缺少 CLI 时拒绝镜像", (t) => {
  const { directory, version, manifest } = fixture(t);
  rmSync(join(directory, "cli-linux-x64"));
  assert.throws(() => mirrorManifest(manifest, version, "repo", directory));
});

test("更新包 URL 或签名不一致时拒绝镜像", (t) => {
  const { directory, version, manifest } = fixture(t);
  const key = "darwin-aarch64";
  const invalidUrl = structuredClone(manifest);
  invalidUrl.platforms[key].url = "https://example.com/wrong";
  assert.throws(
    () => mirrorManifest(invalidUrl, version, "repo", directory),
    /不匹配/,
  );
  manifest.platforms[key].signature = "invalid signature";
  assert.throws(
    () => mirrorManifest(manifest, version, "repo", directory),
    /不匹配/,
  );
});
