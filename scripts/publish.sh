#!/usr/bin/env bash
# publish.sh — 一键发布 @zhizai/cli
#
# 流程（本机只做前半段，后半段由 GitHub Actions 完成）：
#   1. 升版本（默认 patch）并写入 package.json
#   2. make test / make build
#   3. 提交 chore: release vX.Y.Z
#   4. push master + 打 annotated tag vX.Y.Z
#   5. Actions：交叉编译 → GitHub Release → npm publish
#
# 用法：
#   ./scripts/publish.sh              # 一键：默认 patch，并确认
#   ./scripts/publish.sh -y           # 一键：默认 patch，跳过确认
#   ./scripts/publish.sh 0.0.6        # 指定版本
#   ./scripts/publish.sh minor -y     # minor 且不确认
#   ./scripts/publish.sh --dry-run    # 只打印将执行的步骤
#   make publish / make publish V=patch
#
# 前置：master/main 干净工作区、可 push、仓库已配置 NPM_TOKEN Secret

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

YES=0
DRY_RUN=0
BUMP="patch"
EXTRA_ARGS=()

usage() {
  awk 'NR==1{next} /^#/{sub(/^# ?/,""); print; next} {exit}' "$0"
  exit "${1:-0}"
}

for arg in "$@"; do
  case "$arg" in
    -h|--help) usage 0 ;;
    -y|--yes) YES=1 ;;
    --dry-run) DRY_RUN=1; EXTRA_ARGS+=(--dry-run) ;;
    patch|minor|major|[0-9]*.[0-9]*.[0-9]*)
      BUMP="$arg"
      ;;
    *)
      echo "error: unknown argument: $arg" >&2
      usage 1
      ;;
  esac
done

CUR="$(node -p "require('./package.json').version")"

# 计算将变成的版本（不改文件）
NEXT="$(
  node -e "
    const cur = process.argv[1];
    const bump = process.argv[2];
    const [a,b,c] = cur.split('.').map(Number);
    if (/^[0-9]+\\.[0-9]+\\.[0-9]+$/.test(bump)) {
      console.log(bump);
      process.exit(0);
    }
    if (bump === 'patch') console.log([a,b,c+1].join('.'));
    else if (bump === 'minor') console.log([a,b+1,0].join('.'));
    else if (bump === 'major') console.log([a+1,0,0].join('.'));
    else { console.error('bad bump'); process.exit(1); }
  " "$CUR" "$BUMP"
)"

cat <<EOF

======== 智在记录 CLI 一键发布 ========
当前版本:  ${CUR}
目标版本:  ${NEXT}  (bump=${BUMP})
分支:      $(git rev-parse --abbrev-ref HEAD)
远端:      $(git remote get-url origin 2>/dev/null || echo '?')

本机将执行:
  • 测通 + 构建
  • 提交 package.json → ${NEXT}
  • git push + tag v${NEXT}

GitHub Actions 随后自动:
  • 多平台二进制 → GitHub Release v${NEXT}
  • npm publish @zhizai/cli@${NEXT}

进度: https://github.com/BoteAI/zhizai-cli/actions
=======================================

EOF

if [[ "$DRY_RUN" -eq 1 ]]; then
  echo "(dry-run) 不会真正改仓库 / 推送"
fi

if [[ "$YES" -ne 1 ]]; then
  printf "确认发布 v%s？[y/N] " "$NEXT"
  read -r ans
  case "$ans" in
    y|Y|yes|YES) ;;
    *)
      echo "已取消"
      exit 1
      ;;
  esac
fi

exec bash "$ROOT/scripts/release.sh" "$BUMP" "${EXTRA_ARGS[@]+"${EXTRA_ARGS[@]}"}"
