# Refactor Plan — docker-magento-multiple-php

> **Branch:** `refactor/modernize-stack`
> **Date:** 2026-04-29
> **Focus:** Linux-first, modern Magento support, easier onboarding

---

## 1. Audit Summary

### What Works Well
- The **multi-PHP concept** is solid — run multiple Magento projects on different PHP versions from one stack.
- Good set of helper scripts (vhost, database, xdebug, ssl, varnish, shell).
- Shared `sources/` volume means you can manage multiple sites easily.
- Already Linux-focused (Ubuntu base images, `/etc/hosts` automation).

### What's Broken / Outdated

| Area | Problem | Severity |
|------|---------|----------|
| **PHP versions** | Only up to 8.2. Missing **PHP 8.3** (Magento 2.4.7) and **PHP 8.4** (Magento 2.4.8). PHP 7.0–7.3 are EOL for years. | 🔴 Critical |
| **Base images** | Most use `ubuntu:focal` (20.04 EOL Apr 2025) or `ubuntu:bionic` (18.04 EOL). | 🔴 Critical |
| **Elasticsearch** | Pinned at 7.14.1. Magento 2.4.7+ needs ES 8.x or **OpenSearch 2.x**. Magento 2.4.8 needs **OpenSearch 3**. | 🔴 Critical |
| **MySQL** | Default 8.0. Magento 2.4.8 supports **MariaDB 11.4/11.5** and MySQL 8.4. | 🟡 Medium |
| **Composer** | Old PHP containers pin Composer 1.10.19. All current Magento needs **Composer 2.7+**. | 🔴 Critical |
| **Node.js** | Pinned to `setup_14.x` (EOL Apr 2023). Should be **Node 20 LTS** or **22 LTS**. | 🟡 Medium |
| **MailHog** | Unmaintained since 2020, no security updates. **Mailpit** is the drop-in replacement. | 🟡 Medium |
| **Xdebug** | Old containers use xdebug 2.x. PHP 8.x needs **Xdebug 3.x**. Config format changed. | 🟡 Medium |
| **docker-compose** | README tells users to install standalone `docker-compose` v1.28.5. Modern Docker ships with `docker compose` v2 plugin. The `docker-compose` command is deprecated. | 🟡 Medium |
| **Dockerfile quality** | Massive duplication — 8 nearly-identical Dockerfiles. No multi-stage, no `.dockerignore`, no layer optimization. Lots of uncacheable `RUN` layers. tini downloaded from GitHub in some, apt-installed in others. | 🟡 Medium |
| **Script duplication** | ~200 lines of identical "CORE FUNCTIONS" copy-pasted into every script. No shared library. | 🟡 Medium |
| **Script hardcoding** | `create-vhost`, `setup-composer`, `xdebug` hardcode allowed PHP versions. Adding a new PHP version means editing 5+ files. | 🟡 Medium |
| **README** | Entirely in Vietnamese. Good for Vietnamese team, but limits adoption. Should be bilingual or English-first with Vietnamese option. | 🟢 Low |
| **No healthchecks** | No Docker healthchecks on any service. Compose doesn't know if containers are actually ready. | 🟢 Low |
| **No profiles** | All services in one flat compose file. New users must know which services to pick. Should use Docker Compose **profiles**. | 🟢 Low |
| **Security** | `nginx` user has `NOPASSWD:ALL` sudo. Good for dev convenience, but could be tightened. Root password in `.env` committed to repo (`.env` and `env-example` are identical with `root`/`admin` passwords). | 🟢 Low |
| **mhsendmail** | Binary downloaded from MailHog GitHub — won't work after MailHog removal. Mailpit uses standard sendmail. | 🟡 Medium |
| **phpredmin** | Unmaintained project (`sasanrose/phpredmin`). Replace with **RedisInsight** or **redis-commander**. | 🟢 Low |
| **Varnish** | Built from `ubuntu:bionic` (EOL). Varnish 6.0 LTS is also EOL. Should use official `varnish:7.x` image. | 🟡 Medium |
| **Nginx** | Custom-built from Ubuntu. Should use official `nginx:stable-alpine` image. | 🟡 Medium |

---

## 2. Refactor Plan — Phased Approach

### Phase 1: Foundation (Do First)
> Clean up the base, remove dead weight, establish patterns.

#### 1.1 — Keep ALL PHP versions, reorganize into tiers
Legacy Magento projects (2.3.x, 2.4.0–2.4.3) still need older PHP. We keep them all but organize into clear tiers:

**Legacy tier** (PHP 7.x — for maintaining old projects):
- `php70` — Magento 2.1–2.2
- `php71` — Magento 2.2–2.3
- `php72` — Magento 2.3.0–2.3.4
- `php73` — Magento 2.3.3–2.3.7
- `php74` — Magento 2.3.7, 2.4.0–2.4.3 (merge `php74` + `php74-c2` into one, Composer 2)

**Current tier** (PHP 8.x — actively supported Magento):
- `php81` — Magento 2.4.4–2.4.6 (renamed from `php81-c2`)
- `php82` — Magento 2.4.6–2.4.7
- `php83` — Magento 2.4.7–2.4.8 **(NEW)**
- `php84` — Magento 2.4.8+ **(NEW)**

All tiers benefit from the unified Dockerfile (Phase 1.2). Legacy containers still get:
- Updated base image (where possible)
- Xdebug version matching their PHP
- Composer 2 (works fine with old Magento, `-c2` suffix removed)
- Proper Mailpit sendmail (replacing mhsendmail)

Legacy PHP containers are placed under a `legacy` Docker Compose profile so they don't build by default — users opt in only when needed.

#### 1.2 — Add modern PHP versions
- **Add** `php83` (for Magento 2.4.7) — recommended
- **Add** `php84` (for Magento 2.4.8) — latest
- All containers use **Composer 2 only** (no more `-c2` suffix naming confusion).

#### 1.3 — Templatize Dockerfiles
- Create a **single base Dockerfile** `build/php/Dockerfile` with `ARG PHP_VERSION`.
- Use build args in `docker-compose.yml`:
  ```yaml
  php83:
    build:
      context: ./build/php
      args:
        PHP_VERSION: "8.3"
  ```
- Eliminates 90% of Dockerfile duplication.
- Base image: `ubuntu:24.04` (noble, supported until 2029).

#### 1.4 — Script refactoring (detailed below in Section 6)

#### 1.5 — Rename services (drop `-c2` suffix)
- `php74-c2` → `php74` (absorbs old `php74`), `php81-c2` → `php81`
- All PHP containers ship with Composer 2. The `-c2` naming was a migration artifact.
- Old `php74` (Composer 1) is deleted — Composer 2 is backward-compatible with Magento 2.3.x/2.4.x.

---

### Phase 2: Service Modernization

#### 2.1 — Replace Elasticsearch with OpenSearch
- Add **OpenSearch 2.x** as default search engine (MIT-licensed, Magento-recommended).
- Keep Elasticsearch 8.x as optional profile for teams that need it.
- Update `init-magento` script to use `--search-engine=opensearch` for Magento ≥2.4.6.
- Add OpenSearch Dashboards as optional replacement for Kibana.

#### 2.2 — Replace MailHog with Mailpit
- Swap `mailhog/mailhog` → `axllent/mailpit:latest`.
- Same SMTP port (1025), web UI moves from 8025 → 8025 (compatible).
- Remove `mhsendmail` binary from PHP Dockerfiles — Mailpit works with PHP's built-in `sendmail` or use `mhsendmail` from Mailpit's own releases.
- Update PHP config: `sendmail_path = "/usr/local/bin/sendmail -S mailpit:1025"` or use Mailpit's mhsendmail.

#### 2.3 — Update MySQL / Add MariaDB option
- Default: **MySQL 8.4** (LTS).
- Add **MariaDB 11.4** as an alternative profile.
- Make it selectable via `.env`:
  ```env
  DB_ENGINE=mysql      # or mariadb
  MYSQL_VERSION=8.4
  MARIADB_VERSION=11.4
  ```

#### 2.4 — Replace custom Nginx build with official image
- Use `nginx:stable-alpine` instead of building from `ubuntu:bionic`.
- Mount config files the same way. Much smaller image, auto-updated.

#### 2.5 — Replace custom Varnish build with official image
- Use `varnish:7.6-alpine` instead of building from `ubuntu:bionic`.
- Mount VCL config the same way.

#### 2.6 — Update Redis
- Bump from `redis:6.0-alpine` to `redis:7.4-alpine`.
- Magento 2.4.7+ recommends Redis 7.x.

#### 2.7 — Replace phpredmin
- Use `redis/redisinsight:latest` or `rediscommander/redis-commander` — both actively maintained.

#### 2.8 — Update RabbitMQ
- Bump to `rabbitmq:4.1-management-alpine` for Magento 2.4.8 compatibility.

---

### Phase 3: Developer Experience (DX)

#### 3.1 — Add a CLI wrapper (`mage` command)
Create a single `bin/mage` entrypoint script that wraps all operations:
```bash
bin/mage setup                    # Interactive first-time setup (copies .env, builds, etc.)
bin/mage up [services...]         # Start services (docker compose up -d)
bin/mage down                     # Stop everything
bin/mage shell <php-version>      # Enter PHP container
bin/mage db create <name>         # Database operations
bin/mage db import <file> <db>
bin/mage db export <db>
bin/mage db drop <db>
bin/mage db list
bin/mage vhost <domain> <app> <php>  # Create virtual host
bin/mage xdebug on|off <php>      # Toggle xdebug
bin/mage ssl <domain>             # Enable SSL
bin/mage varnish on|off <domain>  # Toggle varnish
bin/mage composer <php> [args]    # Run composer in container
bin/mage magento <php> <dir> [args]  # Run bin/magento
bin/mage install <version>        # Download & install fresh Magento
bin/mage status                   # Show running services & sites
bin/mage logs [service]           # Tail logs
```
- Users remember one command instead of 11 separate scripts.
- `bin/mage setup` provides an interactive onboarding wizard.

#### 3.2 — Docker Compose profiles
Group services into profiles so users don't need to know internals:
```yaml
profiles:
  # Core (always started)
  - name: core  # nginx, mysql/mariadb

  # Search
  - name: search  # opensearch + dashboards

  # Cache
  - name: cache  # redis, varnish

  # Queue
  - name: queue  # rabbitmq

  # Mail
  - name: mail   # mailpit

  # Debug
  - name: debug  # phpmyadmin, redisinsight

  # Legacy PHP (7.x — not built by default, opt-in for old Magento projects)
  - name: legacy  # php70, php71, php72, php73, php74
```

#### 3.3 — Docker healthchecks
Add healthchecks to all services:
```yaml
mysql:
  healthcheck:
    test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
    interval: 10s
    timeout: 5s
    retries: 5
```
This lets `depends_on: { condition: service_healthy }` work properly.

#### 3.4 — `.env` restructure
Better organized, well-commented `.env` with sensible defaults:
```env
# =============================================================================
# PHP (which versions to build — uncomment what you need)
# =============================================================================
# PHP_74=true
# PHP_81=true
PHP_82=true
PHP_83=true
# PHP_84=true

# =============================================================================
# Database
# =============================================================================
DB_ENGINE=mysql          # mysql | mariadb
MYSQL_VERSION=8.4
# MARIADB_VERSION=11.4
MYSQL_ROOT_PASSWORD=root
MYSQL_DATABASE=magento
MYSQL_USER=magento
MYSQL_PASSWORD=magento

# =============================================================================
# Search Engine
# =============================================================================
SEARCH_ENGINE=opensearch  # opensearch | elasticsearch
OPENSEARCH_VERSION=2.19.1
# ELASTICSEARCH_VERSION=8.17.0

# =============================================================================
# Services
# =============================================================================
REDIS_VERSION=7.4
RABBITMQ_VERSION=4.1
VARNISH_VERSION=7.6

# =============================================================================
# Ports (change if conflicts with host)
# =============================================================================
HTTP_PORT=80
HTTPS_PORT=443
MYSQL_PORT=3306
PHPMYADMIN_PORT=8080
MAILPIT_PORT=8025
RABBITMQ_MGMT_PORT=15672
OPENSEARCH_PORT=9200
REDIS_PORT=6379
```

#### 3.5 — Quick-start onboarding
Interactive `bin/mage setup` that:
1. Detects if Docker is installed and running.
2. Copies `env-example` → `.env` if not exists.
3. Asks which PHP version(s) the user needs.
4. Asks which services (search, cache, queue, mail).
5. Builds only selected images.
6. Starts the stack.
7. Prints a summary with URLs and credentials.

---

### Phase 4: Cleanup & Documentation

#### 4.1 — Rewrite README in English
- Clear getting-started in 5 steps.
- Architecture diagram (mermaid).
- Compatibility matrix table (Magento version → PHP + services).
- Keep Vietnamese version as `README.vi.md`.

#### 4.2 — Add `.dockerignore`
Prevent sending unnecessary context to Docker daemon:
```
.git
data/
databases/
logs/
sources/
images/
*.md
```

#### 4.3 — Add Magento compatibility matrix
Document which `bin/mage` preset to use:

| Magento Version | PHP | Search | MySQL | Redis | Composer | Profile |
|----------------|-----|--------|-------|-------|----------|----------|
| 2.1–2.2 | 7.0, 7.1 | — | 5.7 | 5.x | 1.x | `legacy` |
| 2.3.0–2.3.4 | 7.2, 7.3 | ES 6.x–7.x | 5.7 / 8.0 | 5.x | 1.x / 2.x | `legacy` |
| 2.3.5–2.3.7 | 7.3, 7.4 | ES 7.x | 8.0 | 5.x–6.x | 1.x / 2.x | `legacy` |
| 2.4.0–2.4.3 | 7.4 | ES 7.x | 8.0 | 6.x | 2.x | `legacy` |
| 2.4.4–2.4.5 | 8.1 | ES 7.17 / OS 1.2 | 8.0 | 6.2+ | 2.x | default |
| 2.4.6 | 8.1, 8.2 | ES 8.x / OS 2.x | 8.0 | 7.0+ | 2.x | default |
| 2.4.7 | 8.2, 8.3 | ES 8.x / OS 2.x | 8.0, 8.4 | 7.0+ | 2.7+ | default |
| 2.4.8 | 8.3, 8.4 | OS 2.x, OS 3 | 8.4 / MariaDB 11.4 | 7.2+ | 2.9+ | default |

#### 4.4 — Add `Makefile` aliases
For users who prefer `make`:
```makefile
up:      bin/mage up
down:    bin/mage down
shell:   bin/mage shell $(PHP)
status:  bin/mage status
```

---

## 3. File Changes Summary

### Delete (consolidated into unified Dockerfile)
```
build/php70/              # Replaced by build/php/Dockerfile ARG
build/php71/              # Replaced by build/php/Dockerfile ARG
build/php72/              # Replaced by build/php/Dockerfile ARG
build/php73/              # Replaced by build/php/Dockerfile ARG
build/php74/              # Merged into php74-c2 → php74
build/php74-c2/           # Replaced by build/php/Dockerfile ARG
build/php81-c2/           # Replaced by build/php/Dockerfile ARG
build/php82/              # Replaced by build/php/Dockerfile ARG
build/nginx/              # Replace with official image
build/varnish/            # Replace with official image
build/elasticsearch/      # Replace with OpenSearch
```

> **Note:** All PHP versions (7.0–8.4) are KEPT as services.
> They are consolidated into a single `build/php/Dockerfile` using build args.
> Legacy PHP (7.x) services are placed under a `legacy` Compose profile.

### Rename
```
conf/php/php74-c2/ → conf/php/php74/
conf/php/php81-c2/ → conf/php/php81/
```

### Create
```
build/php/Dockerfile           # Unified, parameterized PHP Dockerfile
build/php/.dockerignore
bin/mage                       # CLI wrapper
scripts/lib/common.sh          # Shared bash functions
docker-compose.yml             # Rewritten with profiles
env-example                    # Restructured
.dockerignore                  # Root-level
README.md                      # English rewrite
README.vi.md                   # Vietnamese (moved)
Makefile                       # Convenience aliases
conf/php/php83/                # New PHP 8.3 config
conf/php/php84/                # New PHP 8.4 config
```

### Modify
```
scripts/create-vhost           # Source common.sh, auto-detect PHP versions
scripts/database               # Source common.sh, support docker compose v2
scripts/xdebug                 # Source common.sh, support Xdebug 3 config
scripts/ssl                    # Source common.sh
scripts/varnish                # Source common.sh
scripts/shell                  # Support docker compose v2
scripts/mysql                  # Support docker compose v2
scripts/list-services          # Support docker compose v2
scripts/init-magento           # Support OpenSearch, PHP 8.3/8.4, Magento 2.4.7/2.4.8
scripts/setup-composer         # Source common.sh, auto-detect PHP versions
scripts/fixowner               # Source common.sh
```

---

## 4. Implementation Order

```
Phase 1 (Foundation)          ~1 day
  1.1  Remove EOL PHP         ██░░░░░░░░  15 min
  1.2  Add PHP 8.3/8.4        ██░░░░░░░░  30 min
  1.3  Unified Dockerfile     ████░░░░░░  1-2 hr
  1.4  Extract common.sh      ███░░░░░░░  1 hr
  1.5  Rename services        █░░░░░░░░░  10 min

Phase 2 (Services)            ~1 day
  2.1  OpenSearch              ██░░░░░░░░  30 min
  2.2  Mailpit                 █░░░░░░░░░  15 min
  2.3  MySQL/MariaDB           ██░░░░░░░░  30 min
  2.4  Official Nginx          █░░░░░░░░░  15 min
  2.5  Official Varnish        █░░░░░░░░░  15 min
  2.6  Redis 7.x              █░░░░░░░░░  5 min
  2.7  Replace phpredmin       █░░░░░░░░░  10 min
  2.8  RabbitMQ 4.x           █░░░░░░░░░  5 min

Phase 3 (DX)                  ~1-2 days
  3.1  bin/mage CLI            █████░░░░░  2-3 hr
  3.2  Compose profiles        ███░░░░░░░  1 hr
  3.3  Healthchecks            ██░░░░░░░░  30 min
  3.4  .env restructure       ██░░░░░░░░  30 min
  3.5  Setup wizard            ███░░░░░░░  1 hr

Phase 4 (Docs)                ~half day
  4.1  README rewrite          ████░░░░░░  1-2 hr
  4.2  .dockerignore           █░░░░░░░░░  5 min
  4.3  Compat matrix           █░░░░░░░░░  15 min
  4.4  Makefile                █░░░░░░░░░  10 min
```

---

## 5. Migration Guide for Existing Users

1. **Backup** your `sources/`, `data/mysql/`, and any custom nginx configs in `conf/nginx/conf.d/`.
2. `docker compose down -v --remove-orphans` (stop old stack).
3. `git pull` and checkout the new branch.
4. Run `bin/mage setup` — it will migrate your `.env` and rebuild images.
5. Update nginx vhost configs: rename `php74-c2` → `php74`, `php81-c2` → `php81` in your `.conf` files.
6. If you need legacy PHP (7.x): `bin/mage up --profile=legacy php72 nginx mysql` 
7. `bin/mage up` — start the new stack.

> **Legacy users:** All PHP 7.x versions still work. They just require `--profile=legacy` flag
> so they're not built/started by default (saving resources for most users).

---

## 6. Script Refactoring — Detailed Plan

### Current State Analysis

The `scripts/` directory has **3,870 total lines** across 11 scripts:

| Script | Lines | Purpose |
|--------|------:|--------|
| `create-vhost` | 896 | Create nginx virtual hosts (M1, M2, WP, Laravel, default) |
| `database` | 607 | Create/drop/import/export/list MySQL databases |
| `init-magento` | 565 | Download & install fresh Magento |
| `varnish` | 556 | Enable/disable/status Varnish for a domain |
| `xdebug` | 452 | Enable/disable/status Xdebug per PHP version |
| `ssl` | 438 | Generate SSL cert & nginx vhost with mkcert |
| `setup-composer` | 330 | Setup Composer auth for repo.magento.com |
| `fixowner` | 12 | Fix file ownership to 1000:1000 |
| `mysql` | 7 | Open MySQL shell |
| `shell` | 6 | Open bash shell in PHP container |
| `list-services` | 1 | List running services |

### Problem #1: Massive Boilerplate Duplication (~1,300 lines wasted)

**7 scripts** each contain an identical **~189-line "CORE FUNCTIONS" block** with:
- Color variables (`_bold`, `_red`, `_green`, etc.)
- Logging functions (`_arrow`, `_success`, `_error`, `_warning`, `_die`)
- OS detection (`_isOsDebian`, `_isOsRedHat`, `_isOsMac`) — mostly unused
- User helpers (`_seekConfirmation`, `_isConfirmed`, `askYesOrNo`)
- `_printPoweredBy` ASCII art banner
- `checkCmdDependencies` with identical dependency list

That's **~1,300 lines** of identical copy-paste across the codebase.

### Problem #2: Hardcoded PHP Version Lists (6 files to edit)

Every time a new PHP version is added, you must edit hardcoded validation in:
- `create-vhost` line 299: `@(php70|php71|php72|php73|php74|php74-c2|php81-c2|php82)`
- `init-magento` line 316: same pattern
- `setup-composer` line 263: only goes up to `php74-c2`
- `xdebug` line 338: missing `php82`
- Help text in all scripts

They're already out of sync — `setup-composer` doesn't know about `php81-c2` or `php82`, `xdebug` doesn't know about `php82`.

### Problem #3: Duplicated Business Logic (~8 functions repeated)

| Function | Duplicated in |
|----------|---------------|
| `reloadNginx()` | create-vhost, ssl, varnish |
| `checkMysqlContainerRunning()` | database, varnish |
| `checkNginxContainerRunning()` | create-vhost, varnish |
| `getMysqlInformation()` | database, init-magento |
| `checkDomainExist()` | ssl, varnish |
| `checkPhpContainerRunning()` | varnish, xdebug |
| `sanitizeArgs()` / `getPureDomain()` | create-vhost, init-magento, ssl |
| `checkComposerAuth()` | setup-composer, init-magento |

### Problem #4: `docker-compose` (v1) hardcoded everywhere

48 occurrences of `docker-compose` across all scripts. Docker Compose v1 standalone binary is deprecated. Modern Docker uses `docker compose` (plugin). Need to support both with a wrapper function.

---

### Solution: 3-Layer Architecture

```
scripts/
├── lib/
│   ├── common.sh          # Layer 1: UI, colors, logging, prompts
│   ├── docker.sh           # Layer 2: Docker/Compose helpers
│   └── services.sh         # Layer 3: Service-specific shared logic
├── create-vhost            # Slim scripts (~50-150 lines each)
├── database
├── fixowner
├── init-magento
├── list-services
├── mysql
├── setup-composer
├── shell
├── ssl
├── varnish
└── xdebug
```

#### Layer 1: `scripts/lib/common.sh` (~100 lines)
Extract the shared boilerplate **once**:

```bash
#!/usr/bin/env bash
# Shared functions for all scripts

# ── Colors & formatting ──
_bold=$(tput bold 2>/dev/null || echo '')
_reset=$(tput sgr0 2>/dev/null || echo '')
_red=$(tput setaf 1 2>/dev/null || echo '')
_green=$(tput setaf 76 2>/dev/null || echo '')
_tan=$(tput setaf 3 2>/dev/null || echo '')
_blue=$(tput setaf 38 2>/dev/null || echo '')
_purple=$(tput setaf 171 2>/dev/null || echo '')

# ── Logging ──
_arrow()   { printf '➜ %s\n' "$@"; }
_success() { printf '%s✔ %s%s\n' "$_green" "$@" "$_reset"; }
_error()   { printf '%s✖ %s%s\n' "$_red" "$@" "$_reset"; }
_warning() { printf '%s➜ %s%s\n' "$_tan" "$@" "$_reset"; }
_die()     { _error "$@"; exit 1; }

# ── Prompts ──
ask_yes_no() {
    local prompt="$1"
    read -rp "${_bold}${prompt} [y/N]:${_reset} " answer
    [[ "$answer" =~ ^[Yy]$ ]]
}

# ── Resolve project root (where docker-compose.yml lives) ──
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[1]}")/.."; pwd)"
SCRIPTS_DIR="${ROOT_DIR}/scripts"
SOURCE_DIR="${ROOT_DIR}/sources"
CONF_DIR="${ROOT_DIR}/conf"

# ── Banner ──
_print_banner() {
    echo -e "${_green}"
    echo '    ____             __                __  ___                        __        '
    echo '   / __ \____  _____/ /_____  _____   /  |/  /___ _____ ____  ____  / /_____ _ '
    echo '  / / / / __ \/ ___/ //_/ _ \/ ___/  / /|_/ / __ `/ __ `/ _ \/ __ \/ __/ __ `/'
    echo ' / /_/ / /_/ / /__/ ,< /  __/ /     / /  / / /_/ / /_/ /  __/ / / / /_/ /_/ / '
    echo '/_____/\____/\___/_/|_|\___/_/     /_/  /_/\__,_/\__, /\___/_/ /_/\__/\__,_/  '
    echo '                                                /____/                        '
    echo -e "${_reset}"
}
```

#### Layer 2: `scripts/lib/docker.sh` (~80 lines)
Docker & Compose wrappers + PHP version detection:

```bash
#!/usr/bin/env bash
# Docker helpers — sourced by scripts that interact with containers

source "${BASH_SOURCE%/*}/common.sh"

# ── Compose command (v1 fallback → v2) ──
dc() {
    if docker compose version &>/dev/null; then
        docker compose "$@"
    elif command -v docker-compose &>/dev/null; then
        docker-compose "$@"
    else
        _die "Neither 'docker compose' nor 'docker-compose' found. Install Docker."
    fi
}

# ── Auto-detect available PHP versions from docker-compose.yml ──
get_available_php_versions() {
    dc config --services 2>/dev/null | grep '^php' | sort
}

# ── Validate a PHP version against what's actually defined ──
validate_php_version() {
    local php_ver="$1"
    if ! get_available_php_versions | grep -qx "$php_ver"; then
        _error "Invalid PHP version: $php_ver"
        _arrow "Available versions: $(get_available_php_versions | tr '\n' ', ' | sed 's/,$//')"
        exit 1
    fi
}

# ── Check if a service is running ──
require_service() {
    local svc="$1"
    if ! dc ps --services --filter "status=running" 2>/dev/null | grep -qx "$svc"; then
        _die "Service '$svc' is not running. Start it with: $(dc_cmd) up -d $svc"
    fi
}

# ── Reload nginx safely ──
reload_nginx() {
    _arrow "Testing nginx configuration..."
    dc exec nginx nginx -t || _die "Nginx config test failed"
    dc exec nginx nginx -s reload || _die "Nginx reload failed"
    _success "Nginx reloaded"
}

# ── Get MySQL root password from running container ──
get_mysql_root_password() {
    local container
    container=$(docker inspect -f '{{.Name}}' $(dc ps -q mysql) | cut -c2-)
    docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$container" \
        | grep MYSQL_ROOT_PASSWORD | cut -d= -f2-
}
```

#### Layer 3: `scripts/lib/services.sh` (~60 lines)
Shared domain/service logic used by multiple scripts:

```bash
#!/usr/bin/env bash
# Shared service helpers — nginx domains, PHP detection, etc.

source "${BASH_SOURCE%/*}/docker.sh"

NGINX_CONF_DIR="${CONF_DIR}/nginx/conf.d"
NGINX_SSL_DIR="${CONF_DIR}/nginx/ssl"

# ── Sanitize domain (strip protocol, lowercase) ──
sanitize_domain() {
    local domain="$1"
    domain=$(echo "$domain" | sed 's|https\?://||' | awk -F'[:/]' '{print $1}' | tr '[:upper:]' '[:lower:]')
    echo "$domain"
}

# ── Validate domain format ──
validate_domain() {
    local domain="$1"
    if [[ ! "$domain" =~ ^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$ ]]; then
        _die "Invalid domain name: $domain"
    fi
}

# ── Check domain nginx config exists ──
require_domain_config() {
    local domain="$1"
    [[ -f "${NGINX_CONF_DIR}/${domain}.conf" ]] || _die "No nginx config for '$domain'. Create it first with create-vhost."
}

# ── Check domain SSL config exists ──
require_domain_ssl() {
    local domain="$1"
    [[ -f "${NGINX_CONF_DIR}/${domain}-ssl.conf" ]] || _die "No SSL config for '$domain'. Enable SSL first with: ./scripts/ssl --domain=$domain"
}

# ── Get PHP version from an existing nginx vhost config ──
get_domain_php_version() {
    local domain="$1"
    grep ':9001' "${NGINX_CONF_DIR}/${domain}.conf" | awk 'NR==1{print $2}' | cut -d: -f1
}

# ── Get document root from an existing nginx vhost config ──
get_domain_docroot() {
    local domain="$1"
    grep '/home/public_html/' "${NGINX_CONF_DIR}/${domain}.conf" | grep -v fastcgi_param | awk '{print $3}' | awk -F/ '{print $4}' | awk -F\; '{print $1}' | head -1
}

# ── Add /etc/hosts entry ──
add_etc_hosts() {
    local domain="$1"
    if grep -qE "127\.0\.0\.1[[:space:]]+${domain}" /etc/hosts; then
        _warning "$domain already in /etc/hosts"
    else
        _arrow "Adding $domain to /etc/hosts (may require sudo password)"
        echo "127.0.0.1  ${domain}" | sudo tee -a /etc/hosts >/dev/null || _die "Cannot write to /etc/hosts"
        _success "Added to /etc/hosts"
    fi
}
```

---

### Script-by-Script Refactoring

Each script drops from hundreds of lines to just its unique logic:

#### `create-vhost` (896 → ~250 lines)
```bash
#!/usr/bin/env bash
source "$(dirname "$0")/lib/services.sh"

# Only unique code remains:
# - _printUsage (updated help text)
# - processArgs (unchanged arg parsing)
# - validateArgs → uses validate_php_version() from docker.sh
# - vhost templates (prepareM1VhostContent, prepareM2VhostContent, etc.)
# - main flow
#
# REMOVED: ~189 lines CORE FUNCTIONS, reloadNginx, checkAppStackContainerRunning,
#          sanitizeArgs, getPureDomain, checkDomain, createEtcHostEntry
# REPLACED BY: reload_nginx(), require_service(), sanitize_domain(),
#              validate_domain(), add_etc_hosts() from lib/
```

#### `database` (607 → ~200 lines)
```bash
#!/usr/bin/env bash
source "$(dirname "$0")/lib/docker.sh"

# Only unique code remains:
# - _printUsage
# - processArgs (subcommand parsing: create/drop/import/export/list)
# - validateArgs
# - checkDatabaseName, checkDatabaseFileName
# - createMysqlDatabase, dropMysqlDatabase, importMysqlDatabase, exportMysqlDatabase
# - listMysqlDatabase
#
# REMOVED: ~189 lines CORE FUNCTIONS, getMysqlInformation, checkMysqlContainerRunning
# REPLACED BY: get_mysql_root_password(), require_service('mysql') from lib/
```

#### `xdebug` (452 → ~80 lines)
```bash
#!/usr/bin/env bash
source "$(dirname "$0")/lib/docker.sh"

# Only unique code remains:
# - _printUsage
# - processArgs (enable/disable/status)
# - enableXdebug, disableXdebug, statusXdebug
#
# REMOVED: ~189 lines CORE FUNCTIONS, checkPhpContainerRunning, hardcoded PHP list
# REPLACED BY: validate_php_version(), require_service() from lib/
# PHP version list auto-detected — never goes stale
```

#### `ssl` (438 → ~100 lines)
```bash
#!/usr/bin/env bash
source "$(dirname "$0")/lib/services.sh"

# Only unique code remains:
# - _printUsage
# - processArgs
# - checkSslCommand (mkcert check)
# - createCertificate
# - prepareSslVhostContent
#
# REMOVED: ~189 lines CORE FUNCTIONS, checkDomainExist, sanitizeArgs, getPureDomain, reloadNginx
# REPLACED BY: require_domain_config(), sanitize_domain(), reload_nginx() from lib/
```

#### `varnish` (556 → ~150 lines)
```bash
#!/usr/bin/env bash
source "$(dirname "$0")/lib/services.sh"

# Only unique code remains:
# - _printUsage
# - processArgs (enable/disable/status)
# - enableVarnish, disableVarnish, statusVarnish
# - checkDomainMagentoFullPageCacheStatus
# - checkNginxSSLVhostVarnishProxyPass
#
# REMOVED: ~189 lines CORE FUNCTIONS, checkDomainExist, checkDomainSslExist,
#          checkNginxContainerRunning, checkMysqlContainerRunning,
#          checkPhpContainerRunning, getDomainDocroot, getDomainPhpVersion, reloadNginx
# REPLACED BY: require_domain_config(), require_domain_ssl(), require_service(),
#              get_domain_docroot(), get_domain_php_version(), reload_nginx() from lib/
```

#### `init-magento` (565 → ~200 lines)
```bash
#!/usr/bin/env bash
source "$(dirname "$0")/lib/services.sh"

# Only unique code remains:
# - _printUsage
# - processArgs
# - validateMagentoVersion (updated for 2.4.7/2.4.8 + OpenSearch)
# - downloadMagentoVersion, installMagentoVersion
# - printSuccessMessage
#
# REMOVED: ~189 lines CORE FUNCTIONS, getMysqlInformation, validateBaseServices,
#          sanitizeArgs, createSourceCodeFolder, createMagentoDomain, checkComposerAuth
# REPLACED BY: get_mysql_root_password(), require_service(), sanitize_domain() from lib/
#              checkComposerAuth moved to lib/services.sh (shared with setup-composer)
```

#### `setup-composer` (330 → ~40 lines)
```bash
#!/usr/bin/env bash
source "$(dirname "$0")/lib/docker.sh"

# Only unique code remains:
# - _printUsage
# - processArgs
# - main
#
# REMOVED: ~174 lines CORE FUNCTIONS, checkComposerAuth, validateBaseServices
# REPLACED BY: validate_php_version(), require_service(), check_composer_auth() from lib/
```

#### Small scripts (unchanged logic, just modernized)
```bash
# list-services (1 → ~5 lines)
#!/usr/bin/env bash
source "$(dirname "$0")/lib/docker.sh"
dc ps --services --filter "status=running"

# shell (6 → ~8 lines)
#!/usr/bin/env bash
source "$(dirname "$0")/lib/docker.sh"
[[ -z "$1" ]] && _die "Usage: ./scripts/shell <php-version>\nExample: ./scripts/shell php83"
validate_php_version "$1"
require_service "$1"
dc exec --user nginx "$@" bash

# mysql (7 → ~6 lines)
#!/usr/bin/env bash
source "$(dirname "$0")/lib/docker.sh"
require_service mysql
dc exec mysql mysql -uroot -p"$(get_mysql_root_password)"

# fixowner (12 → ~6 lines)
#!/usr/bin/env bash
source "$(dirname "$0")/lib/common.sh"
_arrow "Fixing file ownership in sources/..."
sudo chown -R 1000:1000 "${SOURCE_DIR}/"* || _die "Cannot fix ownership"
_success "Done"
```

---

### Line Count Before vs After

| Component | Before | After | Savings |
|-----------|-------:|------:|--------:|
| CORE FUNCTIONS (×7 scripts) | 1,316 | 0 | **-1,316** |
| `lib/common.sh` | 0 | ~100 | +100 |
| `lib/docker.sh` | 0 | ~80 | +80 |
| `lib/services.sh` | 0 | ~60 | +60 |
| `create-vhost` | 896 | ~250 | -646 |
| `database` | 607 | ~200 | -407 |
| `init-magento` | 565 | ~200 | -365 |
| `varnish` | 556 | ~150 | -406 |
| `xdebug` | 452 | ~80 | -372 |
| `ssl` | 438 | ~100 | -338 |
| `setup-composer` | 330 | ~40 | -290 |
| Small scripts (4) | 26 | ~25 | ~same |
| **TOTAL** | **3,870** | **~1,285** | **-67%** |

### Key Design Decisions

1. **Auto-detect PHP versions** — `get_available_php_versions()` reads from `docker-compose.yml` instead of hardcoded lists. Adding `php85` later = just add it to compose, scripts work automatically.

2. **`dc()` wrapper** — Supports both `docker compose` (v2) and `docker-compose` (v1 fallback). Migrate once, forget forever.

3. **Source chain**: `common.sh` ← `docker.sh` ← `services.sh`. Scripts source the deepest layer they need:
   - `fixowner` needs only `common.sh` (no Docker interaction)
   - `xdebug`, `database` need `docker.sh` (container operations)
   - `create-vhost`, `ssl`, `varnish` need `services.sh` (nginx/domain logic)

4. **Backward-compatible CLI** — All existing command signatures (`./scripts/xdebug enable --php-version=php72`) stay the same. Users don't need to relearn anything.

5. **Each script keeps its own `_printUsage`** — Help text is unique to each script, so it stays inline rather than trying to genericize it.

---

*Ready to implement. Start with Phase 1?*
