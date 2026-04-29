#!/usr/bin/env bash
#
# projects.sh — Project registry management (projects.json)
#
# Usage: source "$(dirname "$0")/lib/projects.sh"
# (automatically sources services.sh → docker.sh → common.sh)
#

[[ -n "$_PROJECTS_SH_LOADED" ]] && return 0
_PROJECTS_SH_LOADED=1

source "$(dirname "${BASH_SOURCE[0]}")/services.sh"

PROJECTS_FILE="${ROOT_DIR}/projects.json"

# ── Ensure projects.json exists ───────────────────────────────────────────────
_ensure_projects_file() {
    if [[ ! -f "$PROJECTS_FILE" ]]; then
        echo '{}' > "$PROJECTS_FILE"
    fi
}

# ── Read a project field ──────────────────────────────────────────────────────
# Usage: project_get <domain> <field>
# Returns empty string if not found
project_get() {
    local domain="$1" field="$2"
    _ensure_projects_file
    python3 -c "
import json, sys
with open('${PROJECTS_FILE}') as f:
    data = json.load(f)
p = data.get('${domain}', {})
v = p.get('${field}', '')
if isinstance(v, bool):
    print('true' if v else 'false')
else:
    print(v)
" 2>/dev/null
}

# ── Check if project exists ───────────────────────────────────────────────────
project_exists() {
    local domain="$1"
    _ensure_projects_file
    python3 -c "
import json
with open('${PROJECTS_FILE}') as f:
    data = json.load(f)
exit(0 if '${domain}' in data else 1)
" 2>/dev/null
}

# ── List all projects ─────────────────────────────────────────────────────────
# Outputs: domain|php|app|db_service|db_name|search|enabled (one per line)
project_list_raw() {
    _ensure_projects_file
    python3 -c "
import json
with open('${PROJECTS_FILE}') as f:
    data = json.load(f)
for domain, p in sorted(data.items()):
    print('|'.join([
        domain,
        p.get('php', ''),
        p.get('app', ''),
        p.get('db_service', 'mysql'),
        p.get('db_name', ''),
        p.get('search', 'opensearch'),
        'true' if p.get('enabled', True) else 'false'
    ]))
" 2>/dev/null
}

# ── Get all enabled projects ──────────────────────────────────────────────────
project_list_enabled() {
    project_list_raw | awk -F'|' '$7 == "true" { print $0 }'
}

# ── Save/update a project ────────────────────────────────────────────────────
# Usage: project_set <domain> <php> <app> <db_service> <db_name> <search> <enabled>
project_set() {
    local domain="$1" php="$2" app="$3" db_service="$4" db_name="$5" search="$6" enabled="$7"
    _ensure_projects_file
    python3 -c "
import json
with open('${PROJECTS_FILE}', 'r') as f:
    data = json.load(f)
data['${domain}'] = {
    'php': '${php}',
    'app': '${app}',
    'db_service': '${db_service}',
    'db_name': '${db_name}',
    'search': '${search}',
    'enabled': $( [[ "$enabled" == "true" ]] && echo "True" || echo "False" )
}
with open('${PROJECTS_FILE}', 'w') as f:
    json.dump(data, f, indent=2)
" 2>/dev/null
}

# ── Update a single field ────────────────────────────────────────────────────
# Usage: project_update <domain> <field> <value>
project_update() {
    local domain="$1" field="$2" value="$3"
    _ensure_projects_file
    python3 -c "
import json
with open('${PROJECTS_FILE}', 'r') as f:
    data = json.load(f)
if '${domain}' in data:
    val = '${value}'
    if val in ('true', 'false'):
        val = val == 'true'
    data['${domain}']['${field}'] = val
    with open('${PROJECTS_FILE}', 'w') as f:
        json.dump(data, f, indent=2)
" 2>/dev/null
}

# ── Remove a project ─────────────────────────────────────────────────────────
project_remove() {
    local domain="$1"
    _ensure_projects_file
    python3 -c "
import json
with open('${PROJECTS_FILE}', 'r') as f:
    data = json.load(f)
data.pop('${domain}', None)
with open('${PROJECTS_FILE}', 'w') as f:
    json.dump(data, f, indent=2)
" 2>/dev/null
}

# ── Collect unique services needed by enabled projects ────────────────────────
# Returns space-separated list of docker compose services to start
project_compute_services() {
    _ensure_projects_file
    local services="nginx redis mailpit"
    local php_versions=""
    local db_services=""
    local search_services=""

    while IFS='|' read -r domain php app db_svc db_name search enabled; do
        [[ -z "$domain" ]] && continue

        # Collect unique PHP versions
        if ! echo "$php_versions" | grep -qw "$php"; then
            php_versions="${php_versions} ${php}"
        fi

        # Collect unique DB services
        if ! echo "$db_services" | grep -qw "$db_svc"; then
            db_services="${db_services} ${db_svc}"
        fi

        # Collect unique search services
        if ! echo "$search_services" | grep -qw "$search"; then
            search_services="${search_services} ${search}"
        fi
    done < <(project_list_enabled)

    echo "${services}${php_versions}${db_services}${search_services}" | tr -s ' '
}

# ── Compute required docker compose profiles ──────────────────────────────────
project_compute_profiles() {
    local profiles=""

    while IFS='|' read -r domain php app db_svc db_name search enabled; do
        [[ -z "$domain" ]] && continue

        # Legacy PHP needs legacy profile
        if echo "$php" | grep -qE "^php7[0-4]$"; then
            if ! echo "$profiles" | grep -qw "legacy"; then
                profiles="${profiles} legacy"
            fi
        fi

        # MariaDB needs mariadb profile
        if [[ "$db_svc" == "mariadb" ]]; then
            if ! echo "$profiles" | grep -qw "mariadb"; then
                profiles="${profiles} mariadb"
            fi
        fi

        # MySQL 8.0 needs mysql80 profile
        if [[ "$db_svc" == "mysql80" ]]; then
            if ! echo "$profiles" | grep -qw "mysql80"; then
                profiles="${profiles} mysql80"
            fi
        fi

        # Elasticsearch needs elasticsearch profile
        if [[ "$search" == "elasticsearch" ]]; then
            if ! echo "$profiles" | grep -qw "elasticsearch"; then
                profiles="${profiles} elasticsearch"
            fi
        fi
    done < <(project_list_enabled)

    echo "$profiles" | tr -s ' '
}

# ── Require that a project exists ─────────────────────────────────────────────
require_project() {
    local domain="$1"
    if ! project_exists "$domain"; then
        _die "Project '${domain}' not found. Run: bin/mage project add ${domain}"
    fi
}

# ── Get DB connection details for a project ───────────────────────────────────
# Sets: PROJECT_DB_SERVICE, PROJECT_DB_NAME, PROJECT_DB_HOST
project_db_info() {
    local domain="$1"
    require_project "$domain"
    PROJECT_DB_SERVICE=$(project_get "$domain" "db_service")
    PROJECT_DB_NAME=$(project_get "$domain" "db_name")
    # The hostname inside docker network is the service name
    PROJECT_DB_HOST="$PROJECT_DB_SERVICE"
}
