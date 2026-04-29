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

#### 1.4 — Extract shared script library
- Create `scripts/lib/common.sh` with all the shared functions (`_arrow`, `_success`, `_error`, `_die`, `_printPoweredBy`, `checkCmdDependencies`, etc.).
- Each script becomes ~50 lines instead of ~300 by sourcing `common.sh`.

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

*Ready to implement. Start with Phase 1?*
