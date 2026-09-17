import { readFileSync } from "node:fs";

const tag = process.argv[2];
const config = JSON.parse(readFileSync("src-tauri/tauri.conf.json", "utf8"));
const pkg = JSON.parse(readFileSync("package.json", "utf8"));
const cargo = readFileSync("src-tauri/Cargo.toml", "utf8");
const rustVersion = cargo
  .split("[package]")[1]
  ?.split(/^\[/m)[0]
  ?.match(/^version = "([^"]+)"/m)?.[1];
if (
  tag !== `v${config.version}` ||
  config.version !== pkg.version ||
  config.version !== rustVersion
) {
  throw new Error(
    `发布标签及应用版本必须一致：tag=${tag}, Tauri=${config.version}, npm=${pkg.version}, Rust=${rustVersion}。请先运行 npm run bump <version>。`,
  );
}
console.log(`Release version verified: ${config.version}`);
