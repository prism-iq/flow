#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

export FLOW_ROOT="$PROJECT_ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

cleanup() {
    log_info "Shutting down services..."
    kill $(jobs -p) 2>/dev/null || true
    exit 0
}

trap cleanup SIGINT SIGTERM

start_go() {
    log_info "Starting Go orchestrator on :8080..."
    cd "$PROJECT_ROOT"
    if [ -f "bin/flow-server" ]; then
        ./bin/flow-server &
    else
        go run ./cmd/server &
    fi
}

start_node() {
    log_info "Starting Node async service on :3001..."
    cd "$PROJECT_ROOT/services/node"
    if [ -d "node_modules" ]; then
        npm start &
    else
        npm install && npm start &
    fi
}

start_python() {
    log_info "Starting Python LLM service on :8001..."
    cd "$PROJECT_ROOT/services/llm"
    if [ -d "venv" ]; then
        source venv/bin/activate
    fi
    python -m uvicorn src.main:app --host 0.0.0.0 --port 8001 &
}

start_discord() {
    if [ -z "${DISCORD_TOKEN:-}" ]; then
        log_info "DISCORD_TOKEN not set, skipping Discord bot"
        return 0
    fi

    log_info "Starting Discord bot..."
    cd "$PROJECT_ROOT/services/discord"
    if [ -d "node_modules" ]; then
        npm start &
    else
        npm install && npm start &
    fi
}

case "${1:-all}" in
    go)      start_go; wait ;;
    node)    start_node; wait ;;
    python)  start_python; wait ;;
    discord) start_discord; wait ;;
    all)
        start_python
        sleep 2
        start_go
        sleep 1
        start_node
        start_discord
        log_success "All services started"
        wait
        ;;
    *)
        echo "Usage: $0 {go|node|python|discord|all}"
        exit 1
        ;;
esac
