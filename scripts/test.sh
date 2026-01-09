#!/bin/bash
# Test Flow examples

set -e

cd "$(dirname "$0")/.."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Build flow if needed
if [ ! -f "./flow" ]; then
    ./scripts/build.sh
fi

echo "Testing Flow examples..."
echo ""

PASSED=0
FAILED=0

# Test each example
for example in examples/*.flow; do
    name=$(basename "$example")
    echo -n "Testing $name... "

    if OUTPUT=$(./flow run "$example" 2>&1); then
        echo -e "${GREEN}PASS${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}FAIL${NC}"
        echo "Output: $OUTPUT"
        FAILED=$((FAILED + 1))
    fi
done

echo ""
echo "Results: $PASSED passed, $FAILED failed"

if [ $FAILED -gt 0 ]; then
    exit 1
fi
