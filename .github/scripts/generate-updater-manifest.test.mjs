import assert from "node:assert/strict";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import {
  expectedUpdaterAssetNames,
  generateUpdaterManifest,
} from "./generate-updater-manifest.mjs";

function createFixture(version = "1.2.3") {
  const directory = mkdtempSync(join(tmpdir(), "omr-updater-manifest-"));
  const names = expectedUpdaterAssetNames(version);
  const assets = names.map((name, index) => ({
    id: index + 1,
    name,
    label: name,
    state: "uploaded",
    size: 100,
    browser_download_url: `https://github.com/Mintimate/oh-my-rime-cli/releases/download/v${version}/${name}`,
  }));

  for (const name of names.filter((name) => name.endsWith(".sig"))) {
    writeFileSync(join(directory, name), `signature:${name}\n`);
  }

  return {
    directory,
    release: {
      tag_name: `v${version}`,
      body: "release notes",
      assets,
    },
  };
}

test("生成完整的跨平台 updater 清单", (t) => {
  const { directory, release } = createFixture();
  t.after(() => rmSync(directory, { recursive: true, force: true }));

  const manifest = generateUpdaterManifest({
    version: "1.2.3",
    tag: "v1.2.3",
    release,
    signatureDirectory: directory,
    pubDate: "2026-09-02T00:00:00.000Z",
  });

  assert.equal(manifest.version, "1.2.3");
  assert.equal(manifest.notes, "release notes");
  assert.equal(Object.keys(manifest.platforms).length, 11);
  assert.deepEqual(
    manifest.platforms["windows-x86_64"],
    manifest.platforms["windows-x86_64-nsis"],
  );
  assert.deepEqual(
    manifest.platforms["linux-x86_64"],
    manifest.platforms["linux-x86_64-appimage"],
  );
  assert.equal(
    manifest.platforms["darwin-aarch64"].url,
    "https://github.com/Mintimate/oh-my-rime-cli/releases/download/v1.2.3/Oh-My-Rime_1.2.3_macOS_arm64.app.tar.gz",
  );
});

test("缺少安装资产时拒绝生成清单", (t) => {
  const { directory, release } = createFixture();
  t.after(() => rmSync(directory, { recursive: true, force: true }));
  release.assets = release.assets.filter(
    (asset) => !asset.name.endsWith("Linux_x64.rpm"),
  );

  assert.throws(
    () =>
      generateUpdaterManifest({
        version: "1.2.3",
        tag: "v1.2.3",
        release,
        signatureDirectory: directory,
      }),
    /Linux_x64\.rpm/,
  );
});

test("版本与标签不一致时拒绝生成清单", (t) => {
  const { directory, release } = createFixture();
  t.after(() => rmSync(directory, { recursive: true, force: true }));

  assert.throws(
    () =>
      generateUpdaterManifest({
        version: "1.2.3",
        tag: "v1.2.4",
        release,
        signatureDirectory: directory,
      }),
    /版本 1\.2\.3 与标签 v1\.2\.4 不一致/,
  );
});

for (const [label, mutate] of [
  [
    "缺少签名",
    (r) => {
      r.assets = r.assets.filter(
        (a) => !a.name.endsWith("Windows_x64.msi.sig"),
      );
    },
  ],
  [
    "重复产物",
    (r) => {
      r.assets.push(r.assets[0]);
    },
  ],
  [
    "空产物",
    (r) => {
      r.assets[0].size = 0;
    },
  ],
  [
    "错误下载地址",
    (r) => {
      r.assets[0].browser_download_url = "https://example.com/wrong";
    },
  ],
]) {
  test(`${label}时拒绝发布`, (t) => {
    const { directory, release } = createFixture();
    t.after(() => rmSync(directory, { recursive: true, force: true }));
    mutate(release);
    assert.throws(() =>
      generateUpdaterManifest({
        version: "1.2.3",
        tag: "v1.2.3",
        release,
        signatureDirectory: directory,
      }),
    );
  });
}

test("保留测试版 SemVer", (t) => {
  const version = "2.1.1-test.4";
  const { directory, release } = createFixture(version);
  t.after(() => rmSync(directory, { recursive: true, force: true }));
  assert.equal(
    generateUpdaterManifest({
      version,
      tag: `v${version}`,
      release,
      signatureDirectory: directory,
    }).version,
    version,
  );
});
