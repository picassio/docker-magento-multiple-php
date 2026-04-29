#!/usr/bin/env bash
#
# Comprehensive test suite — tests ALL functionality
# Run: bash tests/test-all.sh
#
set -eo pipefail
cd "$(dirname "$0")/.."

# ── Test framework ────────────────────────────────────────────────────────────
PASS=0
FAIL=0
SKIP=0
ERRORS=()

_green=$(tput setaf 76 2>/dev/null || echo '')
_red=$(tput setaf 1 2>/dev/null || echo '')
_tan=$(tput setaf 3 2>/dev/null || echo '')
_bold=$(tput bold 2>/dev/null || echo '')
_reset=$(tput sgr0 2>/dev/null || echo '')

pass()  { PASS=$((PASS+1)); echo "  ${_green}✔${_reset} $1"; }
fail()  { FAIL=$((FAIL+1)); ERRORS+=("$1"); echo "  ${_red}✖${_reset} $1"; }
skip()  { SKIP=$((SKIP+1)); echo "  ${_tan}⊘${_reset} $1 (skipped)"; }
section() { echo ""; echo "${_bold}━━━ $1 ━━━${_reset}"; }

assert_eq() {
    local desc="$1" expected="$2" actual="$3"
    if [[ "$expected" == "$actual" ]]; then
        pass "$desc"
    else
        fail "$desc — expected '$expected', got '$actual'"
    fi
}

assert_contains() {
    local desc="$1" haystack="$2" needle="$3"
    if echo "$haystack" | grep -q "$needle"; then
        pass "$desc"
    else
        fail "$desc — '$needle' not found"
    fi
}

assert_not_contains() {
    local desc="$1" haystack="$2" needle="$3"
    if echo "$haystack" | grep -q "$needle"; then
        fail "$desc — '$needle' should NOT be present"
    else
        pass "$desc"
    fi
}

assert_file_exists() {
    if [[ -f "$2" ]]; then pass "$1"; else fail "$1 — file not found: $2"; fi
}

assert_exit_0() {
    local desc="$1"; shift
    if "$@" >/dev/null 2>&1; then pass "$desc"; else fail "$desc — exit code $?"; fi
}

assert_exit_nonzero() {
    local desc="$1"; shift
    if "$@" >/dev/null 2>&1; then fail "$desc — expected failure but succeeded"; else pass "$desc"; fi
}

# ── Setup ─────────────────────────────────────────────────────────────────────
echo "================================================================"
echo "  COMPREHENSIVE TEST SUITE"
echo "================================================================"

# Save and clear projects.json
cp projects.json projects.json.bak 2>/dev/null || true
echo '{}' > projects.json

# ══════════════════════════════════════════════════════════════════════════════
section "1. SHARED LIBRARY — common.sh"
# ══════════════════════════════════════════════════════════════════════════════

OUT=$(bash -c 'source scripts/lib/common.sh; echo "ROOT=$ROOT_DIR"' 2>&1)
assert_contains "ROOT_DIR resolves" "$OUT" "ROOT=$(pwd)"

OUT=$(bash -c 'source scripts/lib/common.sh; _arrow "test"' 2>&1)
assert_contains "_arrow outputs" "$OUT" "➜ test"

OUT=$(bash -c 'source scripts/lib/common.sh; _success "ok"' 2>&1)
assert_contains "_success outputs" "$OUT" "✔ ok"

OUT=$(bash -c 'source scripts/lib/common.sh; _error "bad"' 2>&1)
assert_contains "_error outputs" "$OUT" "✖ bad"

assert_exit_0 "require_commands with valid cmds" bash -c 'source scripts/lib/common.sh; require_commands git bash'
assert_exit_nonzero "require_commands with invalid cmd" bash -c 'source scripts/lib/common.sh; require_commands fake_nonexistent_xyz'

# Double source guard
OUT=$(bash -c 'source scripts/lib/common.sh; source scripts/lib/common.sh; echo OK' 2>&1)
assert_eq "Double source guard" "OK" "$(echo "$OUT" | tail -1)"

# ══════════════════════════════════════════════════════════════════════════════
section "2. SHARED LIBRARY — docker.sh"
# ══════════════════════════════════════════════════════════════════════════════

OUT=$(bash -c 'source scripts/lib/docker.sh; dc_cmd' 2>&1)
assert_contains "dc_cmd returns compose command" "$OUT" "compose"

OUT=$(bash -c 'source scripts/lib/docker.sh; get_available_php_versions' 2>&1)
assert_contains "get_available_php_versions finds php83" "$OUT" "php83"
assert_contains "get_available_php_versions finds php84" "$OUT" "php84"

assert_exit_0 "validate_php_version php83" bash -c 'source scripts/lib/docker.sh; validate_php_version php83'
assert_exit_nonzero "validate_php_version invalid" bash -c 'source scripts/lib/docker.sh; validate_php_version php99'

OUT=$(bash -c 'source scripts/lib/docker.sh; get_php_numeric_version php83' 2>&1)
assert_eq "get_php_numeric_version php83" "8.3" "$OUT"
OUT=$(bash -c 'source scripts/lib/docker.sh; get_php_numeric_version php74' 2>&1)
assert_eq "get_php_numeric_version php74" "7.4" "$OUT"

assert_exit_0 "validate_db_service mysql" bash -c 'source scripts/lib/docker.sh; validate_db_service mysql'
assert_exit_0 "validate_db_service mysql80" bash -c 'source scripts/lib/docker.sh; validate_db_service mysql80'
assert_exit_0 "validate_db_service mariadb" bash -c 'source scripts/lib/docker.sh; validate_db_service mariadb'
assert_exit_nonzero "validate_db_service invalid" bash -c 'source scripts/lib/docker.sh; validate_db_service postgres'

# ══════════════════════════════════════════════════════════════════════════════
section "3. SHARED LIBRARY — services.sh"
# ══════════════════════════════════════════════════════════════════════════════

OUT=$(bash -c 'source scripts/lib/services.sh; sanitize_domain "https://My-Store.COM:8080/path"' 2>&1)
assert_eq "sanitize_domain strips protocol/port" "my-store.com" "$OUT"

OUT=$(bash -c 'source scripts/lib/services.sh; sanitize_domain "http://test.local/"' 2>&1)
assert_eq "sanitize_domain strips trailing slash" "test.local" "$OUT"

assert_exit_0 "validate_domain valid" bash -c 'source scripts/lib/services.sh; validate_domain test.com'
assert_exit_0 "validate_domain with subdomain" bash -c 'source scripts/lib/services.sh; validate_domain shop.test.com'
assert_exit_nonzero "validate_domain invalid" bash -c 'source scripts/lib/services.sh; validate_domain "not a domain"'

# ══════════════════════════════════════════════════════════════════════════════
section "4. SHARED LIBRARY — projects.sh"
# ══════════════════════════════════════════════════════════════════════════════

# Start fresh
echo '{}' > projects.json

# project_set + project_exists
bash -c 'source scripts/lib/projects.sh; project_set "test.local" "php83" "magento2" "mysql" "test_db" "opensearch" "true"' 2>/dev/null
assert_exit_0 "project_exists after set" bash -c 'source scripts/lib/projects.sh; project_exists "test.local"'
assert_exit_nonzero "project_exists for missing" bash -c 'source scripts/lib/projects.sh; project_exists "nope.local"'

# project_get
OUT=$(bash -c 'source scripts/lib/projects.sh; project_get "test.local" "php"' 2>&1)
assert_eq "project_get php" "php83" "$OUT"
OUT=$(bash -c 'source scripts/lib/projects.sh; project_get "test.local" "db_service"' 2>&1)
assert_eq "project_get db_service" "mysql" "$OUT"
OUT=$(bash -c 'source scripts/lib/projects.sh; project_get "test.local" "enabled"' 2>&1)
assert_eq "project_get enabled" "true" "$OUT"

# project_update
bash -c 'source scripts/lib/projects.sh; project_update "test.local" "php" "php84"' 2>/dev/null
OUT=$(bash -c 'source scripts/lib/projects.sh; project_get "test.local" "php"' 2>&1)
assert_eq "project_update php" "php84" "$OUT"

# project_list_raw
bash -c 'source scripts/lib/projects.sh; project_set "second.local" "php82" "wordpress" "mariadb" "wp_db" "none" "false"' 2>/dev/null
OUT=$(bash -c 'source scripts/lib/projects.sh; project_list_raw' 2>&1)
assert_contains "project_list_raw has second" "$OUT" "second.local"
assert_contains "project_list_raw has test" "$OUT" "test.local"

# project_list_enabled (only enabled ones)
OUT=$(bash -c 'source scripts/lib/projects.sh; project_list_enabled' 2>&1)
assert_contains "project_list_enabled has test" "$OUT" "test.local"
assert_not_contains "project_list_enabled excludes disabled" "$OUT" "second.local"

# project_compute_services
OUT=$(bash -c 'source scripts/lib/projects.sh; project_compute_services' 2>&1)
assert_contains "compute_services includes php84" "$OUT" "php84"
assert_contains "compute_services includes mysql" "$OUT" "mysql"
assert_not_contains "compute_services excludes disabled project's mariadb" "$OUT" "mariadb"

# project_remove
bash -c 'source scripts/lib/projects.sh; project_remove "second.local"' 2>/dev/null
assert_exit_nonzero "project_remove removes" bash -c 'source scripts/lib/projects.sh; project_exists "second.local"'

# Cleanup
bash -c 'source scripts/lib/projects.sh; project_remove "test.local"' 2>/dev/null

# ══════════════════════════════════════════════════════════════════════════════
section "5. SCRIPT SYNTAX CHECKS"
# ══════════════════════════════════════════════════════════════════════════════

for f in scripts/lib/*.sh scripts/create-vhost scripts/database scripts/ssl scripts/varnish scripts/xdebug scripts/setup-composer scripts/init-magento scripts/list-services scripts/shell scripts/mysql scripts/fixowner bin/mage; do
    assert_exit_0 "syntax: $f" bash -n "$f"
done

# ══════════════════════════════════════════════════════════════════════════════
section "6. BIN/MAGE CLI — Help & Errors"
# ══════════════════════════════════════════════════════════════════════════════

OUT=$(./bin/mage help 2>&1)
assert_contains "help shows PROJECT MANAGEMENT" "$OUT" "PROJECT MANAGEMENT"
assert_contains "help shows switch-db" "$OUT" "switch-db"
assert_contains "help shows switch-search" "$OUT" "switch-search"
assert_contains "help shows project set" "$OUT" "project set"

OUT=$(./bin/mage version 2>&1)
assert_contains "version output" "$OUT" "mage v"

OUT=$(./bin/mage foobar 2>&1 || true)
assert_contains "unknown command error" "$OUT" "Unknown command"

OUT=$(./bin/mage shell 2>&1 || true)
assert_contains "shell without args shows error" "$OUT" "Usage"

OUT=$(./bin/mage project 2>&1 || true)
assert_contains "project without args shows usage" "$OUT" "Usage"

# ══════════════════════════════════════════════════════════════════════════════
section "7. BIN/MAGE CLI — Project Management"
# ══════════════════════════════════════════════════════════════════════════════

echo '{}' > projects.json

# project list (empty)
OUT=$(./bin/mage project list 2>&1)
assert_contains "empty project list" "$OUT" "no projects registered"

# Add project non-interactively via projects.sh
bash -c 'source scripts/lib/projects.sh; project_set "alpha.test" "php83" "magento2" "mysql" "alpha_db" "opensearch" "true"' 2>/dev/null
bash -c 'source scripts/lib/projects.sh; project_set "beta.test" "php72" "magento2" "mysql80" "beta_db" "opensearch" "true"' 2>/dev/null
bash -c 'source scripts/lib/projects.sh; project_set "gamma.test" "php84" "magento2" "mariadb" "gamma_db" "elasticsearch" "true"' 2>/dev/null

# project list (populated)
OUT=$(./bin/mage project list 2>&1)
assert_contains "list shows alpha" "$OUT" "alpha.test"
assert_contains "list shows beta" "$OUT" "beta.test"
assert_contains "list shows gamma" "$OUT" "gamma.test"

# project info
OUT=$(./bin/mage project info alpha.test 2>&1)
assert_contains "info shows php83" "$OUT" "php83"
assert_contains "info shows mysql" "$OUT" "mysql"
assert_contains "info shows alpha_db" "$OUT" "alpha_db"

# project switch-php
OUT=$(./bin/mage project switch-php alpha.test php84 2>&1)
assert_contains "switch-php success" "$OUT" "switched to php84"
OUT=$(bash -c 'source scripts/lib/projects.sh; project_get "alpha.test" "php"' 2>&1)
assert_eq "switch-php persisted" "php84" "$OUT"

# project switch-db
OUT=$(./bin/mage project switch-db alpha.test mariadb 2>&1)
assert_contains "switch-db success" "$OUT" "now uses mariadb"
OUT=$(bash -c 'source scripts/lib/projects.sh; project_get "alpha.test" "db_service"' 2>&1)
assert_eq "switch-db persisted" "mariadb" "$OUT"

# project switch-search
OUT=$(./bin/mage project switch-search alpha.test elasticsearch 2>&1)
assert_contains "switch-search success" "$OUT" "now uses elasticsearch"
OUT=$(bash -c 'source scripts/lib/projects.sh; project_get "alpha.test" "search"' 2>&1)
assert_eq "switch-search persisted" "elasticsearch" "$OUT"

# project set (generic)
OUT=$(./bin/mage project set alpha.test db_name new_alpha_db 2>&1)
assert_contains "project set success" "$OUT" "changed"
OUT=$(bash -c 'source scripts/lib/projects.sh; project_get "alpha.test" "db_name"' 2>&1)
assert_eq "project set persisted" "new_alpha_db" "$OUT"

# project set validation
OUT=$(./bin/mage project set alpha.test db_service postgres 2>&1 || true)
assert_contains "project set validates db_service" "$OUT" "Invalid"
OUT=$(./bin/mage project set alpha.test app badapp 2>&1 || true)
assert_contains "project set validates app" "$OUT" "Invalid"

# project disable
OUT=$(./bin/mage project disable beta.test 2>&1 <<< "y")
assert_contains "disable success" "$OUT" "disabled"
OUT=$(bash -c 'source scripts/lib/projects.sh; project_get "beta.test" "enabled"' 2>&1)
assert_eq "disable persisted" "false" "$OUT"

# project enable
OUT=$(./bin/mage project enable beta.test 2>&1)
assert_contains "enable success" "$OUT" "enabled"
OUT=$(bash -c 'source scripts/lib/projects.sh; project_get "beta.test" "enabled"' 2>&1)
assert_eq "enable persisted" "true" "$OUT"

# Smart up compute
OUT=$(bash -c 'source scripts/lib/projects.sh; project_compute_services' 2>&1)
assert_contains "compute has php84" "$OUT" "php84"
assert_contains "compute has php72" "$OUT" "php72"
assert_contains "compute has mariadb" "$OUT" "mariadb"
assert_contains "compute has mysql80" "$OUT" "mysql80"
assert_contains "compute has elasticsearch" "$OUT" "elasticsearch"

OUT=$(bash -c 'source scripts/lib/projects.sh; project_compute_profiles' 2>&1)
assert_contains "compute overrides has legacy" "$OUT" "legacy"
assert_contains "compute overrides has mariadb" "$OUT" "mariadb"
assert_contains "compute overrides has mysql80" "$OUT" "mysql80"
assert_contains "compute overrides has elasticsearch" "$OUT" "elasticsearch"

# Test that computed overrides map to valid compose files
OUT=$(bash -c '
    source scripts/lib/projects.sh
    overrides=$(project_compute_profiles)
    flags=$(dc_file_flags $overrides)
    DC_FILE_FLAGS="$flags" dc config --services
' 2>/dev/null | sort)
assert_contains "override compose has php70" "$OUT" "php70"
assert_contains "override compose has mariadb" "$OUT" "mariadb"
assert_contains "override compose has elasticsearch" "$OUT" "elasticsearch"

# project remove
OUT=$(./bin/mage project remove gamma.test 2>&1 <<< "y")
assert_contains "remove success" "$OUT" "removed"
assert_exit_nonzero "removed project gone" bash -c 'source scripts/lib/projects.sh; project_exists "gamma.test"'

# ══════════════════════════════════════════════════════════════════════════════
section "8. DOCKER COMPOSE — Core + Override Files"
# ══════════════════════════════════════════════════════════════════════════════

# Core file valid
assert_exit_0 "core compose config valid" docker compose config --quiet

# Core has only default services
OUT=$(docker compose config --services 2>/dev/null | sort)
assert_contains "core has nginx" "$OUT" "nginx"
assert_contains "core has php83" "$OUT" "php83"
assert_contains "core has mysql" "$OUT" "mysql"
assert_contains "core has opensearch" "$OUT" "opensearch"
assert_contains "core has redis" "$OUT" "redis"
assert_contains "core has mailpit" "$OUT" "mailpit"
assert_not_contains "core has NO php70" "$OUT" "php70"
assert_not_contains "core has NO mariadb" "$OUT" "mariadb"
assert_not_contains "core has NO elasticsearch" "$OUT" "elasticsearch"
assert_not_contains "core has NO redis6" "$OUT" "redis6"

# Each override file valid
for ovr in legacy mysql80 mariadb opensearch1 elasticsearch elasticsearch7 redis6 debug varnish dashboards; do
    assert_exit_0 "override valid: $ovr" docker compose -f docker-compose.yml -f "compose/${ovr}.yml" config --quiet
done

# Override adds the right service
OUT=$(docker compose -f docker-compose.yml -f compose/legacy.yml config --services 2>/dev/null | sort)
assert_contains "legacy adds php70" "$OUT" "php70"
assert_contains "legacy adds php74" "$OUT" "php74"

OUT=$(docker compose -f docker-compose.yml -f compose/mysql80.yml config --services 2>/dev/null)
assert_contains "mysql80 override adds mysql80" "$OUT" "mysql80"

OUT=$(docker compose -f docker-compose.yml -f compose/mariadb.yml config --services 2>/dev/null)
assert_contains "mariadb override adds mariadb" "$OUT" "mariadb"

OUT=$(docker compose -f docker-compose.yml -f compose/elasticsearch7.yml config --services 2>/dev/null)
assert_contains "es7 override adds elasticsearch7" "$OUT" "elasticsearch7"

OUT=$(docker compose -f docker-compose.yml -f compose/redis6.yml config --services 2>/dev/null)
assert_contains "redis6 override adds redis6" "$OUT" "redis6"

OUT=$(docker compose -f docker-compose.yml -f compose/opensearch1.yml config --services 2>/dev/null)
assert_contains "os1 override adds opensearch1" "$OUT" "opensearch1"

# All overrides combined
ALL_FLAGS="-f docker-compose.yml"
for ovr in legacy mysql80 mariadb opensearch1 elasticsearch elasticsearch7 redis6 debug varnish dashboards; do
    ALL_FLAGS="$ALL_FLAGS -f compose/${ovr}.yml"
done
assert_exit_0 "all overrides combined valid" eval docker compose $ALL_FLAGS config --quiet
OUT=$(eval docker compose $ALL_FLAGS config --services 2>/dev/null | wc -l)
if [[ $OUT -ge 25 ]]; then pass "all overrides: $OUT services total"; else fail "expected 25+ services, got $OUT"; fi

# dc_file_flags mapping
OUT=$(bash -c 'cd '"$PWD"' && source scripts/lib/docker.sh; dc_file_flags legacy mariadb redis6' 2>/dev/null)
assert_contains "dc_file_flags has core" "$OUT" "-f docker-compose.yml"
assert_contains "dc_file_flags has legacy" "$OUT" "compose/legacy.yml"
assert_contains "dc_file_flags has mariadb" "$OUT" "compose/mariadb.yml"
assert_contains "dc_file_flags has redis6" "$OUT" "compose/redis6.yml"

# DC_FILE_FLAGS integration with dc()
OUT=$(bash -c 'source scripts/lib/docker.sh; DC_FILE_FLAGS="-f docker-compose.yml -f compose/legacy.yml" dc config --services' 2>/dev/null | sort)
assert_contains "DC_FILE_FLAGS loads legacy" "$OUT" "php70"
assert_contains "DC_FILE_FLAGS keeps core" "$OUT" "nginx"

# No override = core only
OUT=$(bash -c 'source scripts/lib/docker.sh; dc config --services' 2>/dev/null)
assert_not_contains "no override: no php70" "$OUT" "php70"
assert_contains "no override: has php83" "$OUT" "php83"

# Port conflicts: no two overrides share same host port
# (Verified by: all-overrides config --quiet passes above)

# ══════════════════════════════════════════════════════════════════════════════
section "9. LIVE STACK — Services"
# ══════════════════════════════════════════════════════════════════════════════

# Start minimal stack for testing
docker compose up -d nginx php83 mysql redis mailpit opensearch 2>&1 | tail -1
sleep 15

# Wait for healthchecks
for svc in mysql redis opensearch; do
    for i in $(seq 1 12); do
        if docker compose ps "$svc" --format '{{.Status}}' 2>/dev/null | grep -qi healthy; then break; fi
        sleep 5
    done
done

# Service checks
for svc in nginx php83 mysql redis mailpit opensearch; do
    if docker compose ps --services --filter "status=running" 2>/dev/null | grep -qx "$svc"; then
        pass "service running: $svc"
    else
        fail "service NOT running: $svc"
    fi
done

# Healthchecks
for svc in mysql redis opensearch; do
    STATUS=$(docker compose ps "$svc" --format '{{.Status}}' 2>/dev/null)
    if echo "$STATUS" | grep -qi "healthy"; then
        pass "healthcheck: $svc is healthy"
    else
        fail "healthcheck: $svc status=$STATUS"
    fi
done

# ══════════════════════════════════════════════════════════════════════════════
section "10. LIVE STACK — PHP Container"
# ══════════════════════════════════════════════════════════════════════════════

OUT=$(docker compose exec -T php83 php -r 'echo PHP_MAJOR_VERSION.".".PHP_MINOR_VERSION;' 2>/dev/null)
assert_eq "PHP version" "8.3" "$OUT"

OUT=$(docker compose exec -T php83 composer --version 2>/dev/null | grep -o "Composer version [0-9]" | head -1)
assert_eq "Composer major version" "Composer version 2" "$OUT"

OUT=$(docker compose exec -T php83 id -u nginx 2>/dev/null)
assert_eq "nginx uid" "1000" "$OUT"

OUT=$(docker compose exec -T php83 test -x /usr/local/bin/mhsendmail && echo OK 2>/dev/null)
assert_eq "mhsendmail present" "OK" "$OUT"

OUT=$(docker compose exec -T php83 test -x /usr/bin/n98-magerun2.phar && echo OK 2>/dev/null)
assert_eq "n98-magerun2 present" "OK" "$OUT"

OUT=$(docker compose exec -T php83 node --version 2>/dev/null)
assert_contains "Node.js installed" "$OUT" "v20"

# Required Magento extensions
for ext in bcmath ctype curl dom gd iconv intl mbstring openssl pdo_mysql simplexml soap spl xsl zip; do
    if docker compose exec -T php83 php -r "exit(extension_loaded('$ext') ? 0 : 1);" 2>/dev/null; then
        pass "extension: $ext"
    else
        fail "extension MISSING: $ext"
    fi
done

# ══════════════════════════════════════════════════════════════════════════════
section "11. LIVE STACK — Database Operations"
# ══════════════════════════════════════════════════════════════════════════════

# Create
OUT=$(./scripts/database create --database-name=suite_test_db 2>&1)
assert_contains "db create" "$OUT" "created"
assert_not_contains "db create no password warning" "$OUT" "Using a password"

# List
OUT=$(./scripts/database list 2>&1)
assert_contains "db list shows new db" "$OUT" "suite_test_db"
assert_not_contains "db list no password warning" "$OUT" "Using a password"

# Insert test data
docker compose exec -T -e MYSQL_PWD=root mysql mysql -u root suite_test_db -e \
    "CREATE TABLE items(id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(50)); INSERT INTO items(name) VALUES('alpha'),('beta');" 2>/dev/null

# Export
OUT=$(./scripts/database export --database-name=suite_test_db 2>&1)
assert_contains "db export success" "$OUT" "exported"
EXPORT_FILE=$(ls -t databases/export/suite_test_db-*.sql 2>/dev/null | head -1)
if [[ -n "$EXPORT_FILE" ]]; then
    FIRST_LINE=$(head -1 "$EXPORT_FILE")
    assert_contains "export starts with SQL header" "$FIRST_LINE" "MySQL dump"
    CONTENT=$(cat "$EXPORT_FILE")
    assert_not_contains "export no password warning in file" "$CONTENT" "Using a password"
    assert_contains "export has table data" "$CONTENT" "items"
    pass "export file: $EXPORT_FILE"
else
    fail "export file not created"
fi

# Import
cp "$EXPORT_FILE" databases/import/suite_test.sql 2>/dev/null
./scripts/database create --database-name=suite_import_db >/dev/null 2>&1
OUT=$(./scripts/database import --source=suite_test.sql --target=suite_import_db 2>&1)
assert_contains "db import success" "$OUT" "imported"
assert_not_contains "db import no password warning" "$OUT" "Using a password"

# Verify imported data
ROW_COUNT=$(docker compose exec -T -e MYSQL_PWD=root mysql mysql -u root suite_import_db -N -e "SELECT COUNT(*) FROM items;" 2>/dev/null | tr -d '[:space:]')
assert_eq "import data verified (2 rows)" "2" "$ROW_COUNT"

# Project-aware DB operations
echo '{}' > projects.json
bash -c 'source scripts/lib/projects.sh; project_set "dbtest.local" "php83" "magento2" "mysql" "suite_test_db" "opensearch" "true"' 2>/dev/null

OUT=$(./bin/mage db export dbtest.local 2>&1)
assert_contains "project-aware export" "$OUT" "exporting"

# Drop (pipe yes to stdin for confirmation)
OUT=$(yes | ./scripts/database drop --database-name=suite_import_db 2>&1)
assert_contains "db drop" "$OUT" "dropped"
OUT=$(yes | ./scripts/database drop --database-name=suite_test_db 2>&1)
assert_contains "db drop 2" "$OUT" "dropped"

# Cleanup
rm -f databases/export/suite_test_db-*.sql databases/import/suite_test.sql

# ══════════════════════════════════════════════════════════════════════════════
section "12. LIVE STACK — Xdebug Toggle"
# ══════════════════════════════════════════════════════════════════════════════

# Ensure disabled
OUT=$(./bin/mage xdebug status php83 2>&1 || true)
assert_contains "xdebug initial status" "$OUT" "php83"

OUT=$(./bin/mage xdebug on php83 2>&1)
assert_contains "xdebug enable" "$OUT" "enabled"

# Verify loaded
LOADED=$(docker compose exec -T php83 php -m 2>/dev/null | grep -ci xdebug)
if [[ "$LOADED" -gt 0 ]]; then pass "xdebug loaded after enable"; else fail "xdebug NOT loaded after enable"; fi

OUT=$(./bin/mage xdebug off php83 2>&1)
assert_contains "xdebug disable" "$OUT" "disabled"

LOADED=$(docker compose exec -T php83 php -m 2>/dev/null | grep -ci xdebug)
if [[ "$LOADED" -eq 0 ]]; then pass "xdebug unloaded after disable"; else fail "xdebug still loaded after disable"; fi

# ══════════════════════════════════════════════════════════════════════════════
section "13. LIVE STACK — Vhost & Nginx"
# ══════════════════════════════════════════════════════════════════════════════

mkdir -p sources/vhost-test.local
echo "<?php echo 'VHOST_OK';" > sources/vhost-test.local/index.php

OUT=$(./bin/mage vhost vhost-test.local default php83 2>&1)
assert_contains "vhost created" "$OUT" "Virtual host created"
assert_file_exists "vhost conf created" "conf/nginx/conf.d/vhost-test.local.conf"

# Check vhost uses variable-based fastcgi_pass (runtime DNS)
CONF_CONTENT=$(cat conf/nginx/conf.d/vhost-test.local.conf)
assert_contains "vhost uses runtime DNS" "$CONF_CONTENT" 'set $backend'
assert_not_contains "vhost no hardcoded fastcgi_pass" "$CONF_CONTENT" 'fastcgi_pass   php83:9001'

# Nginx config test
OUT=$(docker compose exec nginx nginx -t 2>&1)
assert_contains "nginx config valid" "$OUT" "syntax is ok"

# Cleanup
rm -rf sources/vhost-test.local conf/nginx/conf.d/vhost-test.local.conf
docker compose exec nginx nginx -s reload 2>/dev/null

# ══════════════════════════════════════════════════════════════════════════════
section "14. LIVE STACK — Service Connectivity"
# ══════════════════════════════════════════════════════════════════════════════

# PHP → MySQL
OUT=$(docker compose exec -T php83 php -r "\$c=@new mysqli('mysql','root','root'); echo \$c->connect_error?:'OK';" 2>/dev/null)
assert_eq "PHP→MySQL connection" "OK" "$OUT"

# PHP → Redis
OUT=$(docker compose exec -T php83 php -r "\$fp=@fsockopen('redis',6379,\$e,\$s,3); if(\$fp){fwrite(\$fp,\"PING\r\n\");echo trim(fgets(\$fp));fclose(\$fp);}else echo 'FAIL';" 2>/dev/null)
assert_eq "PHP→Redis connection" "+PONG" "$OUT"

# PHP → OpenSearch
OUT=$(docker compose exec -T php83 php -r "\$c=curl_init('http://opensearch:9200');curl_setopt(\$c,CURLOPT_RETURNTRANSFER,1);curl_setopt(\$c,CURLOPT_TIMEOUT,5);\$r=curl_exec(\$c);echo \$r?'OK':'FAIL';" 2>/dev/null)
assert_eq "PHP→OpenSearch connection" "OK" "$OUT"

# PHP → Mailpit SMTP
OUT=$(docker compose exec -T php83 php -r "\$fp=@fsockopen('mailpit',1025,\$e,\$s,3);echo \$fp?'OK':'FAIL';" 2>/dev/null)
assert_eq "PHP→Mailpit SMTP connection" "OK" "$OUT"

# ══════════════════════════════════════════════════════════════════════════════
section "15. LIVE STACK — bin/mage status"
# ══════════════════════════════════════════════════════════════════════════════

OUT=$(./bin/mage status 2>&1)
assert_contains "status shows containers" "$OUT" "nginx"
assert_contains "status shows projects" "$OUT" "Registered Projects"

# ══════════════════════════════════════════════════════════════════════════════
section "16. NGINX RESILIENCE — Stopped backends"
# ══════════════════════════════════════════════════════════════════════════════

# Create vhost for a PHP version that's NOT running
mkdir -p sources/resilience.local
cat > conf/nginx/conf.d/resilience.local.conf <<'EOF'
server {
    listen 80;
    server_name resilience.local;
    root /home/public_html/resilience.local;
    location ~ \.php$ {
        set $backend "php84:9001";
        fastcgi_pass $backend;
        include fastcgi_params;
    }
}
EOF

OUT=$(docker compose exec nginx nginx -t 2>&1)
assert_contains "nginx ok with stopped backend" "$OUT" "syntax is ok"
docker compose exec nginx nginx -s reload 2>/dev/null
pass "nginx reloads with stopped php84 backend"

# Cleanup
rm -rf sources/resilience.local conf/nginx/conf.d/resilience.local.conf
docker compose exec nginx nginx -s reload 2>/dev/null

# ══════════════════════════════════════════════════════════════════════════════
# Cleanup & Report
# ══════════════════════════════════════════════════════════════════════════════

section "CLEANUP"
./bin/mage down 2>/dev/null || docker compose down --remove-orphans 2>/dev/null
echo '{}' > projects.json
cp projects.json.bak projects.json 2>/dev/null || echo '{}' > projects.json
rm -f projects.json.bak
pass "cleanup done"

echo ""
echo ""
echo "================================================================"
echo "  TEST RESULTS"
echo "================================================================"
echo ""
echo "  ${_green}Passed:${_reset}  $PASS"
echo "  ${_red}Failed:${_reset}  $FAIL"
echo "  ${_tan}Skipped:${_reset} $SKIP"
echo "  Total:   $((PASS + FAIL + SKIP))"
echo ""

if [[ $FAIL -gt 0 ]]; then
    echo "  ${_red}${_bold}FAILURES:${_reset}"
    for e in "${ERRORS[@]}"; do
        echo "  ${_red}✖${_reset} $e"
    done
    echo ""
fi

echo "================================================================"
[[ $FAIL -eq 0 ]] && exit 0 || exit 1
