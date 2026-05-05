#!/usr/bin/env bash
#
# common.sh — Shared UI, logging, prompts, and path resolution
#
# Usage: source "$(dirname "$0")/lib/common.sh"
#

# Prevent double-sourcing
[[ -n "$_COMMON_SH_LOADED" ]] && return 0
_COMMON_SH_LOADED=1

set -euo pipefail

# ── Colors & formatting ──────────────────────────────────────────────────────
# Graceful fallback if tput is unavailable (e.g. non-interactive / CI)
_bold=$(tput bold 2>/dev/null || echo '')
_underline=$(tput sgr 0 1 2>/dev/null || echo '')
_reset=$(tput sgr0 2>/dev/null || echo '')

_red=$(tput setaf 1 2>/dev/null || echo '')
_green=$(tput setaf 76 2>/dev/null || echo '')
_tan=$(tput setaf 3 2>/dev/null || echo '')
_blue=$(tput setaf 38 2>/dev/null || echo '')
_purple=$(tput setaf 171 2>/dev/null || echo '')

# ── Logging ───────────────────────────────────────────────────────────────────
_arrow()   { printf '➜ %s\n' "$@"; }
_success() { printf '%s✔ %s%s\n' "$_green" "$@" "$_reset"; }
_error()   { printf '%s✖ %s%s\n' "$_red" "$@" "$_reset"; }
_warning() { printf '%s➜ %s%s\n' "$_tan" "$@" "$_reset"; }
_header()  { printf '\n%s%s==========  %s  ==========%s\n' "$_bold" "$_purple" "$@" "$_reset"; }
_note()    { printf '%s%s%sNote:%s %s%s%s\n' "$_underline" "$_bold" "$_blue" "$_reset" "$_blue" "$@" "$_reset"; }
_die()     { _error "$@"; exit 1; }

# ── Debug mode ────────────────────────────────────────────────────────────────
DEBUG=${DEBUG:-0}
_debug() {
    if [[ "$DEBUG" = 1 ]]; then
        "$@"
    fi
}

enable_debug() {
    DEBUG=1
    set -o xtrace
}

# ── Prompts ───────────────────────────────────────────────────────────────────
ask_yes_no() {
    local prompt="$1"
    local reply
    read -rp "${_bold}${prompt} [y/N]:${_reset} " reply
    [[ "$reply" =~ ^[Yy]$ ]]
}

# ── Path resolution ───────────────────────────────────────────────────────────
# Resolve project root from any script location (scripts/ or scripts/lib/)
# Works whether sourced from scripts/foo or scripts/lib/bar.sh
_resolve_root() {
    local dir
    # If sourced from another lib file, BASH_SOURCE[0] is this file
    dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    # Walk up: lib/ -> scripts/ -> root
    if [[ "$(basename "$dir")" == "lib" ]]; then
        echo "$(cd "$dir/../.." && pwd)"
    else
        echo "$(cd "$dir/.." && pwd)"
    fi
}

ROOT_DIR="$(_resolve_root)"
SCRIPTS_DIR="${ROOT_DIR}/scripts"
SOURCE_DIR="${ROOT_DIR}/sources"
CONF_DIR="${ROOT_DIR}/conf"

# ── Root dir validation ───────────────────────────────────────────────────────
# ── Fix ownership (when running as root/sudo) ─────────────────────────────────
fix_ownership() {
    # Match ownership to the project root dir's owner
    if [[ $EUID -eq 0 ]]; then
        local owner
        owner=$(stat -c '%u:%g' "$ROOT_DIR" 2>/dev/null) || return 0
        [[ "$owner" == "0:0" ]] && return 0
        for f in "$@"; do
            [[ -e "$f" ]] && chown "$owner" "$f" 2>/dev/null || true
        done
    fi
}

require_root_dir() {
    if [[ ! -f "${ROOT_DIR}/docker-compose.yml" ]]; then
        _die "Cannot find docker-compose.yml. Please run this command from the project root directory."
    fi
}

# ── Banner ────────────────────────────────────────────────────────────────────
_print_banner() {
    cat <<EOF
${_green}
    ____             __                __  ___                        __
   / __ \____  _____/ /_____  _____   /  |/  /___ _____ ____  ____  / /_____ _
  / / / / __ \/ ___/ //_/ _ \/ ___/  / /|_/ / __ \`/ __ \`/ _ \/ __ \/ __/ __ \`/
 / /_/ / /_/ / /__/ ,< /  __/ /     / /  / / /_/ / /_/ /  __/ / / / /_/ /_/ /
/_____/\____/\___/_/|_|\___/_/     /_/  /_/\__,_/\__, /\___/_/ /_/\__/\__,_/
                                                /____/
${_reset}
EOF
}

# ── Dependency checking ───────────────────────────────────────────────────────
require_commands() {
    local missing=0
    for cmd in "$@"; do
        if ! command -v "$cmd" &>/dev/null; then
            _error "Required command not found: ${cmd}"
            missing=$((missing + 1))
        fi
    done
    if [[ $missing -gt 0 ]]; then
        _die "$missing required command(s) missing."
    fi
}
