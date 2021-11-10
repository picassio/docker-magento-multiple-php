#!/usr/bin/env bash
#
# Script to create virtual host for Nginx server
#

# UnComment it if bash is lower than 4.x version
shopt -s extglob

################################################################################
# CORE FUNCTIONS - Do not edit
################################################################################

## Uncomment it for debugging purpose
###set -o errexit
#set -o pipefail
#set -o nounset
#set -o xtrace

#
# VARIABLES
#
_bold=$(tput bold)
_underline=$(tput sgr 0 1)
_reset=$(tput sgr0)

_purple=$(tput setaf 171)
_red=$(tput setaf 1)
_green=$(tput setaf 76)
_tan=$(tput setaf 3)
_blue=$(tput setaf 38)

#
# HEADERS & LOGGING
#
function _debug()
{
    if [[ "$DEBUG" = 1 ]]; then
        "$@"
    fi
}

function _header()
{
    printf '\n%s%s==========  %s  ==========%s\n' "$_bold" "$_purple" "$@" "$_reset"
}

function _arrow()
{
    printf '➜ %s\n' "$@"
}

function _success()
{
    printf '%s✔ %s%s\n' "$_green" "$@" "$_reset"
}

function _error() {
    printf '%s✖ %s%s\n' "$_red" "$@" "$_reset"
}

function _warning()
{
    printf '%s➜ %s%s\n' "$_tan" "$@" "$_reset"
}

function _underline()
{
    printf '%s%s%s%s\n' "$_underline" "$_bold" "$@" "$_reset"
}

function _bold()
{
    printf '%s%s%s\n' "$_bold" "$@" "$_reset"
}

function _note()
{
    printf '%s%s%sNote:%s %s%s%s\n' "$_underline" "$_bold" "$_blue" "$_reset" "$_blue" "$@" "$_reset"
}

function _die()
{
    _error "$@"
    exit 1
}

function _safeExit()
{
    exit 0
}

#
# UTILITY HELPER
#
function _seekConfirmation()
{
  printf '\n%s%s%s' "$_bold" "$@" "$_reset"
  read -p " (y/n) " -n 1
  printf '\n'
}

# Test whether the result of an 'ask' is a confirmation
function _isConfirmed()
{
    if [[ "$REPLY" =~ ^[Yy]$ ]]; then
        return 0
    fi
    return 1
}


function _typeExists()
{
    if type "$1" >/dev/null; then
        return 0
    fi
    return 1
}

function _isOs()
{
    if [[ "${OSTYPE}" == $1* ]]; then
      return 0
    fi
    return 1
}

function _isOsDebian()
{
    if [[ -f /etc/debian_version ]]; then
        return 0
    else
        return 1
    fi
}

function _isOsRedHat()
{
    if [[ -f /etc/redhat-release ]]; then
        return 0
    else
        return 1
    fi
}

function _isOsMac()
{
    if [[ "$(uname -s)" = "Darwin" ]]; then
        return 0
    else
        return 1
    fi
}

function _checkRootUser()
{
    #if [ "$(id -u)" != "0" ]; then
    if [ "$(whoami)" != 'root' ]; then
        _die "You cannot run $0 as non-root user. Please use sudo $0"
    fi
}

function _printPoweredBy()
{
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
# SCRIPT FUNCTIONS
################################################################################
function _printUsage()
{
    echo -n "Docker Mysql tools
Version $VERSION

$(basename "$0") [OPTION] [ARG]...

    Options:
        create                    Create database.
        export                    Export database.
        import                    Import database.
        -h, --help                  Display this help and exit.

    Arg:
      create:
        --database-name             Database name need to create.
      export:
        --database-name             Database name need to export.
      import:
        --source                    Name of the database backup file (in database/import folder) use for import.
        --target                    Name of the target database name import to.

    Examples:
      Create database:
        $(basename "$0") create --database-name=database_name
      Export database:
        $(basename "$0") export --database-name=database_name
      Import database:
        $(basename "$0") import --source=database_file --target=database_name
"
    _printPoweredBy
    exit 1
}

function processArgs()
{
    # Parse Arguments

    case $0 in      
        create|import|export)
            COMMAND="$0"
        ;;      
        -h|--help)
            _printUsage
        ;;
        *)
            # _printUsage
            echo "error"
        ;;
    esac

    for arg in "$1"
    do
        case $arg in
            --database-name=*)
                DATABASE_NAME="${arg#*=}"
            ;;
            --source=*)
                DATABASE_IMPORT_SOURCE_NAME="${arg#*=}"
            ;;            
            --target=*)
                DATABASE_IMPORT_TARGET_NAME="${arg#*=}"
            ;;    
            -h|--help)
                _printUsage
            ;;
            *)
                _printUsage
            ;;
        esac
    done

    # validateArgs
}

function validateArgs()
{
    ERROR_COUNT=0
    if [[ -z "$DATABASE_IMPORT_TARGET_NAME" ]]; then
        _error "--import-target=... parameter is missing."
        ERROR_COUNT=$((ERROR_COUNT + 1))
    fi
    if [[ -z "$DATABASE_IMPORT_SOURCE_NAME" ]]; then
        _error "--source=... parameter is missing."
        ERROR_COUNT=$((ERROR_COUNT + 1))
    fi

    [[ "$ERROR_COUNT" -gt 0 ]] && exit 1
}

function getMysqlRootInformation()
{
    containerNameDB=$(docker inspect -f '{{.Name}}' $(docker-compose ps -q mysql) | cut -c2-)

    mysqlUser=$(docker inspect -f '{{range $index, $value := .Config.Env}}{{println $value}}{{end}}'  $containerNameDB | grep MYSQL_USER)
    user="${mysqlUser/MYSQL_USER=/$replace}" 

    mysqlPass=$(docker inspect -f '{{range $index, $value := .Config.Env}}{{println $value}}{{end}}'  $containerNameDB | grep MYSQL_PASSWORD)
    pass="${mysqlPass/MYSQL_PASSWORD=/$replace}" 

    mysqRootPass=$(docker inspect -f '{{range $index, $value := .Config.Env}}{{println $value}}{{end}}'  $containerNameDB | grep MYSQL_ROOT_PASSWORD)
    rootPass="${mysqRootPass/MYSQL_ROOT_PASSWORD=/$replace}" 
}

function exportMysqlDatabase()
{
    echo "Invalid option"
}

function importMysqlDatabase()
{
    echo "Invalid option"
}

function createMysqlDatabase()
{
    echo "Invalid option"
}

function printSuccessMessage()
{
    _success "YOUR Action had done!"

    # echo "################################################################"
    # echo ""
    # echo " >> Domain               : ${VHOST_DOMAIN}"
    # echo " >> Application          : ${APP_TYPE}"
    # echo " >> PHP version          : ${APP_PHP}"
    # echo " >> Document Root        : ${VHOST_ROOT_DIR}"
    # echo " >> Nginx Config File    : ${NGINX_SITES_CONF_D_FILE}"
    # echo ""
    # echo "################################################################"

}

function doAction()
{
    case $COMMAND in      
        create)
            echo "create"
        ;;      
        import)
            echo "import"
        ;;
        export)
            echo "export"
        ;;
    esac
}

################################################################################
# Main
################################################################################
export LC_CTYPE=C
export LANG=C

DEBUG=0
_debug set -x
VERSION="1"

function main()
{
    # _checkRootUser
    # checkCmdDependencies

    # [[ $# -lt 1 ]] && _printUsage

    processArgs "$@"

    doAction
    printSuccessMessage
    exit 0
}

main "$@"

_debug set +x