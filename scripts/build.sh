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
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

build_go() {
    log_info "Building Go orchestrator..."
    cd "$PROJECT_ROOT"

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        return 1
    fi

    go mod tidy
    CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/flow-server ./cmd/server

    log_success "Go build complete: bin/flow-server"
}

build_rust() {
    log_info "Building Rust parser..."
    cd "$PROJECT_ROOT/native/rust"

    if ! command -v cargo &> /dev/null; then
        log_warn "Cargo not installed, skipping Rust build"
        return 0
    fi

    cargo build --release

    log_success "Rust build complete: target/release/libflow_parser.so"
}

build_cpp() {
    log_info "Building C++ synapses module..."
    cd "$PROJECT_ROOT/native/cpp"

    if ! command -v cmake &> /dev/null; then
        log_warn "CMake not installed, skipping C++ build"
        return 0
    fi

    mkdir -p build
    cd build
    cmake -DCMAKE_BUILD_TYPE=Release ..
    make -j$(nproc)

    log_success "C++ build complete"
}

build_frontend() {
    log_info "Building Svelte frontend..."
    cd "$PROJECT_ROOT/frontend"

    if ! command -v npm &> /dev/null; then
        log_warn "npm not installed, skipping frontend build"
        return 0
    fi

    npm install
    npm run build

    log_success "Frontend build complete: build/"
}

build_all() {
    log_info "Starting full build..."

    mkdir -p "$PROJECT_ROOT/bin"

    build_go
    build_rust
    build_cpp
    build_frontend

    log_success "All builds complete!"
}

case "${1:-all}" in
    go)      build_go ;;
    rust)    build_rust ;;
    cpp)     build_cpp ;;
    frontend) build_frontend ;;
    all)     build_all ;;
    *)
        echo "Usage: $0 {go|rust|cpp|frontend|all}"
        exit 1
        ;;
esac
