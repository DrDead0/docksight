#!/usr/bin/env bash

set -euo pipefail

APP_NAME="docksight-agent"

INSTALL_DIR="/opt/docksight-agent"
CONFIG_DIR="/etc/docksight"

BINARY_PATH="$INSTALL_DIR/$APP_NAME"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
VERSION_FILE="$CONFIG_DIR/version"

SERVICE_NAME="docksight-agent.service"
SERVICE_PATH="/etc/systemd/system/$SERVICE_NAME"

GITHUB_REPO="docksight/docksight"

echo "======================================"
echo " DockSight Agent Installer"
echo "======================================"

#
# Check OS
#

if [[ "$(uname -s)" != "Linux" ]]; then
    echo "ERROR: Only Linux is supported currently."
    exit 1
fi


#
# Detect architecture
#

ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ASSET="docksight-agent-linux-amd64"
        ;;

    aarch64|arm64)
        ASSET="docksight-agent-linux-arm64"
        ;;

    *)
        echo "ERROR: Unsupported architecture: $ARCH"
        exit 1
        ;;
esac


echo "Architecture detected: $ARCH"


#
# Check dependencies
#

for cmd in curl systemctl; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "ERROR: Missing dependency: $cmd"
        exit 1
    fi
done


#
# Create system user
#

if ! id docksight >/dev/null 2>&1; then

    echo "Creating docksight system user..."

    useradd \
        --system \
        --no-create-home \
        --shell /usr/sbin/nologin \
        docksight

fi


#
# Docker permission
#

if getent group docker >/dev/null 2>&1; then

    echo "Adding docksight user to docker group..."

    usermod -aG docker docksight

else

    echo "Docker group not found. Skipping."

fi


#
# Create directories
#

mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"


#
# Detect installation mode
#

if [[ -f "$BINARY_PATH" ]]; then

    MODE="upgrade"

else

    MODE="install"

fi


echo "Installation mode: $MODE"


#
# Download latest binary
#

DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/latest/download/$ASSET"


echo "Downloading:"
echo "$DOWNLOAD_URL"


TMP_BINARY="/tmp/$APP_NAME-new"


curl -L \
    "$DOWNLOAD_URL" \
    -o "$TMP_BINARY"


chmod +x "$TMP_BINARY"



#
# Stop existing service during upgrade
#

if [[ "$MODE" == "upgrade" ]]; then

    echo "Stopping existing service..."

    systemctl stop "$SERVICE_NAME" || true


    echo "Backing up old binary..."

    mv \
        "$BINARY_PATH" \
        "$BINARY_PATH.backup"

fi



#
# Install binary
#

echo "Installing binary..."

mv \
    "$TMP_BINARY" \
    "$BINARY_PATH"


chown docksight:docksight "$BINARY_PATH"



#
# Create config only first time
#

if [[ ! -f "$CONFIG_FILE" ]]; then

    echo "Creating default configuration..."

    cat > "$CONFIG_FILE" <<EOF

server:
  url: ws://localhost:8080/agents

agent:
  uuid: ""

host:
  hostname: $(hostname)

EOF

else

    echo "Existing configuration found. Keeping it."

fi


chown -R docksight:docksight "$CONFIG_DIR"



#
# Install systemd service
#

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"


if [[ -f "$SCRIPT_DIR/docksight-agent.service" ]]; then

    echo "Installing systemd service..."

    cp \
        "$SCRIPT_DIR/docksight-agent.service" \
        "$SERVICE_PATH"

fi



#
# Reload systemd
#

systemctl daemon-reload


#
# Enable service
#

systemctl enable "$SERVICE_NAME"



#
# Start service
#

echo "Starting DockSight Agent..."

systemctl restart "$SERVICE_NAME"



#
# Save version
#

if "$BINARY_PATH" --version >/dev/null 2>&1; then

    "$BINARY_PATH" --version \
        > "$VERSION_FILE"

else

    echo "unknown" > "$VERSION_FILE"

fi


#
# Verify service
#

sleep 3


if systemctl is-active --quiet "$SERVICE_NAME"; then

    echo ""
    echo "======================================"
    echo " DockSight Agent Installed Successfully"
    echo "======================================"

    echo ""
    echo "Service:"
    systemctl status "$SERVICE_NAME" --no-pager

else

    echo ""
    echo "ERROR: DockSight Agent failed to start."

    echo ""
    echo "Logs:"
    journalctl -u "$SERVICE_NAME" --no-pager -n 50

    exit 1

fi