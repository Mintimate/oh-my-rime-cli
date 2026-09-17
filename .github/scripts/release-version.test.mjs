import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import {
  cpSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import test from "node:test";

test("同步测试版本，保留第三方依赖版本，并拒绝错误标签", (t) => {
  const directory = mkdtempSync(join(tmpdir(), "omr-release-version-"));
  t.after(() => rmSync(directory, { recursive: true, force: true }));
  for (const path of [
    "package.json",
    "package-lock.json",
    "src-tauri/Cargo.toml",
    "src-tauri/Cargo.lock",
    "src-tauri/tauri.conf.json",
  ]) {
    cpSync(path, join(directory, path), { recursive: true });
  }
  const lockPath = join(directory, "src-tauri/Cargo.lock");
  const original = readFileSync(lockPath, "utf8");
  const dependency =
    '\n[[package]]\nname = "same-version-fixture"\nversion = "2.1.0"\n';
  writeFileSync(lockPath, original + dependency);
  const version = "2.1.1-test.10";
  execFileSync(
    process.execPath,
    [resolve("script/bump-version.mjs"), version],
    { cwd: directory },
  );
  const updated = readFileSync(lockPath, "utf8");
  assert.ok(updated.endsWith(dependency));
  assert.match(updated, /name = "oh-my-rime"\nversion = "2.1.1-test.10"/);
  for (const path of [
    "package.json",
    "package-lock.json",
    "src-tauri/tauri.conf.json",
  ]) {
    assert.equal(
      JSON.parse(readFileSync(join(directory, path), "utf8")).version,
      version,
    );
  }
  assert.equal(
    JSON.parse(readFileSync(join(directory, "package-lock.json"), "utf8"))
      .packages[""].version,
    version,
  );
  const checker = resolve(".github/scripts/check-release-version.mjs");
  execFileSync(process.execPath, [checker, `v${version}`], { cwd: directory });
  assert.throws(() =>
    execFileSync(process.execPath, [checker, "v2.1.1-test.9"], {
      cwd: directory,
      stdio: "pipe",
    }),
  );
});
