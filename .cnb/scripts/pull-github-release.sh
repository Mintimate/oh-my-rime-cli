#!/usr/bin/env bash
set -euo pipefail

: "${GITHUB_REPOSITORY_SLUG:?}"
: "${TARGET_CNB_REPO:?}"
: "${RELEASE_TAG:?}"
: "${RELEASE_ASSETS_DIR:?}"
: "${GITHUB_RELEASE_METADATA_FILE:?}"

release_tag="${RELEASE_TAG#refs/tags/}"
if [[ "${CNB_EVENT:-}" == web_trigger || "${CNB_EVENT:-}" == api_trigger ]]; then
  release_tag="v$(node -p 'require("./src-tauri/tauri.conf.json").version')"
fi
[[ "$release_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]] || exit 1
wait_seconds="${GITHUB_RELEASE_WAIT_SECONDS:-900}"
poll_seconds="${GITHUB_RELEASE_POLL_SECONDS:-15}"
[[ "$wait_seconds" =~ ^[0-9]+$ && "$poll_seconds" =~ ^[1-9][0-9]*$ ]] || exit 1
mkdir -p "$RELEASE_ASSETS_DIR" "$(dirname "$GITHUB_RELEASE_METADATA_FILE")"
manifest="$RELEASE_ASSETS_DIR/latest.json"
base_url="https://github.com/${GITHUB_REPOSITORY_SLUG}/releases/download/${release_tag}"
deadline=$((SECONDS + wait_seconds))
until curl --fail --location --silent --show-error --connect-timeout 20 --max-time 120 \
  "$base_url/latest.json" --output "$manifest.download"; do
  (( SECONDS < deadline )) || { echo "GitHub Release 未在规定时间内就绪" >&2; exit 1; }
  echo "等待 GitHub Release ${release_tag}"
  sleep "$poll_seconds"
done
mv "$manifest.download" "$manifest"
version="${release_tag#v}"
node .cnb/scripts/mirror-manifest.mjs assets "$version" |
  while IFS= read -r name; do
    [[ "$name" == latest.json ]] && continue
    echo "下载 $name"
    curl --fail --location --silent --show-error --retry 5 --retry-all-errors \
      --connect-timeout 30 --max-time 1800 "$base_url/$name" --output "$RELEASE_ASSETS_DIR/$name.download"
    mv "$RELEASE_ASSETS_DIR/$name.download" "$RELEASE_ASSETS_DIR/$name"
    test -s "$RELEASE_ASSETS_DIR/$name"
  done

# Validate signatures again before publishing the same packages on CNB.
verification_dir="$(mktemp -d)"
trap 'rm -rf "$verification_dir"' EXIT
jq -r '.plugins.updater.pubkey' src-tauri/tauri.conf.json | base64 --decode > "$verification_dir/public.key"
for signature in "$RELEASE_ASSETS_DIR"/*.sig; do
  base64 --decode "$signature" > "$verification_dir/package.minisig"
  minisign -Vm "${signature%.sig}" -p "$verification_dir/public.key" -x "$verification_dir/package.minisig"
done
node .cnb/scripts/mirror-manifest.mjs prepare "$version" "$TARGET_CNB_REPO" \
  "$RELEASE_ASSETS_DIR" "$GITHUB_RELEASE_METADATA_FILE"
