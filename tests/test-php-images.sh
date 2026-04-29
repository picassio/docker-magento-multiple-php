#!/usr/bin/env bash
#
# Verify all PHP images: version, extensions, tools, FPM, user, sendmail
# Runs against built images without starting the full stack.
#
set -eo pipefail

IMAGE_PREFIX="docker-magento-multiple-php"
PASS=0
FAIL=0
TOTAL_CHECKS=0

# ── Color helpers ─────────────────────────────────────────────────────────────
_green=$(tput setaf 76 2>/dev/null || echo '')
_red=$(tput setaf 1 2>/dev/null || echo '')
_tan=$(tput setaf 3 2>/dev/null || echo '')
_bold=$(tput bold 2>/dev/null || echo '')
_reset=$(tput sgr0 2>/dev/null || echo '')

ok()   { TOTAL_CHECKS=$((TOTAL_CHECKS+1)); PASS=$((PASS+1)); echo "  ${_green}✔${_reset} $1"; }
fail() { TOTAL_CHECKS=$((TOTAL_CHECKS+1)); FAIL=$((FAIL+1)); echo "  ${_red}✖${_reset} $1"; }
warn() { echo "  ${_tan}⚠${_reset} $1"; }

run_in() {
    local img="$1"; shift
    docker run --rm "${IMAGE_PREFIX}-${img}" "$@" 2>/dev/null
}

# ── Magento required extensions per version ───────────────────────────────────
# https://experienceleague.adobe.com/docs/commerce-operations/installation-guide/prerequisites/php-settings.html

# Core extensions required by ALL Magento 2.x versions
CORE_EXTS="bcmath ctype curl dom gd iconv intl mbstring openssl pdo_mysql simplexml soap spl xsl zip libxml"

# Additional extensions by Magento version range
# Magento 2.1-2.2 (PHP 7.0-7.1): core + mcrypt + json
# Magento 2.3.x   (PHP 7.2-7.3): core + json + sodium(optional)
# Magento 2.4.0-3  (PHP 7.4):    core + json + sodium + sockets
# Magento 2.4.4+  (PHP 8.1+):    core + sodium + sockets (no json, built-in)

declare -A PHP_MAGENTO_MAP
PHP_MAGENTO_MAP[php70]="2.1-2.2"
PHP_MAGENTO_MAP[php71]="2.2-2.3"
PHP_MAGENTO_MAP[php72]="2.3.0-2.3.4"
PHP_MAGENTO_MAP[php73]="2.3.3-2.3.7"
PHP_MAGENTO_MAP[php74]="2.4.0-2.4.3"
PHP_MAGENTO_MAP[php81]="2.4.4-2.4.6"
PHP_MAGENTO_MAP[php82]="2.4.6-2.4.7"
PHP_MAGENTO_MAP[php83]="2.4.7-2.4.8"
PHP_MAGENTO_MAP[php84]="2.4.8+"

declare -A PHP_EXTRA_EXTS
PHP_EXTRA_EXTS[php70]="mcrypt json"
PHP_EXTRA_EXTS[php71]="mcrypt json"
PHP_EXTRA_EXTS[php72]="json"
PHP_EXTRA_EXTS[php73]="json"
PHP_EXTRA_EXTS[php74]="json sodium sockets"
PHP_EXTRA_EXTS[php81]="sodium sockets"
PHP_EXTRA_EXTS[php82]="sodium sockets"
PHP_EXTRA_EXTS[php83]="sodium sockets"
PHP_EXTRA_EXTS[php84]="sodium sockets"

declare -A PHP_EXPECTED_VERSION
PHP_EXPECTED_VERSION[php70]="7.0"
PHP_EXPECTED_VERSION[php71]="7.1"
PHP_EXPECTED_VERSION[php72]="7.2"
PHP_EXPECTED_VERSION[php73]="7.3"
PHP_EXPECTED_VERSION[php74]="7.4"
PHP_EXPECTED_VERSION[php81]="8.1"
PHP_EXPECTED_VERSION[php82]="8.2"
PHP_EXPECTED_VERSION[php83]="8.3"
PHP_EXPECTED_VERSION[php84]="8.4"

# Magento Composer requirements:
# 2.1-2.3.x (php70-73): Composer 1 only (plugins break on 2)
# 2.4.0-2.4.1 (php74):  Composer 1 recommended, 2 works
# 2.4.2+ (php74):       Composer 2 supported
# 2.4.4+ (php81+):      Composer 2 required
declare -A PHP_COMPOSER_VERSION
PHP_COMPOSER_VERSION[php70]="1"
PHP_COMPOSER_VERSION[php71]="1"
PHP_COMPOSER_VERSION[php72]="1"
PHP_COMPOSER_VERSION[php73]="1"
PHP_COMPOSER_VERSION[php74]="2"
PHP_COMPOSER_VERSION[php81]="2"
PHP_COMPOSER_VERSION[php82]="2"
PHP_COMPOSER_VERSION[php83]="2"
PHP_COMPOSER_VERSION[php84]="2"

# Xdebug version mapping:
# php70: xdebug 2.7.2 (last for 7.0)
# php71-74: xdebug 2.9.8 (last 2.x)
# php81: xdebug 3.4.7
# php82-83: xdebug 3.4.7
# php84: xdebug 3.5.1 (3.4.x doesn't support 8.4)
declare -A PHP_XDEBUG_MAJOR
PHP_XDEBUG_MAJOR[php70]="2"
PHP_XDEBUG_MAJOR[php71]="2"
PHP_XDEBUG_MAJOR[php72]="2"
PHP_XDEBUG_MAJOR[php73]="2"
PHP_XDEBUG_MAJOR[php74]="2"
PHP_XDEBUG_MAJOR[php81]="3"
PHP_XDEBUG_MAJOR[php82]="3"
PHP_XDEBUG_MAJOR[php83]="3"
PHP_XDEBUG_MAJOR[php84]="3"

# ── Run tests ─────────────────────────────────────────────────────────────────

echo "================================================================"
echo "  PHP IMAGE VERIFICATION — All Magento Requirements"
echo "================================================================"

RESULTS=()

for svc in php70 php71 php72 php73 php74 php81 php82 php83 php84; do
    img="${IMAGE_PREFIX}-${svc}"
    expected_ver="${PHP_EXPECTED_VERSION[$svc]}"
    magento_ver="${PHP_MAGENTO_MAP[$svc]}"
    local_fail=0

    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  ${_bold}${svc}${_reset} — Magento ${magento_ver} (expects PHP ${expected_ver})"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # 1. PHP version
    actual_ver=$(run_in "$svc" php -r "echo PHP_MAJOR_VERSION.'.'.PHP_MINOR_VERSION;" || echo "FAIL")
    if [[ "$actual_ver" == "$expected_ver" ]]; then
        actual_full=$(run_in "$svc" php -r "echo PHP_VERSION;")
        ok "PHP version: ${actual_full}"
    else
        fail "PHP version: expected ${expected_ver}, got ${actual_ver}"
        local_fail=$((local_fail+1))
    fi

    # 2. Core Magento extensions
    missing_core=""
    for ext in $CORE_EXTS; do
        if ! run_in "$svc" php -r "if(!extension_loaded('${ext}')) exit(1);" 2>/dev/null; then
            missing_core="${missing_core} ${ext}"
        fi
    done
    if [[ -z "$missing_core" ]]; then
        ok "Core extensions: all ${#CORE_EXTS[@]} present (${CORE_EXTS// /, })"
    else
        fail "Core extensions MISSING:${missing_core}"
        local_fail=$((local_fail+1))
    fi

    # 3. Version-specific extensions
    extra_exts="${PHP_EXTRA_EXTS[$svc]}"
    missing_extra=""
    for ext in $extra_exts; do
        if ! run_in "$svc" php -r "if(!extension_loaded('${ext}')) exit(1);" 2>/dev/null; then
            missing_extra="${missing_extra} ${ext}"
        fi
    done
    if [[ -z "$missing_extra" ]]; then
        ok "Extra extensions: ${extra_exts// /, }"
    else
        fail "Extra extensions MISSING:${missing_extra} (needed: ${extra_exts})"
        local_fail=$((local_fail+1))
    fi

    # 4. Composer version
    comp_ver=$(run_in "$svc" composer --version 2>/dev/null | grep -o "[0-9]*\.[0-9]*\.[0-9]*" | head -1 || echo "NONE")
    comp_major="${comp_ver%%.*}"
    expected_comp="${PHP_COMPOSER_VERSION[$svc]}"
    if [[ "$comp_major" == "$expected_comp" ]]; then
        ok "Composer: v${comp_ver} (major ${comp_major})"
    else
        fail "Composer: v${comp_ver} — expected major ${expected_comp}"
        local_fail=$((local_fail+1))
    fi

    # 5. n98-magerun2
    if run_in "$svc" test -x /usr/bin/n98-magerun2.phar; then
        ok "n98-magerun2: present"
    else
        fail "n98-magerun2: missing"
        local_fail=$((local_fail+1))
    fi

    # 6. Node.js
    node_ver=$(run_in "$svc" node --version 2>/dev/null || echo "NONE")
    if [[ "$node_ver" != "NONE" ]]; then
        ok "Node.js: ${node_ver}"
    else
        warn "Node.js: not available (non-fatal for legacy)"
    fi

    # 7. nginx user (uid 1000)
    uid=$(run_in "$svc" id -u nginx 2>/dev/null || echo "NONE")
    gid=$(run_in "$svc" id -g nginx 2>/dev/null || echo "NONE")
    if [[ "$uid" == "1000" && "$gid" == "1000" ]]; then
        ok "nginx user: uid=${uid} gid=${gid}"
    else
        fail "nginx user: uid=${uid} gid=${gid} (expected 1000:1000)"
        local_fail=$((local_fail+1))
    fi

    # 8. mhsendmail
    if run_in "$svc" test -x /usr/local/bin/mhsendmail; then
        ok "mhsendmail: present"
    else
        fail "mhsendmail: missing"
        local_fail=$((local_fail+1))
    fi

    # 9. Xdebug installed (but disabled by default)
    xdebug_file=$(run_in "$svc" find /etc/php/ -name "*xdebug*" -type f 2>/dev/null | head -1 || echo "")
    xdebug_active=$(run_in "$svc" php -m 2>/dev/null | grep -ci xdebug || echo "0")
    expected_xd_major="${PHP_XDEBUG_MAJOR[$svc]}"
    
    if [[ -n "$xdebug_file" ]]; then
        xd_ver=$(run_in "$svc" php -r "echo phpversion('xdebug') ?: 'not loaded';" 2>/dev/null || echo "installed")
        if [[ "$xdebug_active" == "0" ]]; then
            ok "Xdebug: installed (${xdebug_file}), disabled by default ✓"
        else
            warn "Xdebug: installed and ENABLED (should be disabled by default)"
        fi
    else
        # Xdebug available via pecl but might not be in a file — check pecl
        if run_in "$svc" pecl list 2>/dev/null | grep -qi xdebug; then
            ok "Xdebug: installed via pecl"
        else
            fail "Xdebug: not installed"
            local_fail=$((local_fail+1))
        fi
    fi

    # 10. PHP-FPM binary exists and is executable
    fpm_path="/usr/sbin/php-fpm${expected_ver}"
    if run_in "$svc" test -x "${fpm_path}"; then
        ok "PHP-FPM binary: ${fpm_path}"
    else
        # Check unversioned path
        if run_in "$svc" test -x /usr/sbin/php-fpm; then
            ok "PHP-FPM binary: /usr/sbin/php-fpm (unversioned)"
        else
            fail "PHP-FPM binary: not found at ${fpm_path}"
            local_fail=$((local_fail+1))
        fi
    fi

    # 11. Key directories
    dirs_ok=true
    for d in /home/public_html /home/nginx/.composer /var/lib/php/session; do
        if ! run_in "$svc" test -d "$d"; then
            fail "Directory missing: $d"
            dirs_ok=false
            local_fail=$((local_fail+1))
        fi
    done
    if $dirs_ok; then
        ok "Directories: public_html, .composer, session"
    fi

    # 12. sudo for nginx user
    if run_in "$svc" test -f /etc/sudoers.d/nginx; then
        ok "sudo: nginx NOPASSWD configured"
    else
        fail "sudo: /etc/sudoers.d/nginx missing"
        local_fail=$((local_fail+1))
    fi

    # Summary for this version
    if [[ $local_fail -eq 0 ]]; then
        RESULTS+=("${_green}✔${_reset} ${svc} (PHP ${actual_ver:-?}) — Magento ${magento_ver}: ALL CHECKS PASS")
    else
        RESULTS+=("${_red}✖${_reset} ${svc} (PHP ${actual_ver:-?}) — Magento ${magento_ver}: ${local_fail} FAILURE(S)")
    fi
done

# ── Report ────────────────────────────────────────────────────────────────────
echo ""
echo ""
echo "================================================================"
echo "  RESULTS SUMMARY"
echo "================================================================"
for r in "${RESULTS[@]}"; do
    echo "  $r"
done
echo ""
echo "  Total checks: ${TOTAL_CHECKS}  Passed: ${PASS}  Failed: ${FAIL}"
echo "================================================================"

[[ $FAIL -eq 0 ]] && exit 0 || exit 1
