#!/usr/bin/env bash
# pack-connector.sh — 将 connector/ + skills/ 打成 WorkBuddy 连接器 zip
#
# Usage:
#   ./scripts/pack-connector.sh
#   make connector-zip
#
# 输出: dist/zhizai-cli-connector-<version>.zip
# 版本: 与根目录 package.json 对齐，并写入 connector-meta.json

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: missing command: $1" >&2
    exit 1
  }
}

require_cmd node
require_cmd zip

VERSION="$(node -p "require('./package.json').version")"
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+ ]]; then
  echo "error: invalid package.json version: $VERSION" >&2
  exit 1
fi

CONNECTOR_DIR="$ROOT/connector"
SKILLS_DIR="$ROOT/skills"
DIST_DIR="$ROOT/dist"
STAGING_NAME="zhizai-cli-connector"
STAGING="$DIST_DIR/.staging-$STAGING_NAME"
ZIP_NAME="${STAGING_NAME}-${VERSION}.zip"
ZIP_PATH="$DIST_DIR/$ZIP_NAME"

for f in connector-meta.json cli.json icon.svg README.md; do
  if [[ ! -f "$CONNECTOR_DIR/$f" ]]; then
    echo "error: missing connector/$f" >&2
    exit 1
  fi
done

if [[ ! -d "$SKILLS_DIR" ]]; then
  echo "error: missing skills/" >&2
  exit 1
fi

# Sync connector version with package.json (in-repo + staging).
node <<EOF
const fs = require('fs');
const path = require('path');
const metaPath = path.join('$CONNECTOR_DIR', 'connector-meta.json');
const meta = JSON.parse(fs.readFileSync(metaPath, 'utf8'));
meta.version = '$VERSION';
fs.writeFileSync(metaPath, JSON.stringify(meta, null, 2) + '\n');
EOF

# Every Skill must declare version in frontmatter (WorkBuddy).
MISSING=0
while IFS= read -r -d '' skill; do
  if ! awk '
    BEGIN { in_fm=0; found=0 }
    NR==1 && /^---$/ { in_fm=1; next }
    in_fm && /^---$/ { exit }
    in_fm && /^version:[[:space:]]*/ { found=1 }
    END { exit found ? 0 : 1 }
  ' "$skill"; then
    echo "error: missing frontmatter version: $skill" >&2
    MISSING=1
  fi
done < <(find "$SKILLS_DIR" -name SKILL.md -print0)

if [[ "$MISSING" -ne 0 ]]; then
  echo "error: 请为每个 skills/*/SKILL.md 增加 version 字段（如 version: 1.0.0）" >&2
  exit 1
fi

rm -rf "$STAGING"
mkdir -p "$STAGING/$STAGING_NAME/skills"
cp "$CONNECTOR_DIR/connector-meta.json" "$CONNECTOR_DIR/cli.json" "$CONNECTOR_DIR/icon.svg" "$CONNECTOR_DIR/README.md" \
  "$STAGING/$STAGING_NAME/"
# Copy skills tree (preserve structure); exclude macOS junk
rsync -a --exclude '.DS_Store' "$SKILLS_DIR/" "$STAGING/$STAGING_NAME/skills/"

mkdir -p "$DIST_DIR"
rm -f "$ZIP_PATH"
(
  cd "$STAGING"
  zip -r -q "$ZIP_PATH" "$STAGING_NAME"
)

rm -rf "$STAGING"

echo "OK: $ZIP_PATH"
echo "    version=$VERSION  (synced from package.json)"
echo "    unzip -l $ZIP_PATH | head"
unzip -l "$ZIP_PATH" | head -30
