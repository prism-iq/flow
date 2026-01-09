#!/bin/bash
# Install Flow to /usr/local/bin

set -e

cd "$(dirname "$0")/.."

# Build first
./scripts/build.sh

# Install
echo "Installing Flow..."
sudo cp flow /usr/local/bin/flow
sudo chmod +x /usr/local/bin/flow

# Create config directory
mkdir -p ~/.config/flow

# Copy docs for self-documenting feature
sudo mkdir -p /opt/flow/docs
sudo cp -r docs/* /opt/flow/docs/ 2>/dev/null || true

echo ""
echo "Flow installed successfully!"
echo ""
echo "Usage:"
echo "  flow run hello.flow"
echo "  flow build hello.flow"
echo "  flow show hello.flow"
echo ""
echo "Make sure to set ANTHROPIC_API_KEY:"
echo "  export ANTHROPIC_API_KEY=sk-ant-..."
