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
dc() {
    if docker compose version &>/dev/null; then
        docker compose "$@"
    elif command -v docker-compose &>/dev/null; then
        docker-compose "$@"
    else
        _die "Neither 'docker compose' nor 'docker-compose' found. Please install Docker."
    fi
}

# Return the compose command string (for display in messages)
dc_cmd() {
    if docker compose version &>/dev/null; then
        echo "docker compose"
    else
        echo "docker-compose"
    fi
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

# Get MySQL root password from the running container's environment
get_mysql_root_password() {
    local container
    container=$(docker inspect -f '{{.Name}}' "$(dc ps -q mysql)" | cut -c2-)
    docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$container" \
        | grep '^MYSQL_ROOT_PASSWORD=' | cut -d= -f2-
}

# Get MySQL user from the running container's environment
get_mysql_user() {
    local container
    container=$(docker inspect -f '{{.Name}}' "$(dc ps -q mysql)" | cut -c2-)
    docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$container" \
        | grep '^MYSQL_USER=' | cut -d= -f2-
}

# Get MySQL password from the running container's environment
get_mysql_password() {
    local container
    container=$(docker inspect -f '{{.Name}}' "$(dc ps -q mysql)" | cut -c2-)
    docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$container" \
        | grep '^MYSQL_PASSWORD=' | cut -d= -f2-
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
