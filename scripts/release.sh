#!/usr/bin/env bash
# Cut a release: closes the [Unreleased] section of CHANGELOG.md, commits,
# tags vX.Y.Z and pushes. The version itself lives only in the git tag; CI
# injects it into the binaries and the Docker image.
#
# Usage: scripts/release.sh 1.4.2
set -euo pipefail

VERSION="${1:-}"
[[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "usage: $0 X.Y.Z" >&2; exit 1; }
TAG="v$VERSION"

cd "$(dirname "$0")/.."
[[ -z "$(git status --porcelain)" ]] || { echo "working tree not clean" >&2; exit 1; }
[[ "$(git branch --show-current)" == "main" ]] || { echo "not on main" >&2; exit 1; }
git rev-parse -q --verify "refs/tags/$TAG" >/dev/null && { echo "tag $TAG already exists" >&2; exit 1; }
grep -q '^## \[Unreleased\]' CHANGELOG.md || { echo "no [Unreleased] section in CHANGELOG.md" >&2; exit 1; }

# Refuse to release an empty changelog section.
if ! awk '/^## \[Unreleased\]/{f=1;next} /^## \[/{f=0} f && /^### /{found=1} END{exit !found}' CHANGELOG.md; then
  echo "[Unreleased] section is empty, add entries first" >&2; exit 1
fi

DATE="$(date +%Y-%m-%d)"
perl -0pi -e "s/^## \[Unreleased\]\n/## [Unreleased]\n\n## [$VERSION] - $DATE\n/m" CHANGELOG.md

git add CHANGELOG.md
git commit -q -m "release: $TAG"
git tag "$TAG"
git push -q origin main "$TAG"
echo "released $TAG"
