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

check_dependencies() {
    log_info "Checking dependencies..."

    local missing=()

    command -v go &> /dev/null || missing+=("go")
    command -v node &> /dev/null || missing+=("node")
    command -v npm &> /dev/null || missing+=("npm")
    command -v python3 &> /dev/null || missing+=("python3")

    if [ ${#missing[@]} -ne 0 ]; then
        log_error "Missing dependencies: ${missing[*]}"
        log_info "Please install them before continuing"
        exit 1
    fi

    log_success "All required dependencies found"
}

install_go_deps() {
    log_info "Installing Go dependencies..."
    cd "$PROJECT_ROOT"
    go mod download
    log_success "Go dependencies installed"
}

install_node_deps() {
    log_info "Installing Node.js dependencies..."

    cd "$PROJECT_ROOT/services/node"
    npm install

    cd "$PROJECT_ROOT/services/discord"
    npm install

    cd "$PROJECT_ROOT/frontend"
    npm install

    log_success "Node.js dependencies installed"
}

install_python_deps() {
    log_info "Installing Python dependencies..."
    cd "$PROJECT_ROOT/services/llm"

    if [ ! -d "venv" ]; then
        python3 -m venv venv
    fi

    source venv/bin/activate
    pip install -r requirements.txt

    log_success "Python dependencies installed"
}

setup_env() {
    log_info "Setting up environment..."

    if [ ! -f "$PROJECT_ROOT/.env" ]; then
        cat > "$PROJECT_ROOT/.env" << 'EOF'
# Flow Configuration
PORT=8080
LOG_LEVEL=info

# Service URLs
LLM_SERVICE_URL=http://localhost:8001
RAG_SERVICE_URL=http://localhost:8002
RUST_SERVICE_URL=http://localhost:8003

# Database
PG_HOST=localhost
PG_PORT=5432
PG_DATABASE=flow
PG_USER=flow
PG_PASSWORD=flow

# Discord (optional)
# DISCORD_TOKEN=your_token
# DISCORD_CLIENT_ID=your_client_id
# DISCORD_GUILD_ID=your_guild_id

# LLM Settings
LLM_MODEL_NAME=microsoft/Phi-3-mini-4k-instruct
LLM_NUM_WORKERS=4
EOF
        log_success "Created .env file"
    else
        log_info ".env file already exists"
    fi
}

main() {
    log_info "Flow Installation Script"
    echo ""

    check_dependencies
    install_go_deps
    install_node_deps
    install_python_deps
    setup_env

    echo ""
    log_success "Installation complete!"
    log_info "Run './scripts/build.sh' to build the project"
    log_info "Run './scripts/start.sh' to start all services"
}

main
