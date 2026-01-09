#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Database config
DB_HOST="${PG_HOST:-localhost}"
DB_PORT="${PG_PORT:-5432}"
DB_NAME="${PG_DATABASE:-flow}"
DB_USER="${PG_USER:-flow}"
DB_PASS="${PG_PASSWORD:-flow}"

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

PGPASSWORD="$DB_PASS"
export PGPASSWORD

check_postgres() {
    log_info "Checking PostgreSQL connection..."
    if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" > /dev/null 2>&1; then
        log_error "PostgreSQL is not running at $DB_HOST:$DB_PORT"
        exit 1
    fi
    log_success "PostgreSQL is running"
}

check_age() {
    log_info "Checking Apache AGE extension..."
    result=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT 1 FROM pg_extension WHERE extname = 'age'" 2>/dev/null || echo "")

    if [ "$result" != "1" ]; then
        log_info "Installing Apache AGE extension..."
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "CREATE EXTENSION IF NOT EXISTS age;" 2>/dev/null || true
    fi
    log_success "Apache AGE is available"
}

create_database() {
    log_info "Creating database if not exists..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "CREATE DATABASE $DB_NAME;" 2>/dev/null || true
    log_success "Database $DB_NAME ready"
}

run_migrations() {
    log_info "Running migrations..."

    MIGRATIONS_DIR="$PROJECT_ROOT/db/migrations"

    for migration in "$MIGRATIONS_DIR"/*.sql; do
        if [ -f "$migration" ]; then
            filename=$(basename "$migration")
            log_info "Applying: $filename"
            psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration" 2>&1 || {
                log_error "Migration failed: $filename"
                continue
            }
            log_success "Applied: $filename"
        fi
    done

    log_success "All migrations complete"
}

show_stats() {
    log_info "Database statistics:"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT 'conversations' as table_name, COUNT(*) as count FROM conversations
        UNION ALL
        SELECT 'messages', COUNT(*) FROM messages
        UNION ALL
        SELECT 'emails', COUNT(*) FROM emails
        UNION ALL
        SELECT 'entities', COUNT(*) FROM entities;
    " 2>/dev/null || true
}

case "${1:-up}" in
    up)
        check_postgres
        create_database
        check_age
        run_migrations
        show_stats
        ;;
    down)
        log_info "Dropping all tables..."
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
            DROP TABLE IF EXISTS extracted_dates CASCADE;
            DROP TABLE IF EXISTS extracted_amounts CASCADE;
            DROP TABLE IF EXISTS email_participants CASCADE;
            DROP TABLE IF EXISTS emails CASCADE;
            DROP TABLE IF EXISTS entities CASCADE;
            DROP TABLE IF EXISTS documents CASCADE;
            DROP TABLE IF EXISTS messages CASCADE;
            DROP TABLE IF EXISTS conversations CASCADE;
        "
        log_success "Tables dropped"
        ;;
    reset)
        $0 down
        $0 up
        ;;
    status)
        check_postgres
        show_stats
        ;;
    *)
        echo "Usage: $0 {up|down|reset|status}"
        exit 1
        ;;
esac
