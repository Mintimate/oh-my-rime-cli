import assert from "node:assert/strict";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { tmpdir } from "node:os";
import test from "node:test";
import { expectedReleaseAssets } from "../../.cnb/scripts/mirror-manifest.mjs";
import {
  verifyAsset,
  verifyReleaseAssets,
} from "../../.cnb/scripts/verify-release-assets.mjs";

function fixture(t) {
  const directory = mkdtempSync(join(tmpdir(), "omr-cnb-verify-"));
  t.after(() => rmSync(directory, { recursive: true, force: true }));
  const contents = new Map(
    expectedReleaseAssets("4.0.0").map((name) => [
      name,
      Buffer.from(`文件:${name}`),
    ]),
  );
  for (const [name, data] of contents)
    writeFileSync(join(directory, name), data);
  return {
    directory,
    contents,
    repo: "Mintimate/rime/oh-my-rime-cli",
    tag: "v4.0.0",
    token: "fixture-token",
  };
}

test("不依赖汇总列表，直接验证全部 21 个附件，包括 latest.json", async (t) => {
  const f = fixture(t);
  const verified = [];
  await verifyReleaseAssets({
    ...f,
    onVerified: (name) => verified.push(name),
    fetchImpl: async (url, options) => {
      const u = new URL(url);
      const name = decodeURIComponent(u.pathname.split("/").at(-1));
      if (u.host === "api.cnb.cool") {
        assert.ok(u.pathname.includes("/releases/download/v4.0.0/"));
        assert.equal(options.headers.Authorization, "Bearer fixture-token");
        assert.equal(options.redirect, "manual");
        return new Response(null, {
          status: 302,
          headers: {
            location: `https://assets.example/${encodeURIComponent(name)}`,
          },
        });
      }
      assert.equal(u.host, "assets.example");
      assert.equal(options.headers, undefined, "不得转发令牌给对象存储");
      return new Response(f.contents.get(name));
    },
  });
  assert.equal(verified.length, 21);
  assert.ok(verified.includes("latest.json"));
});

test("相同大小但被修改的附件也必须失败", async (t) => {
  const f = fixture(t);
  const name = "latest.json";
  const tampered = Buffer.from(f.contents.get(name));
  tampered[0] ^= 1;
  await assert.rejects(
    verifyAsset({ ...f, name, fetchImpl: async () => new Response(tampered) }),
    /内容不一致/,
  );
});

test("上传完成后的短暂不可见会重试", async (t) => {
  const f = fixture(t);
  let failures = 0;
  let waits = 0;
  await verifyReleaseAssets({
    ...f,
    wait: async () => waits++,
    fetchImpl: async (url) => {
      const name = decodeURIComponent(new URL(url).pathname.split("/").at(-1));
      if (name === "latest.json" && failures++ === 0)
        return new Response(null, { status: 404 });
      return new Response(f.contents.get(name));
    },
  });
  assert.equal(waits, 1);
});

test("持续缺失的附件不会被当成发布成功", async (t) => {
  const f = fixture(t);
  let calls = 0;
  await assert.rejects(
    verifyReleaseAssets({
      ...f,
      wait: async () => {},
      fetchImpl: async () => {
        calls++;
        return new Response(null, { status: 404 });
      },
    }),
    /HTTP 404/,
  );
  assert.equal(calls, 3);
});

test("拒绝把下载重定向到 HTTP", async (t) => {
  const f = fixture(t);
  await assert.rejects(
    verifyAsset({
      ...f,
      name: "latest.json",
      fetchImpl: async () =>
        new Response(null, {
          status: 302,
          headers: { location: "http://assets.example/file" },
        }),
    }),
    /必须使用 HTTPS/,
  );
});
