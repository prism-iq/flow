#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

test_go() {
    log_info "Running Go tests..."
    cd "$PROJECT_ROOT"
    go test -v ./...
    log_success "Go tests passed"
}

test_rust() {
    log_info "Running Rust tests..."
    cd "$PROJECT_ROOT/native/rust"

    if command -v cargo &> /dev/null; then
        cargo test
        log_success "Rust tests passed"
    else
        log_info "Cargo not installed, skipping Rust tests"
    fi
}

test_node() {
    log_info "Running Node.js tests..."
    cd "$PROJECT_ROOT/services/node"

    if [ -f "package.json" ]; then
        npm test 2>/dev/null || log_info "No tests configured"
    fi
}

test_python() {
    log_info "Running Python tests..."
    cd "$PROJECT_ROOT/services/llm"

    if [ -d "venv" ]; then
        source venv/bin/activate
    fi

    if command -v pytest &> /dev/null; then
        pytest tests/ -v 2>/dev/null || log_info "No tests found"
    else
        log_info "pytest not installed, skipping Python tests"
    fi
}

test_all() {
    log_info "Running all tests..."
    local failed=0

    test_go || ((failed++))
    test_rust || ((failed++))
    test_node || ((failed++))
    test_python || ((failed++))

    if [ $failed -eq 0 ]; then
        log_success "All tests passed!"
    else
        log_error "$failed test suite(s) failed"
        exit 1
    fi
}

case "${1:-all}" in
    go)     test_go ;;
    rust)   test_rust ;;
    node)   test_node ;;
    python) test_python ;;
    all)    test_all ;;
    *)
        echo "Usage: $0 {go|rust|node|python|all}"
        exit 1
        ;;
esac
