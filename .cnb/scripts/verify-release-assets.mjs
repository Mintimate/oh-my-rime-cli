import { createHash } from "node:crypto";
import { createReadStream } from "node:fs";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { expectedReleaseAssets } from "./mirror-manifest.mjs";

async function digest(stream) {
  const hash = createHash("sha256");
  let size = 0;
  for await (const chunk of stream) {
    size += chunk.length;
    hash.update(chunk);
  }
  return { size, hash: hash.digest("hex") };
}

export async function verifyAsset({
  repo,
  tag,
  name,
  directory,
  token,
  fetchImpl = fetch,
}) {
  const expected = await digest(createReadStream(join(directory, name)));
  if (expected.size === 0) throw new Error(`本地附件为空：${name}`);
  const url = `https://api.cnb.cool/${repo}/-/releases/download/${encodeURIComponent(tag)}/${encodeURIComponent(name)}`;
  let response = await fetchImpl(url, {
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: "application/octet-stream",
    },
    redirect: "manual",
    signal: AbortSignal.timeout(30_000),
  });
  if ([301, 302, 303, 307, 308].includes(response.status)) {
    const location = response.headers.get("location");
    if (!location) throw new Error(`附件下载重定向缺少地址：${name}`);
    const destination = new URL(location, url);
    if (destination.protocol !== "https:")
      throw new Error(`附件下载重定向必须使用 HTTPS：${name}`);
    await response.body?.cancel();
    // The storage URL is already signed. Never forward the CNB token to it.
    response = await fetchImpl(destination.href, {
      signal: AbortSignal.timeout(180_000),
    });
  }
  if (response.status !== 200 || !response.body) {
    await response.body?.cancel();
    throw new Error(`无法读取 CNB 附件 ${name}：HTTP ${response.status}`);
  }
  const actual = await digest(response.body);
  if (actual.size !== expected.size || actual.hash !== expected.hash) {
    throw new Error(
      `CNB 附件内容不一致：${name}（期望 ${expected.size} 字节，实际 ${actual.size} 字节）`,
    );
  }
}

export async function verifyReleaseAssets({
  repo,
  tag,
  directory,
  token,
  fetchImpl = fetch,
  wait = (ms) => new Promise((resolve) => setTimeout(resolve, ms)),
  onVerified = () => {},
}) {
  if (
    !/^[\w.-]+(?:\/[\w.-]+)+$/.test(repo) ||
    !/^v\d+\.\d+\.\d+(?:-[\w.-]+)?$/.test(tag) ||
    !token
  ) {
    throw new Error("缺少或无效的仓库、标签或 CNB_TOKEN");
  }
  for (const name of expectedReleaseAssets(tag.slice(1))) {
    for (let attempt = 0; ; attempt++) {
      try {
        await verifyAsset({ repo, tag, name, directory, token, fetchImpl });
        onVerified(name);
        break;
      } catch (error) {
        if (attempt === 2) throw error;
        // Upload confirmation can precede download visibility by a short delay.
        await wait(1000 * (attempt + 1));
      }
    }
  }
}

if (
  process.argv[1] &&
  fileURLToPath(import.meta.url) === resolve(process.argv[1])
) {
  const [repo, tag, directory] = process.argv.slice(2);
  try {
    await verifyReleaseAssets({
      repo,
      tag,
      directory,
      token: process.env.CNB_TOKEN,
      onVerified: (name) => console.log(`Verified CNB asset SHA-256: ${name}`),
    });
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}
