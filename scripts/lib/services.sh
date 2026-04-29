#!/usr/bin/env bash
#
# services.sh — Nginx, domain, and service-specific shared logic
#
# Usage: source "$(dirname "$0")/lib/services.sh"
# (automatically sources docker.sh → common.sh)
#

# Prevent double-sourcing
[[ -n "$_SERVICES_SH_LOADED" ]] && return 0
_SERVICES_SH_LOADED=1

source "$(dirname "${BASH_SOURCE[0]}")/docker.sh"

# ── Path constants ────────────────────────────────────────────────────────────
NGINX_CONF_DIR="${CONF_DIR}/nginx/conf.d"
NGINX_SSL_DIR="${CONF_DIR}/nginx/ssl"
DB_IMPORT_DIR="${ROOT_DIR}/databases/import"
DB_EXPORT_DIR="${ROOT_DIR}/databases/export"

# ── Domain helpers ────────────────────────────────────────────────────────────

# Sanitize domain: strip protocol/port, lowercase
sanitize_domain() {
    local domain="$1"
    # Strip http(s)://
    domain="${domain#http://}"
    domain="${domain#https://}"
    # Strip port and path
    domain="${domain%%:*}"
    domain="${domain%%/*}"
    # Lowercase
    echo "$domain" | tr '[:upper:]' '[:lower:]'
}

# Validate domain format
validate_domain() {
    local domain="$1"
    local pattern="^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,6}$"
    if [[ ! "$domain" =~ $pattern ]]; then
        _die "Invalid domain name: $domain"
    fi
}

# ── Nginx domain config checks ───────────────────────────────────────────────

# Check that an nginx vhost config exists for a domain
require_domain_config() {
    local domain="$1"
    if [[ ! -f "${NGINX_CONF_DIR}/${domain}.conf" ]]; then
        _die "No nginx config found for '${domain}'. Create it first with: ./scripts/create-vhost"
    fi
    _success "Domain config exists: ${domain}.conf"
}

# Check that an SSL vhost config exists for a domain
require_domain_ssl() {
    local domain="$1"
    if [[ ! -f "${NGINX_CONF_DIR}/${domain}-ssl.conf" ]]; then
        _die "No SSL config found for '${domain}'. Enable SSL first with: ./scripts/ssl --domain=${domain}"
    fi
    _success "Domain SSL config exists: ${domain}-ssl.conf"
}

# Check if domain config exists (boolean, no exit)
has_domain_config() {
    local domain="$1"
    [[ -f "${NGINX_CONF_DIR}/${domain}.conf" ]]
}

# Check if SSL config exists (boolean, no exit)
has_domain_ssl() {
    local domain="$1"
    [[ -f "${NGINX_CONF_DIR}/${domain}-ssl.conf" ]]
}

# ── Nginx domain introspection ────────────────────────────────────────────────

# Extract PHP version from an existing nginx vhost config
# e.g. "fastcgi_pass php83:9001;" → "php83"
get_domain_php_version() {
    local domain="$1"
    local conf="${NGINX_CONF_DIR}/${domain}.conf"
    [[ -f "$conf" ]] || _die "Config not found: $conf"
    grep ':9001' "$conf" | awk 'NR==1{print $2}' | cut -d: -f1
}

# Extract document root folder name from nginx vhost config
# e.g. "root /home/public_html/mysite.com;" → "mysite.com"
get_domain_docroot() {
    local domain="$1"
    local conf="${NGINX_CONF_DIR}/${domain}.conf"
    [[ -f "$conf" ]] || _die "Config not found: $conf"
    grep '/home/public_html/' "$conf" \
        | grep -v 'fastcgi_param' \
        | head -1 \
        | sed 's|.*/home/public_html/||' \
        | sed 's|[;/ ].*||'
}

# ── Nginx operations ─────────────────────────────────────────────────────────

# Test nginx config and reload
reload_nginx() {
    _arrow "Testing nginx configuration..."
    if dc exec nginx nginx -t 2>&1; then
        dc exec nginx nginx -s reload || _die "Nginx reload failed"
        _success "Nginx reloaded"
    else
        _die "Nginx configuration test failed. Check your config files."
    fi
}

# ── /etc/hosts management ────────────────────────────────────────────────────

# Add domain to /etc/hosts if not already present
add_etc_hosts() {
    local domain="$1"
    if grep -qE "127\.0\.0\.1[[:space:]]+${domain}" /etc/hosts 2>/dev/null; then
        _warning "${domain} already exists in /etc/hosts"
    else
        _arrow "Adding ${domain} to /etc/hosts (may require sudo password)..."
        echo "127.0.0.1  ${domain}" | sudo tee -a /etc/hosts >/dev/null \
            || _die "Cannot write to /etc/hosts"
        _success "Added ${domain} to /etc/hosts"
    fi
}

# ── Common service requirements ───────────────────────────────────────────────

# Require nginx + a PHP version to be running
require_web_stack() {
    local php_ver="$1"
    require_service "nginx"
    require_service "$php_ver"
}

# Require nginx + mysql + a PHP version to be running
require_full_stack() {
    local php_ver="$1"
    require_service "nginx"
    require_service "mysql"
    require_service "$php_ver"
}
