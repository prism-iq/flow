#!/bin/bash
# Test Flow examples

set -e

cd "$(dirname "$0")/.."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Check for API key
if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo -e "${RED}Error: ANTHROPIC_API_KEY not set${NC}"
    echo "export ANTHROPIC_API_KEY=sk-ant-..."
    exit 1
fi

# Build flow if needed
if [ ! -f "./flow" ]; then
    ./scripts/build.sh
fi

echo "Testing Flow examples..."
echo ""

# Test hello.flow
echo "Testing hello.flow..."
OUTPUT=$(./flow run examples/hello.flow)
if [[ "$OUTPUT" == *"Hello, World!"* ]]; then
    echo -e "${GREEN}PASS${NC}: hello.flow"
else
    echo -e "${RED}FAIL${NC}: hello.flow"
    echo "Output: $OUTPUT"
fi

echo ""
echo "Tests complete."
