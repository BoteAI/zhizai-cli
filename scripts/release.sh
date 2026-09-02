#!/usr/bin/env bash
# release.sh — bump version (optional), push master, tag vX.Y.Z to trigger
# GitHub Actions: build binaries → GitHub Release → npm publish @zhizai/cli
#
# Usage:
#   ./scripts/release.sh              # use package.json version
#   ./scripts/release.sh 0.0.2        # set version then release
#   ./scripts/release.sh patch        # npm version patch (0.0.1 -> 0.0.2)
#   ./scripts/release.sh minor|major
#   ./scripts/release.sh --dry-run    # print steps only
#
# Prerequisites:
#   - clean working tree (script will commit version bump only)
#   - push access to origin
#   - GitHub Actions secret NPM_TOKEN (repo Settings → Secrets)

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DRY_RUN=0
BUMP_ARG=""

usage() {
  sed -n '2,16p' "$0" | sed 's/^# \{0,1\}//'
  exit "${1:-0}"
}

for arg in "$@"; do
  case "$arg" in
    -h|--help) usage 0 ;;
    --dry-run) DRY_RUN=1 ;;
    patch|minor|major|[0-9]*.[0-9]*.[0-9]*)
      if [[ -n "$BUMP_ARG" ]]; then
        echo "error: only one version argument allowed" >&2
        exit 1
      fi
      BUMP_ARG="$arg"
      ;;
    *)
      echo "error: unknown argument: $arg" >&2
      usage 1
      ;;
  esac
done

run() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "+ $*"
  else
    "$@"
  fi
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: missing command: $1" >&2
    exit 1
  }
}

require_cmd git
require_cmd node
require_cmd npm
require_cmd make

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$BRANCH" != "master" && "$BRANCH" != "main" ]]; then
  echo "error: release from master/main only (current: $BRANCH)" >&2
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  # Allow only package.json dirty when we are about to bump; otherwise fail.
  DIRTY="$(git status --porcelain)"
  if [[ -n "$BUMP_ARG" ]]; then
    EXTRA="$(echo "$DIRTY" | grep -vE '^( M|M |MM) package\.json$' || true)"
    if [[ -n "$EXTRA" ]]; then
      echo "error: working tree has unrelated changes; commit or stash first:" >&2
      echo "$DIRTY" >&2
      exit 1
    fi
  else
    # version already set in package.json is OK if that's the only change
    EXTRA="$(echo "$DIRTY" | grep -vE '^( M|M |MM|\?\?) package\.json$' || true)"
    # also ignore untracked .npmrc
    EXTRA="$(echo "$EXTRA" | grep -vE '^\?\? \.npmrc$' || true)"
    if [[ -n "$EXTRA" ]]; then
      echo "error: working tree not clean; commit or stash first:" >&2
      echo "$DIRTY" >&2
      exit 1
    fi
  fi
fi

if [[ -n "$BUMP_ARG" ]]; then
  case "$BUMP_ARG" in
    patch|minor|major)
      run npm version "$BUMP_ARG" --no-git-tag-version
      ;;
    *)
      run npm version "$BUMP_ARG" --no-git-tag-version
      ;;
  esac
fi

VERSION="$(node -p "require('./package.json').version")"
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
  echo "error: invalid package.json version: $VERSION" >&2
  exit 1
fi
TAG="v${VERSION}"

if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "error: tag already exists locally: $TAG" >&2
  exit 1
fi

if git ls-remote --tags origin "refs/tags/${TAG}" | grep -q "$TAG"; then
  echo "error: tag already exists on origin: $TAG" >&2
  exit 1
fi

echo "==> Running tests"
run make test

echo "==> Building"
run make build

if git diff --quiet -- package.json && git diff --cached --quiet -- package.json; then
  :
else
  echo "==> Committing package.json version ${VERSION}"
  run git add package.json
  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "+ git commit -m \"chore: release ${TAG}\""
  else
    git commit -m "chore: release ${TAG}"
  fi
fi

# Ensure package.json version is committed on the branch tip
if ! git show HEAD:package.json >/dev/null 2>&1; then
  echo "error: package.json missing from HEAD" >&2
  exit 1
fi
HEAD_VER="$(git show HEAD:package.json | node -p "JSON.parse(require('fs').readFileSync(0,'utf8')).version")"
if [[ "$HEAD_VER" != "$VERSION" ]]; then
  echo "==> Committing package.json version ${VERSION} (was ${HEAD_VER} on HEAD)"
  run git add package.json
  if [[ "$DRY_RUN" -eq 1 ]]; then
    echo "+ git commit -m \"chore: release ${TAG}\""
  else
    git commit -m "chore: release ${TAG}"
  fi
fi

echo "==> Pushing ${BRANCH}"
run git push -u origin "$BRANCH"

echo "==> Creating and pushing tag ${TAG}"
run git tag -a "$TAG" -m "Release ${TAG}"
run git push origin "$TAG"

echo
echo "Done. GitHub Actions will:"
echo "  1) build multi-platform binaries"
echo "  2) create GitHub Release ${TAG}"
echo "  3) npm publish @zhizai/cli@${VERSION}"
echo
echo "Watch: https://github.com/BoteAI/zhizai-cli/actions"
echo "npm:   https://www.npmjs.com/package/@zhizai/cli"
echo
echo "After green:"
echo "  npm install -g @zhizai/cli@${VERSION}"
echo "  zhizai --version"
echo "  zz --version"
