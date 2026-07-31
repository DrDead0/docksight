#!/usr/bin/env bash
#
# Builds the two DockSight release assets.
#
#   docksight-platform-<version>.tar.gz   the compose application
#   docksight-cli-<version>-<os>-<arch>   the CLI, one per target
#
# The two are independent artifacts: the CLI is never packaged inside the
# platform bundle, and the bundle is never embedded in the CLI. The filenames
# below must stay in sync with cmd/internal/release/asset.go, which is what
# the installer uses to discover them.
#
# Usage: scripts/build-release.sh v0.0.7

set -euo pipefail

VERSION="${1:-}"

if [ -z "$VERSION" ]; then
	echo "usage: $0 <version>   e.g. $0 v0.0.7" >&2
	exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BUNDLE="$ROOT/bundle"
CLI_DIR="$ROOT/apps/cli"
MODULE="github.com/rodriguecyber/docksight/apps/cli"
OUT="${DOCKSIGHT_RELEASE_DIR:-$ROOT/release}"

# Files the installer expects to find in the bundle. Shipping without
# .env.example leaves an installation with no generated credentials.
REQUIRED="dockersight-installation.yml .env.example default.conf"

missing=""

for file in $REQUIRED; do
	[ -f "$BUNDLE/$file" ] || missing="$missing $file"
done

if [ -n "$missing" ]; then
	echo "bundle/ is missing:$missing" >&2
	exit 1
fi

mkdir -p "$OUT"

# ---------------------------------------------------------------------------
# Platform bundle
# ---------------------------------------------------------------------------

printf '%s\n' "$VERSION" >"$BUNDLE/VERSION"

PLATFORM="$OUT/docksight-platform-$VERSION.tar.gz"

# Archived with a top-level bundle/ directory, which the installer strips.
tar -czf "$PLATFORM" -C "$ROOT" bundle

echo "built $(basename "$PLATFORM")"

# ---------------------------------------------------------------------------
# CLI binaries
# ---------------------------------------------------------------------------

TARGETS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"

for target in $TARGETS; do

	os="${target%/*}"
	arch="${target#*/}"

	ext=""
	[ "$os" = "windows" ] && ext=".exe"

	binary="$OUT/docksight-cli-$VERSION-$os-$arch$ext"

	GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -C "$CLI_DIR" \
		-ldflags "-s -w -X $MODULE/cmd/internal/buildinfo.Version=$VERSION" \
		-o "$binary" .

	echo "built $(basename "$binary")"
done

echo
echo "Upload every file above to the $VERSION GitHub release."
