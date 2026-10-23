#!/usr/bin/env bash
# Bump curl-impersonate to the latest upstream release (or the given tag):
# updates the Dockerfile ARG, syncs browsers.json with the release's wrapper
# scripts (new profiles get name/version/os from upstream's signature files)
# and adds a CHANGELOG entry under [Unreleased]. The "-latest" aliases follow
# browsers.json automatically.
#
# Usage: scripts/update-curl-impersonate.sh [vX.Y.Z]
# Prints "up to date" and exits 0 when there is nothing to do.
# Needs: gh, jq, tar.
set -euo pipefail

REPO="lexiforest/curl-impersonate"
cd "$(dirname "$0")/.."

CURRENT="$(sed -n 's/^ARG CURL_IMPERSONATE_VERSION=//p' Dockerfile)"
TARGET="${1:-$(gh release view -R "$REPO" --json tagName -q .tagName)}"
[[ -n "$CURRENT" && -n "$TARGET" ]] || { echo "cannot determine versions" >&2; exit 1; }
if [[ "$CURRENT" == "$TARGET" ]]; then
  echo "up to date ($CURRENT)"
  exit 0
fi

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

gh release download "$TARGET" -R "$REPO" -D "$TMP" -p "curl-impersonate-${TARGET}.x86_64-linux-gnu.tar.gz"
tar -tzf "$TMP"/*.tar.gz | sed 's|^\./||' | grep '^curl_' | sort > "$TMP/upstream"
jq -r '.browsers[].wrapper_script' browsers.json | sort > "$TMP/ours"

ADDED="$(comm -13 "$TMP/ours" "$TMP/upstream")"
REMOVED="$(comm -23 "$TMP/ours" "$TMP/upstream")"

# Build entries for new profiles from tests/signatures/<name>.yaml.
echo '[]' > "$TMP/new.json"
for wrapper in $ADDED; do
  name="${wrapper#curl_}"
  yaml="$(gh api "repos/$REPO/contents/tests/signatures/$name.yaml?ref=$TARGET" -q .content | base64 -d)"
  field() { sed -n "/^browser:/,/^[a-z]/s/^  $1: *'\{0,1\}\([^']*\)'\{0,1\}$/\1/p" <<<"$yaml" | head -1; }
  bname="$(field name)"; version="$(field version)"; os="$(field os)"
  [[ -n "$bname" && -n "$version" && -n "$os" ]] || { echo "missing metadata for $name" >&2; exit 1; }
  device=""
  case "$os" in
    windows_10*) os="win10" ;;
    macos*) os="macos" ;;
    ios*) os="ios"; device="iphone" ;;
    android*) os="android"; device="android" ;;
  esac
  jq --arg n "$name" --arg b "$bname" --arg v "$version" --arg o "$os" --arg d "$device" --arg w "$wrapper" \
    '. + [{name:$n, browser:({name:$b, version:$v, os:$o} + (if $d == "" then {} else {device:$d} end)), binary:"curl-impersonate", wrapper_script:$w}]' \
    "$TMP/new.json" > "$TMP/new.tmp" && mv "$TMP/new.tmp" "$TMP/new.json"
done

removed_json="$(printf '%s\n' $REMOVED | jq -R . | jq -s 'map(select(. != ""))')"
jq --indent 4 --slurpfile new "$TMP/new.json" --argjson removed "$removed_json" \
  '.browsers = ([.browsers[] | select(.wrapper_script as $w | $removed | index($w) | not)] + $new[0])' \
  browsers.json > "$TMP/browsers.json" && mv "$TMP/browsers.json" browsers.json

sed -i.bak "s/^ARG CURL_IMPERSONATE_VERSION=.*/ARG CURL_IMPERSONATE_VERSION=$TARGET/" Dockerfile && rm Dockerfile.bak

# Changelog entry under [Unreleased].
entry="- curl-impersonate updated from $CURRENT to $TARGET."
[[ -n "$ADDED" ]] && entry+=" New browser profiles: $(echo $ADDED | sed 's/curl_//g; s/ /, /g')."
[[ -n "$REMOVED" ]] && entry+=" Removed: $(echo $REMOVED | sed 's/curl_//g; s/ /, /g')."
# Append to an existing "### Changed" in [Unreleased], or start one.
ENTRY="$entry" perl -0pi -e '
  s/(^## \[Unreleased\]\n(?:(?!^## \[).)*?^### Changed\n)/$1$ENV{ENTRY}\n/ms
    or s/^## \[Unreleased\]\n\n?/## [Unreleased]\n\n### Changed\n$ENV{ENTRY}\n\n/m' CHANGELOG.md

echo "updated $CURRENT -> $TARGET"
[[ -n "$ADDED" ]] && echo "added: $(echo $ADDED)"
[[ -n "$REMOVED" ]] && echo "removed: $(echo $REMOVED)"
exit 0
