#!/usr/bin/env bash
set -euo pipefail

# Verify the actual uploaded packages against the public key embedded in the app.
# Requires gh, jq, base64 and minisign. Download through the API while still draft.
manifest="$1"
release_json="$2"
config="$3"
verification_dir="$(mktemp -d)"
trap 'rm -rf "$verification_dir"' EXIT
jq -r '.plugins.updater.pubkey' "$config" | base64 --decode > "$verification_dir/public.key"

jq -c '[.platforms[]] | unique_by(.url)[]' "$manifest" |
  while IFS= read -r entry; do
    url="$(jq -r '.url' <<<"$entry")"
    api_url="$(jq -er --arg url "$url" '
      [.assets[] | select(.browser_download_url == $url)]
      | if length == 1 then .[0].url else error("Expected one update asset") end
    ' "$release_json")"
    gh api -H 'Accept: application/octet-stream' "$api_url" > "$verification_dir/package"
    jq -r '.signature' <<<"$entry" | base64 --decode > "$verification_dir/package.minisig"
    minisign -Vm "$verification_dir/package" -p "$verification_dir/public.key" -x "$verification_dir/package.minisig"
  done
