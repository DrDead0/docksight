#!/usr/bin/env bash
#
# Uploads every artifact built for a version to its GitHub release, then
# verifies the release actually carries all of them.
#
# Uploading by hand is easy to get half-right: the agent binaries were missed
# on four consecutive releases, and the failure only surfaces later on a
# remote host as "no asset matching docksight-agent-*".
#
# Usage: scripts/publish-release.sh v0.0.12

set -euo pipefail

VERSION="${1:-}"

if [ -z "$VERSION" ]; then
	echo "usage: $0 <version>   e.g. $0 v0.0.12" >&2
	exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${DOCKSIGHT_RELEASE_DIR:-$ROOT/release}"
# GITHUB_REPOSITORY is set by GitHub Actions (owner/repo of the run), so the
# workflow targets the repository it runs in; manually it defaults upstream.
REPO="${GITHUB_REPOSITORY:-Open-Source-Kigali/docksight}"

if ! command -v gh >/dev/null 2>&1; then
	echo "gh is not installed: https://cli.github.com" >&2
	exit 1
fi

# Everything the release must carry. Keep in sync with build-release.sh.
EXPECTED="
docksight-platform-$VERSION.tar.gz
docksight-cli-$VERSION-linux-amd64
docksight-cli-$VERSION-linux-arm64
docksight-cli-$VERSION-darwin-amd64
docksight-cli-$VERSION-darwin-arm64
docksight-cli-$VERSION-windows-amd64.exe
docksight-agent-$VERSION-linux-amd64
docksight-agent-$VERSION-linux-arm64
docksight-agent-$VERSION-windows-amd64.exe
"

# ---------------------------------------------------------------------------
# Every artifact must exist locally before anything is uploaded
# ---------------------------------------------------------------------------

missing=""

for asset in $EXPECTED; do
	[ -f "$OUT/$asset" ] || missing="$missing $asset"
done

if [ -n "$missing" ]; then
	echo "not built yet:$missing" >&2
	echo "run: scripts/build-release.sh $VERSION" >&2
	exit 1
fi

# ---------------------------------------------------------------------------
# Upload
# ---------------------------------------------------------------------------

if ! gh release view "$VERSION" --repo "$REPO" >/dev/null 2>&1; then
	echo "release $VERSION does not exist, creating it"
	gh release create "$VERSION" --repo "$REPO" --title "DockSight $VERSION" --notes ""
fi

for asset in $EXPECTED; do
	gh release upload "$VERSION" "$OUT/$asset" --repo "$REPO" --clobber
	echo "uploaded $asset"
done

# ---------------------------------------------------------------------------
# Verify — the release, not the local directory, is what installers read
# ---------------------------------------------------------------------------

published="$(gh release view "$VERSION" --repo "$REPO" --json assets --jq '.assets[].name')"

absent=""

for asset in $EXPECTED; do
	printf '%s\n' "$published" | grep -qx "$asset" || absent="$absent $asset"
done

if [ -n "$absent" ]; then
	echo "upload incomplete, missing from the release:$absent" >&2
	exit 1
fi

echo
echo "$VERSION publishes all $(printf '%s\n' $EXPECTED | wc -l | tr -d ' ') assets"

# A release that is a draft or a prerelease is invisible to /releases/latest,
# so installers would keep resolving the previous version.
state="$(gh release view "$VERSION" --repo "$REPO" --json isDraft,isPrerelease --jq '[.isDraft, .isPrerelease] | @tsv')"

case "$state" in
*true*)
	echo "warning: $VERSION is a draft or prerelease — 'latest' will not resolve to it" >&2
	;;
esac
