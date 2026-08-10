#!/usr/bin/env bash
#
# Builds the DockSight release assets.
#
#   docksight-platform-<version>.tar.gz     the compose application
#   docksight-cli-<version>-<os>-<arch>     the CLI, one per target
#   docksight-agent-<version>-<os>-<arch>   the agent, one per target
#
# These are independent artifacts: the CLI is never packaged inside the
# platform bundle, the bundle is never embedded in the CLI, and the agent
# ships on its own because it is installed on remote Docker hosts. The
# filenames below must stay in sync with cmd/internal/release/asset.go, which
# is what the installer uses to discover them.
#
# Usage: scripts/build-release.sh v0.0.1

set -euo pipefail

VERSION="${1:-}"

if [ -z "$VERSION" ]; then
	echo "usage: $0 <version>   e.g. $0 v0.0.1" >&2
	exit 1
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
COMMIT="$(git -C "$ROOT" rev-parse --short HEAD)"
BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
BUNDLE="$ROOT/bundle"
CLI_DIR="$ROOT/apps/cli"
AGENT_DIR="$ROOT/apps/agent"
MODULE="github.com/Open-Source-Kigali/docksight/apps/cli"
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
		-ldflags "-s -w \
			-X $MODULE/cmd/internal/buildinfo.Version=$VERSION \
			-X $MODULE/cmd/internal/buildinfo.Commit=$COMMIT \
			-X $MODULE/cmd/internal/buildinfo.BuildDate=$BUILD_DATE" \
		-o "$binary" .

	# Building on a filesystem without a Unix execute bit (NTFS) produces a
	# 644 binary, and scp preserves that mode all the way to the server.
	chmod +x "$binary" 2>/dev/null || true

	echo "built $(basename "$binary")"
done

# ---------------------------------------------------------------------------
# Agent binaries
#
# The agent is its own Go module and runs on remote Docker hosts, so it is
# built only for the platforms an agent can run beside a Docker Engine — no
# darwin, where the engine lives inside a VM the agent cannot sit next to.
# ---------------------------------------------------------------------------

AGENT_TARGETS="linux/amd64 linux/arm64 windows/amd64"

for target in $AGENT_TARGETS; do

	os="${target%/*}"
	arch="${target#*/}"

	# Windows will not execute an extension-less file, and
	# release.AgentBinaryName() already appends .exe via Target.Extension() —
	# publishing without it makes discovery ask for a name the release does
	# not carry.
	ext=""
	[ "$os" = "windows" ] && ext=".exe"

	binary="$OUT/docksight-agent-$VERSION-$os-$arch$ext"

	# Agent is its own module (docksight-agent), not the CLI module path.
	GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -C "$AGENT_DIR" \
		-ldflags "-s -w \
			-X docksight-agent/internal/version.Version=$VERSION \
			-X docksight-agent/internal/version.Commit=$COMMIT \
			-X docksight-agent/internal/version.BuildDate=$BUILD_DATE" \
		-o "$binary" ./cmd/agent

	chmod +x "$binary" 2>/dev/null || true

	echo "built $(basename "$binary")"
done

echo
echo "Upload every file above to the $VERSION GitHub release."
