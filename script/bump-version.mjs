import { readFileSync, writeFileSync } from "node:fs";

const version = process.argv[2];
const match = version?.match(
  /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$/,
);
if (!match || match[4]?.split(".").some((part) => /^0\d+$/.test(part))) {
  throw new Error(
    "请输入有效版本号，例如 npm run bump 2.1.1-test.1（不带 v 前缀）",
  );
}
const files = new Map();
for (const path of [
  "package.json",
  "package-lock.json",
  "src-tauri/tauri.conf.json",
]) {
  const json = JSON.parse(readFileSync(path, "utf8"));
  json.version = version;
  if (path === "package-lock.json") json.packages[""].version = version;
  files.set(path, `${JSON.stringify(json, null, 2)}\n`);
}
for (const [path, pattern] of [
  ["src-tauri/Cargo.toml", /(\[package\][\s\S]*?^version = ")[^"]+("$)/m],
  [
    "src-tauri/Cargo.lock",
    /(\[\[package\]\]\nname = "oh-my-rime"\nversion = ")[^"]+("$)/m,
  ],
]) {
  const content = readFileSync(path, "utf8");
  if (!pattern.test(content)) throw new Error(`无法定位 ${path} 中的应用版本`);
  files.set(path, content.replace(pattern, `$1${version}$2`));
}
for (const [path, content] of files) writeFileSync(path, content);
console.log(
  `应用版本已同步为 ${version}。提交后推送 v${version} 标签即可发布。`,
);
