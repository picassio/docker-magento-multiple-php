#!/usr/bin/env bash
#
# docker.sh — Docker & Compose helpers, PHP version detection, service checks
#
# Usage: source "$(dirname "$0")/lib/docker.sh"
# (automatically sources common.sh)
#

# Prevent double-sourcing
[[ -n "$_DOCKER_SH_LOADED" ]] && return 0
_DOCKER_SH_LOADED=1

source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

# ── Compose command (v2 preferred, v1 fallback) ──────────────────────────────
# ── Compose file resolution ───────────────────────────────────────────────────
# Maps override names to compose files.
# Usage: dc_compose_files  → returns "-f ... -f ..." flags for active overrides

# Map of override names → compose file paths
declare -A COMPOSE_OVERRIDES 2>/dev/null || true
COMPOSE_OVERRIDES=(
    [legacy]="compose/legacy.yml"
    [mysql80]="compose/mysql80.yml"
    [mariadb]="compose/mariadb.yml"
    [opensearch1]="compose/opensearch1.yml"
    [elasticsearch]="compose/elasticsearch.yml"
    [elasticsearch7]="compose/elasticsearch7.yml"
    [redis6]="compose/redis6.yml"
    [debug]="compose/debug.yml"
    [varnish]="compose/varnish.yml"
    [dashboards]="compose/dashboards.yml"
)

# Build -f flags for a list of override names
# Usage: dc_file_flags legacy mariadb redis6
dc_file_flags() {
    local flags="-f docker-compose.yml"
    for name in "$@"; do
        local file="${COMPOSE_OVERRIDES[$name]:-}"
        if [[ -n "$file" && -f "${ROOT_DIR}/${file}" ]]; then
            flags="$flags -f $file"
        fi
    done
    echo "$flags"
}

# Core compose wrapper. Reads DC_FILE_FLAGS for extra -f flags.
dc() {
    local cmd
    if docker compose version &>/dev/null; then
        cmd="docker compose"
    elif command -v docker-compose &>/dev/null; then
        cmd="docker-compose"
    else
        _die "Neither 'docker compose' nor 'docker-compose' found."
    fi

    if [[ -n "${DC_FILE_FLAGS:-}" ]]; then
        eval "$cmd $DC_FILE_FLAGS" '"$@"'
    else
        $cmd "$@"
    fi
}

dc_cmd() {
    if docker compose version &>/dev/null; then echo "docker compose"; else echo "docker-compose"; fi
}

# ── PHP version helpers ───────────────────────────────────────────────────────

# Auto-detect all PHP services defined in docker-compose.yml
get_available_php_versions() {
    dc config --services 2>/dev/null | grep '^php' | sort
}

# Validate a PHP version string against defined services
validate_php_version() {
    local php_ver="$1"

    if [[ -z "$php_ver" ]]; then
        _error "--php-version=... parameter is missing."
        _arrow "Available: $(get_available_php_versions | tr '\n' ', ' | sed 's/,$//')"
        exit 1
    fi

    if ! get_available_php_versions | grep -qx "$php_ver"; then
        _error "Invalid PHP version: $php_ver"
        _arrow "Available: $(get_available_php_versions | tr '\n' ', ' | sed 's/,$//')"
        exit 1
    fi
}

# Get the PHP FPM major.minor version number from service name
# e.g. php83 -> 8.3, php74 -> 7.4
get_php_numeric_version() {
    local svc="$1"
    local ver="${svc#php}"     # "83", "74", "81"
    local major="${ver:0:1}"   # "8", "7"
    local minor="${ver:1}"     # "3", "4", "1"
    echo "${major}.${minor}"
}

# ── Service checks ────────────────────────────────────────────────────────────

# List currently running services
list_running_services() {
    dc ps --services --filter "status=running" 2>/dev/null
}

# Check if a service is running, die if not
require_service() {
    local svc="$1"
    if ! list_running_services | grep -qx "$svc"; then
        _die "Service '$svc' is not running. Start it with: $(dc_cmd) up -d $svc"
    fi
}

# Check if a service is running (boolean, no exit)
is_service_running() {
    local svc="$1"
    list_running_services | grep -qx "$svc"
}

# ── MySQL helpers ─────────────────────────────────────────────────────────────

# ── Database helpers (supports mysql, mysql80, mariadb) ───────────────────────

# Get root password from a running DB container
# Usage: get_db_root_password [service_name]
get_db_root_password() {
    local svc="${1:-mysql}"
    local container
    container=$(docker inspect -f '{{.Name}}' "$(dc ps -q "$svc")" | cut -c2-)
    docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$container" \
        | grep '^MYSQL_ROOT_PASSWORD=' | cut -d= -f2-
}

# Backward compat
get_mysql_root_password() { get_db_root_password "${1:-mysql}"; }
get_mysql_user()          { local svc="${1:-mysql}"; local c; c=$(docker inspect -f '{{.Name}}' "$(dc ps -q "$svc")" | cut -c2-); docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$c" | grep '^MYSQL_USER=' | cut -d= -f2-; }
get_mysql_password()      { local svc="${1:-mysql}"; local c; c=$(docker inspect -f '{{.Name}}' "$(dc ps -q "$svc")" | cut -c2-); docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$c" | grep '^MYSQL_PASSWORD=' | cut -d= -f2-; }

# Validate db_service name
validate_db_service() {
    local svc="$1"
    if [[ "$svc" != @(mysql|mysql80|mariadb|postgres) ]]; then
        _die "Invalid db_service: $svc. Must be: mysql, mysql80, mariadb, or postgres"
    fi
}

validate_search_service() {
    local svc="$1"
    if [[ "$svc" != @(opensearch|opensearch1|elasticsearch|elasticsearch7|none) ]]; then
        _die "Invalid search: $svc. Must be: opensearch, opensearch1, elasticsearch, elasticsearch7, or none"
    fi
}

validate_redis_service() {
    local svc="$1"
    if [[ "$svc" != @(redis|redis6|none) ]]; then
        _die "Invalid redis: $svc. Must be: redis, redis6, or none"
    fi
}

# ── Composer auth helpers ─────────────────────────────────────────────────────

# Check and setup Composer auth for repo.magento.com
check_composer_auth() {
    local php_svc="$1"

    local public_key private_key
    public_key="$(dc exec -T --user nginx "$php_svc" composer config --global http-basic.repo.magento.com.username 2>/dev/null || true)"
    private_key="$(dc exec -T --user nginx "$php_svc" composer config --global http-basic.repo.magento.com.password 2>/dev/null || true)"

    # Trim whitespace / carriage returns
    public_key="${public_key%$'\r'}"
    private_key="${private_key%$'\r'}"

    if [[ -n "$public_key" && -n "$private_key" ]]; then
        _success "Composer auth for repo.magento.com already configured"
        return 0
    fi

    echo
    echo "Composer authentication required (repo.magento.com public and private keys):"
    read -rp "    Username (public key): " public_key
    read -rp "    Password (private key): " private_key
    echo

    if [[ -z "$public_key" || -z "$private_key" ]]; then
        _die "Composer auth keys are required. Get them from https://marketplace.magento.com/customer/accessKeys/"
    fi

    _arrow "Configuring Magento repo auth..."
    dc exec -T --user nginx "$php_svc" composer config --global http-basic.repo.magento.com "$public_key" "$private_key" \
        || _die "Cannot configure Magento auth"
    _success "Composer auth configured"
}
