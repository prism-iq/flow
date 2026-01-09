#!/bin/bash
# Build the Flow transpiler

set -e

cd "$(dirname "$0")/.."

echo "Building Flow..."

# Download dependencies
go mod tidy

# Build
go build -o flow cmd/flow/main.go

echo "Built: ./flow"
echo ""
echo "To install globally:"
echo "  sudo cp flow /usr/local/bin/"
