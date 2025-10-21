#!/usr/bin/env bash
#
# Common utility functions for docker-magento-multiple-php scripts
# This library provides shared functionality across all management scripts
#

################################################################################
# COLOR AND FORMATTING VARIABLES
################################################################################

# Terminal formatting
_bold=$(tput bold 2>/dev/null || echo '')
_underline=$(tput sgr 0 1 2>/dev/null || echo '')
_reset=$(tput sgr0 2>/dev/null || echo '')

# Color definitions
_purple=$(tput setaf 171 2>/dev/null || echo '')
_red=$(tput setaf 1 2>/dev/null || echo '')
_green=$(tput setaf 76 2>/dev/null || echo '')
_tan=$(tput setaf 3 2>/dev/null || echo '')
_blue=$(tput setaf 38 2>/dev/null || echo '')

################################################################################
# OUTPUT FUNCTIONS
################################################################################

# Execute commands only in debug mode
_debug() {
    if [[ "$DEBUG" = 1 ]]; then
        "$@"
    fi
}

# Print formatted header
_header() {
    printf '\n%s%s==========  %s  ==========%s\n' "$_bold" "$_purple" "$@" "$_reset"
}

# Print arrow indicator
_arrow() {
    printf '➜ %s\n' "$@"
}

# Print success message
_success() {
    printf '%s✔ %s%s\n' "$_green" "$@" "$_reset"
}

# Print error message
_error() {
    printf '%s✖ %s%s\n' "$_red" "$@" "$_reset"
}

# Print warning message
_warning() {
    printf '%s➜ %s%s\n' "$_tan" "$@" "$_reset"
}

# Print underlined text
_underline() {
    printf '%s%s%s%s\n' "$_underline" "$_bold" "$@" "$_reset"
}

# Print bold text
_bold() {
    printf '%s%s%s\n' "$_bold" "$@" "$_reset"
}

# Print note
_note() {
    printf '%s%s%sNote:%s %s%s%s\n' "$_underline" "$_bold" "$_blue" "$_reset" "$_blue" "$@" "$_reset"
}

# Print error message and exit
_die() {
    _error "$@"
    exit 1
}

# Safe exit
_safeExit() {
    exit 0
}

################################################################################
# USER INTERACTION FUNCTIONS
################################################################################

# Seek confirmation from user
_seekConfirmation() {
    printf '\n%s%s%s' "$_bold" "$@" "$_reset"
    read -r -p " (y/n) " -n 1
    printf '\n'
}

# Test whether the result of an 'ask' is a confirmation
_isConfirmed() {
    if [[ "$REPLY" =~ ^[Yy]$ ]]; then
        return 0
    fi
    return 1
}

# Ask yes or no question
askYesOrNo() {
    REPLY=""
    while [ -z "$REPLY" ]; do
        read -r -e -p "$1 $YES_NO_PROMPT" -n1 REPLY
        REPLY=$(echo "${REPLY}" | tr '[:lower:]' '[:upper:]')
        case $REPLY in
            "$YES_CAPS") return 0 ;;
            "$NO_CAPS") return 1 ;;
            *) REPLY="" ;;
        esac
    done
}

################################################################################
# VALIDATION FUNCTIONS
################################################################################

# Check if command/type exists
_typeExists() {
    if type "$1" >/dev/null 2>&1; then
        return 0
    fi
    return 1
}

# Check if running specific OS
_isOs() {
    if [[ "${OSTYPE}" == $1* ]]; then
        return 0
    fi
    return 1
}

# Check if running Debian-based OS
_isOsDebian() {
    if [[ -f /etc/debian_version ]]; then
        return 0
    else
        return 1
    fi
}

# Check if running RedHat-based OS
_isOsRedHat() {
    if [[ -f /etc/redhat-release ]]; then
        return 0
    else
        return 1
    fi
}

# Check if running macOS
_isOsMac() {
    if [[ "$(uname -s)" = "Darwin" ]]; then
        return 0
    else
        return 1
    fi
}

# Check if script is run as root user
_checkRootUser() {
    if [ "$(whoami)" != 'root' ]; then
        _die "You cannot run $0 as non-root user. Please use sudo $0"
    fi
}

# Check if required commands are available
checkCmdDependencies() {
    local _dependencies=(
        wget
        cat
        basename
        mkdir
        cp
        mv
        rm
        chown
        chmod
        date
        find
        awk
        docker-compose
        docker
    )

    for cmd in "${_dependencies[@]}"; do
        hash "${cmd}" &>/dev/null || _die "'${cmd}' command not found."
    done
}

# Check if a command exists
command_exists() {
    type "$1" &>/dev/null
}

################################################################################
# DOCKER FUNCTIONS
################################################################################

# Get list of running services
getRunningServices() {
    docker-compose ps --services --filter "status=running" 2>/dev/null
}

# Check if specific service is running
isServiceRunning() {
    local service_name="$1"
    if [[ $(getRunningServices | grep -w "${service_name}") ]]; then
        return 0
    else
        return 1
    fi
}

# Get MySQL container information
# Sets global variable: rootPass
getMysqlInformation() {
    local containerNameDB
    containerNameDB=$(docker inspect -f '{{.Name}}' "$(docker-compose ps -q mysql)" | cut -c2-)

    local mysqRootPass
    mysqRootPass=$(docker inspect -f '{{range $index, $value := .Config.Env}}{{println $value}}{{end}}' "$containerNameDB" | grep MYSQL_ROOT_PASSWORD)
    rootPass="${mysqRootPass/MYSQL_ROOT_PASSWORD=/}"
}

# Reload Nginx configuration
reloadNginx() {
    local _nginxTest
    _nginxTest=$(docker-compose exec nginx nginx -t 2>&1)
    if [[ $? -eq 0 ]]; then
        docker-compose exec nginx nginx -s reload || _die "Nginx couldn't be reloaded."
    else
        echo "$_nginxTest"
        return 1
    fi
}

################################################################################
# UTILITY FUNCTIONS
################################################################################

# Extract pure domain from URL
getPureDomain() {
    echo "$VHOST_DOMAIN" | awk -F'[:\\/]' '{print $4}'
}

# Initialize common yes/no prompt variables
initYesNoPrompt() {
    YES_STRING=$"y"
    NO_STRING=$"n"
    YES_NO_PROMPT=$"[y/n]: "
    YES_CAPS=$(echo "${YES_STRING}" | tr '[:lower:]' '[:upper:]')
    NO_CAPS=$(echo "${NO_STRING}" | tr '[:lower:]' '[:upper:]')
}

################################################################################
# BRANDING
################################################################################

# Print powered by banner
_printPoweredBy() {
    local mp_ascii
    mp_ascii='
                ____  __  __    _    ____ _____ ___  ____   ____
               / ___||  \/  |  / \  |  _ \_   _/ _ \/ ___| / ___|
               \___ \| |\/| | / _ \ | |_) || || | | \___ \| |
                ___) | |  | |/ ___ \|  _ < | || |_| |___) | |___
               |____/|_|  |_/_/   \_\_| \_\|_| \___/|____/ \____|

'
    cat <<EOF
${_green}
$mp_ascii

################################################################################
${_reset}
EOF
}

################################################################################
# ENVIRONMENT SETUP
################################################################################

# Set up common environment variables
setupCommonEnvironment() {
    export LC_CTYPE=C
    export LANG=C
    DEBUG=${DEBUG:-0}
}

# Enable strict error handling
enableStrictMode() {
    set -o errexit
    set -o pipefail
    set -o nounset
}
