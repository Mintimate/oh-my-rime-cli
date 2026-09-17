#!/usr/bin/env bash

set -euo pipefail

: "${TARGET_CNB_REPO:?TARGET_CNB_REPO is required}"
: "${CNB_TOKEN:?CNB_TOKEN is required}"
: "${RELEASE_TAG:?RELEASE_TAG is required}"
: "${RELEASE_ASSETS_DIR:?RELEASE_ASSETS_DIR is required}"
: "${GITHUB_RELEASE_METADATA_FILE:?GITHUB_RELEASE_METADATA_FILE is required}"

release_tag="${RELEASE_TAG#refs/tags/}"
release_name="$(jq -r '.name // .tag_name' "${GITHUB_RELEASE_METADATA_FILE}")"
release_prerelease="$(jq -r '.prerelease // false' "${GITHUB_RELEASE_METADATA_FILE}")"
release_body="$(jq -r '.body // ""' "${GITHUB_RELEASE_METADATA_FILE}")"
if [[ "${release_prerelease}" == "true" ]]; then
  release_make_latest=false
else
  release_make_latest=true
fi

cnb_cli() {
  cnb "$@"
}

response_status() {
  jq -r '.status // 0' <<<"$1"
}

require_status() {
  local response="$1"
  shift
  local actual
  actual="$(response_status "$response")"

  for expected in "$@"; do
    if [[ "${actual}" == "${expected}" ]]; then
      return 0
    fi
  done

  echo "CNB API returned unexpected status ${actual}" >&2
  # Do not print API payloads: upload responses may include signed URLs.
  return 1
}

release_payload="$(jq -cn \
  --arg tag_name "${release_tag}" \
  --arg name "${release_name}" \
  --arg body "${release_body}" \
  --arg target_commitish "${release_tag}" \
  --arg make_latest "${release_make_latest}" \
  --argjson prerelease "${release_prerelease}" \
  '{
    tag_name: $tag_name,
    target_commitish: $target_commitish,
    name: $name,
    body: $body,
    draft: true,
    prerelease: $prerelease,
    make_latest: "false"
  }')"

release_response="$(
  cnb_cli releases get-release-by-tag \
    --repo "${TARGET_CNB_REPO}" \
    --tag "${release_tag}" \
    --verbose
)"

if [[ "$(response_status "${release_response}")" == "200" ]]; then
  release_id="$(jq -r '.data.id' <<<"${release_response}")"
  existing_assets="$(jq -c '.data.assets // []' <<<"${release_response}")"
  update_payload="$(jq -cn \
    --arg name "${release_name}" \
    --arg body "${release_body}" \
    --arg make_latest "${release_make_latest}" \
    --argjson prerelease "${release_prerelease}" \
    '{
      name: $name,
      body: $body,
      draft: true,
      prerelease: $prerelease,
      make_latest: "false"
    }')"
  update_response="$(
    cnb_cli releases patch-release \
      --repo "${TARGET_CNB_REPO}" \
      --release-id "${release_id}" \
      --data "${update_payload}" \
      --verbose
  )"
  require_status "${update_response}" 200
else
  require_status "${release_response}" 404
  create_response="$(
    cnb_cli releases post-release \
      --repo "${TARGET_CNB_REPO}" \
      --data "${release_payload}" \
      --verbose
  )"
  require_status "${create_response}" 201
  release_id="$(jq -r '.data.id' <<<"${create_response}")"
  existing_assets="[]"
fi

shopt -s nullglob
assets=("${RELEASE_ASSETS_DIR}"/*)
if (( ${#assets[@]} == 0 )); then
  echo "No GitHub Release assets found in ${RELEASE_ASSETS_DIR}" >&2
  exit 1
fi

for asset in "${assets[@]}"; do
  asset_name="$(basename "${asset}")"
  asset_size="$(wc -c <"${asset}" | tr -d '[:space:]')"
  upload_response="$(
    cnb_cli releases post-release-asset-upload-url \
      --repo "${TARGET_CNB_REPO}" \
      --release-id "${release_id}" \
      --asset-name "${asset_name}" \
      --size "${asset_size}" \
      --ttl 0 \
      --overwrite \
      --verbose
  )"
  require_status "${upload_response}" 201

  upload_url="$(jq -r '.data.upload_url' <<<"${upload_response}")"
  verify_url="$(jq -r '.data.verify_url' <<<"${upload_response}")"

  curl --fail --silent --show-error \
    --retry 3 \
    --request PUT \
    --upload-file "${asset}" \
    "${upload_url}"

  verify_path="$(node -e 'console.log(decodeURIComponent(new URL(process.argv[1]).pathname))' "${verify_url}")"
  verify_suffix="${verify_path#*/asset-upload-confirmation/}"
  upload_token="${verify_suffix%%/*}"
  asset_path="${verify_suffix#*/}"

  if [[ -z "${upload_token}" || -z "${asset_path}" || "${asset_path}" == "${verify_suffix}" ]]; then
    echo "Could not parse the CNB upload confirmation URL" >&2
    exit 1
  fi

  confirm_response="$(
    cnb_cli releases post-release-asset-upload-confirmation \
      --repo "${TARGET_CNB_REPO}" \
      --release-id "${release_id}" \
      --upload-token "${upload_token}" \
      --asset-path "${asset_path}" \
      --ttl 0 \
      --verbose
  )"
  require_status "${confirm_response}" 200
  echo "Uploaded ${asset_name} to CNB Release ${release_tag}"
done

final_response="$(
  cnb_cli releases get-release-by-tag \
    --repo "${TARGET_CNB_REPO}" \
    --tag "${release_tag}" \
    --verbose
)"
require_status "${final_response}" 200

for asset in "${assets[@]}"; do
  name="$(basename "$asset")"
  size="$(wc -c <"$asset" | tr -d '[:space:]')"
  jq -e --arg name "$name" --argjson size "$size" '
    [.data.assets[] | select(.name == $name and .size == $size)] | length == 1
  ' <<<"$final_response" > /dev/null || { echo "Missing or incomplete CNB asset: $name" >&2; exit 1; }
done
publish_payload="$(jq -cn --arg make_latest "$release_make_latest" --argjson prerelease "$release_prerelease" \
  '{draft: false, prerelease: $prerelease, make_latest: $make_latest}')"
publish_response="$(cnb_cli releases patch-release --repo "$TARGET_CNB_REPO" \
  --release-id "$release_id" --data "$publish_payload" --verbose)"
require_status "$publish_response" 200

echo "CNB Release ${release_tag} is ready with ${#assets[@]} assets"
