#!/usr/bin/env bash
#
# Test that each Magento version starts with the correct requirement stack.
# Verifies: PHP version, MySQL connectivity, search engine, Redis
#
set -eo pipefail

cd "$(dirname "$0")/.."
source scripts/lib/docker.sh

PASS=0
FAIL=0
SKIP=0
RESULTS=()

# ── Helpers ───────────────────────────────────────────────────────────────────
start_stack() {
    local profile="$1"
    shift
    _arrow "Starting: $*"
    if [[ -n "$profile" ]]; then
        dc --profile "$profile" up -d "$@" 2>&1 | tail -3
    else
        dc up -d "$@" 2>&1 | tail -3
    fi
}

stop_stack() {
    dc --profile legacy --profile elasticsearch down --remove-orphans 2>/dev/null || true
}

wait_healthy() {
    local svc="$1"
    local max_wait="${2:-60}"
    local elapsed=0
    while [[ $elapsed -lt $max_wait ]]; do
        local status
        status=$(dc ps "$svc" --format '{{.Status}}' 2>/dev/null || echo "")
        if echo "$status" | grep -qi "healthy"; then
            return 0
        fi
        # For services without healthcheck, just check running
        if echo "$status" | grep -qi "^Up" && ! echo "$status" | grep -qi "health"; then
            return 0
        fi
        sleep 3
        elapsed=$((elapsed + 3))
    done
    return 1
}

check_php_version() {
    local svc="$1"
    local expected="$2"
    local actual
    actual=$(dc exec -T "$svc" php -r "echo PHP_MAJOR_VERSION.'.'.PHP_MINOR_VERSION;" 2>/dev/null || echo "FAIL")
    if [[ "$actual" == "$expected" ]]; then
        echo "  ✔ PHP version: $actual"
        return 0
    else
        echo "  ✖ PHP version: expected $expected, got $actual"
        return 1
    fi
}

check_mysql() {
    local php_svc="$1"
    local result
    result=$(dc exec -T "$php_svc" php -r "
        \$c = @new mysqli('mysql', 'root', 'root');
        echo \$c->connect_error ? 'FAIL:'.\$c->connect_error : 'OK';
    " 2>/dev/null || echo "FAIL")
    if [[ "$result" == "OK" ]]; then
        echo "  ✔ MySQL connection: OK"
        return 0
    else
        echo "  ✖ MySQL connection: $result"
        return 1
    fi
}

check_redis() {
    local php_svc="$1"
    local result
    result=$(dc exec -T "$php_svc" php -r "
        \$fp = @fsockopen('redis', 6379, \$errno, \$errstr, 3);
        if (\$fp) { fwrite(\$fp, \"PING\r\n\"); echo trim(fgets(\$fp)); fclose(\$fp); } else { echo 'FAIL'; }
    " 2>/dev/null || echo "FAIL")
    if echo "$result" | grep -q "PONG"; then
        echo "  ✔ Redis connection: OK"
        return 0
    else
        echo "  ✖ Redis connection: $result"
        return 1
    fi
}

check_search() {
    local php_svc="$1"
    local search_host="$2"
    local result
    result=$(dc exec -T "$php_svc" php -r "
        \$ch = curl_init('http://${search_host}:9200');
        curl_setopt(\$ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt(\$ch, CURLOPT_TIMEOUT, 5);
        \$r = curl_exec(\$ch);
        echo \$r ? 'OK' : 'FAIL';
    " 2>/dev/null || echo "FAIL")
    if [[ "$result" == "OK" ]]; then
        echo "  ✔ Search ($search_host): OK"
        return 0
    else
        echo "  ✖ Search ($search_host): $result"
        return 1
    fi
}

run_test() {
    local name="$1"
    local php_svc="$2"
    local php_ver="$3"
    local search_svc="${4:-}"
    local extra_checks="${5:-}"

    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  TEST: $name"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    local test_pass=true

    # Check PHP version
    if ! check_php_version "$php_svc" "$php_ver"; then test_pass=false; fi

    # Check MySQL
    if ! check_mysql "$php_svc"; then test_pass=false; fi

    # Check Redis
    if ! check_redis "$php_svc"; then test_pass=false; fi

    # Check search engine if specified
    if [[ -n "$search_svc" ]]; then
        if ! check_search "$php_svc" "$search_svc"; then test_pass=false; fi
    else
        echo "  — Search: not required"
    fi

    # Check Magento-required extensions
    local ext_check
    ext_check=$(dc exec -T "$php_svc" php -r "
        \$required = ['bcmath','curl','gd','intl','mbstring','pdo_mysql','soap','zip'];
        \$missing = array_filter(\$required, function(\$e){ return !extension_loaded(\$e); });
        echo empty(\$missing) ? 'OK' : 'MISSING:'.implode(',',\$missing);
    " 2>/dev/null || echo "FAIL")
    if [[ "$ext_check" == "OK" ]]; then
        echo "  ✔ Extensions: all required present"
    else
        echo "  ✖ Extensions: $ext_check"
        test_pass=false
    fi

    if $test_pass; then
        echo "  ══ PASS ══"
        PASS=$((PASS + 1))
        RESULTS+=("✔ $name")
    else
        echo "  ══ FAIL ══"
        FAIL=$((FAIL + 1))
        RESULTS+=("✖ $name")
    fi
}

# ── Test Scenarios ────────────────────────────────────────────────────────────

echo "================================================================"
echo "  MAGENTO STACK COMPATIBILITY TESTS"
echo "================================================================"

# --- Magento 2.1-2.2: PHP 7.0, MySQL only (no search) ---
_header "Magento 2.1-2.2 Stack"
stop_stack
start_stack "legacy" php70 nginx mysql redis mailpit
wait_healthy mysql 60
sleep 5
run_test "Magento 2.1-2.2 (PHP 7.0 + MySQL)" "php70" "7.0" "" ""

# --- Magento 2.3.x: PHP 7.2, ES/MySQL ---
_header "Magento 2.3.x Stack"
stop_stack
start_stack "legacy" php72 nginx mysql opensearch redis mailpit
wait_healthy mysql 60
wait_healthy opensearch 60
sleep 5
run_test "Magento 2.3.x (PHP 7.2 + MySQL + OpenSearch)" "php72" "7.2" "opensearch" ""

# --- Magento 2.4.0-2.4.3: PHP 7.4, ES/MySQL ---
_header "Magento 2.4.0-2.4.3 Stack"
stop_stack
start_stack "legacy" php74 nginx mysql opensearch redis mailpit
wait_healthy mysql 60
wait_healthy opensearch 60
sleep 5
run_test "Magento 2.4.0-2.4.3 (PHP 7.4 + MySQL + OpenSearch)" "php74" "7.4" "opensearch" ""

# --- Magento 2.4.6: PHP 8.2, OS/MySQL ---
_header "Magento 2.4.6 Stack"
stop_stack
start_stack "" php82 nginx mysql opensearch redis mailpit
wait_healthy mysql 60
wait_healthy opensearch 60
sleep 5
run_test "Magento 2.4.6 (PHP 8.2 + MySQL + OpenSearch)" "php82" "8.2" "opensearch" ""

# --- Magento 2.4.7: PHP 8.3, OS/MySQL 8.4 ---
_header "Magento 2.4.7 Stack"
stop_stack
start_stack "" php83 nginx mysql opensearch redis mailpit
wait_healthy mysql 60
wait_healthy opensearch 60
sleep 5
run_test "Magento 2.4.7 (PHP 8.3 + MySQL 8.4 + OpenSearch)" "php83" "8.3" "opensearch" ""

# --- Magento 2.4.8: PHP 8.4, OS/MySQL 8.4 ---
_header "Magento 2.4.8 Stack"
stop_stack
start_stack "" php84 nginx mysql opensearch redis mailpit
wait_healthy mysql 60
wait_healthy opensearch 60
sleep 5
run_test "Magento 2.4.8 (PHP 8.4 + MySQL 8.4 + OpenSearch)" "php84" "8.4" "opensearch" ""

# ── Cleanup & Report ──────────────────────────────────────────────────────────
stop_stack

echo ""
echo ""
echo "================================================================"
echo "  RESULTS SUMMARY"
echo "================================================================"
for r in "${RESULTS[@]}"; do
    echo "  $r"
done
echo ""
echo "  Passed: $PASS  Failed: $FAIL  Skipped: $SKIP"
echo "================================================================"

[[ $FAIL -eq 0 ]] && exit 0 || exit 1
