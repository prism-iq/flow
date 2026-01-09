#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

VERSION="${VERSION:-latest}"
REGISTRY="${REGISTRY:-}"

log_info() { echo -e "\033[0;34m[INFO]\033[0m $1"; }
log_success() { echo -e "\033[0;32m[SUCCESS]\033[0m $1"; }

build_image() {
    local name=$1
    local context=$2
    local dockerfile=$3

    local tag="flow-${name}:${VERSION}"
    if [ -n "$REGISTRY" ]; then
        tag="${REGISTRY}/${tag}"
    fi

    log_info "Building ${tag}..."
    docker build -t "$tag" -f "$dockerfile" "$context"
    log_success "Built ${tag}"
}

# Build Go orchestrator
build_image "orchestrator" "$PROJECT_ROOT" "$PROJECT_ROOT/Dockerfile"

# Build Node service
build_image "node-service" "$PROJECT_ROOT/services/node" "$PROJECT_ROOT/services/node/Dockerfile"

# Build Python LLM service
build_image "llm-service" "$PROJECT_ROOT/services/llm" "$PROJECT_ROOT/services/llm/Dockerfile"

# Build Discord bot
build_image "discord-bot" "$PROJECT_ROOT/services/discord" "$PROJECT_ROOT/services/discord/Dockerfile"

# Build frontend
build_image "frontend" "$PROJECT_ROOT/frontend" "$PROJECT_ROOT/frontend/Dockerfile"

log_success "All images built!"
