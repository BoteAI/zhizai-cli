#!/usr/bin/env bash
# publish-npm-packages.sh — publish platform optional packages then @zhizai/cli
#
# Expected layout:
#   dist/ contains release archives named zhizai-cli_<ver>_<goos>_<goarch>.{tar.gz,zip}
#   npm/platforms/<key>/package.json templates exist in repo
#
# Usage (CI):
#   VERSION=0.0.11 DIST_DIR=dist bash scripts/publish-npm-packages.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

VERSION="${VERSION:-$(node -p "require('./package.json').version")}"
DIST_DIR="${DIST_DIR:-dist}"
NPM_ACCESS="${NPM_ACCESS:-public}"

if [[ ! -d "$DIST_DIR" ]]; then
  echo "error: dist dir not found: $DIST_DIR" >&2
  exit 1
fi

node scripts/sync-npm-versions.js "$VERSION"

while IFS= read -r key; do
  [[ -z "$key" ]] && continue

  meta="$(node -e "
    const { PLATFORMS, releaseAssetName } = require('./scripts/platform.js');
    const spec = PLATFORMS.find(p => p.key === process.argv[1]);
    if (!spec) process.exit(2);
    console.log([spec.pkg, releaseAssetName(process.argv[2], spec), spec.binary, spec.archiveExt].join('\t'));
  " "$key" "$VERSION")"
  IFS=$'\t' read -r pkg_name asset_name binary_name archive_ext <<<"$meta"

  asset_path="${DIST_DIR}/${asset_name}"
  if [[ ! -f "$asset_path" ]]; then
    echo "error: missing release asset: $asset_path" >&2
    exit 1
  fi

  pkg_dir="npm/platforms/${key}"
  bin_dir="${pkg_dir}/bin"
  rm -rf "$bin_dir"
  mkdir -p "$bin_dir"

  staging="$(mktemp -d)"
  if [[ "$archive_ext" == ".zip" ]]; then
    unzip -qo "$asset_path" -d "$staging"
  else
    tar -xzf "$asset_path" -C "$staging"
  fi

  if [[ -f "${staging}/${binary_name}" ]]; then
    cp "${staging}/${binary_name}" "${bin_dir}/${binary_name}"
  else
    found="$(find "$staging" -type f -name "$binary_name" | head -n 1 || true)"
    if [[ -z "$found" ]]; then
      rm -rf "$staging"
      echo "error: binary ${binary_name} not found in ${asset_path}" >&2
      exit 1
    fi
    cp "$found" "${bin_dir}/${binary_name}"
  fi
  rm -rf "$staging"

  if [[ "$binary_name" != *.exe ]]; then
    chmod 755 "${bin_dir}/${binary_name}"
  fi

  echo "==> npm publish ${pkg_name}@${VERSION}"
  (
    cd "$pkg_dir"
    npm publish --access "$NPM_ACCESS"
  )
done < <(node -p "require('./scripts/platform.js').PLATFORMS.map(p => p.key).join('\n')")

echo "==> npm publish @zhizai/cli@${VERSION}"
npm publish --access "$NPM_ACCESS"

echo "Done publishing @zhizai/cli and platform packages @${VERSION}"
