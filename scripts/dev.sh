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

cleanup() {
    log_info "Stopping dev servers..."
    kill $(jobs -p) 2>/dev/null || true
    exit 0
}

trap cleanup SIGINT SIGTERM

log_info "Starting development environment..."

# Start Go with hot reload (requires air)
cd "$PROJECT_ROOT"
if command -v air &> /dev/null; then
    air &
else
    go run ./cmd/server &
fi

# Start Node with watch mode
cd "$PROJECT_ROOT/services/node"
npm run dev 2>/dev/null &

# Start frontend dev server
cd "$PROJECT_ROOT/frontend"
npm run dev &

log_success "Dev environment started"
log_info "Go:       http://localhost:8080"
log_info "Node:     http://localhost:3001"
log_info "Frontend: http://localhost:5173"

wait
