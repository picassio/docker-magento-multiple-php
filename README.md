# 🐘 Docker Magento Multi-PHP

> Multi-PHP Docker development stack for Magento (and WordPress, Laravel)

**PHP 7.0 – 8.4** · **Magento 2.1 – 2.4.8** · **Linux-focused** · **Docker Compose v2**

---

## Quick Start

```bash
git clone https://github.com/picassio/docker-magento-multiple-php.git ~/docker-magento
cd ~/docker-magento
cp env-example .env

# Check & tune your system (sysctl, THP, Docker log rotation, etc.)
bin/mage doctor fix

# Register your first project (interactive wizard)
bin/mage project add mysite.com
# → Picks: PHP 8.3, magento2, mysql, mysite_com DB, opensearch

# Start — only starts services your projects actually need
bin/mage up

# Everything uses your domain name — no need to remember service names
bin/mage shell mysite.com              # Opens bash in the right PHP container
bin/mage composer mysite.com install   # Runs composer in the right container
bin/mage db import mysite.com dump.sql # Imports into the right DB service
bin/mage magento mysite.com cache:flush
```

---

## Features

- **Project management** — register projects with `bin/mage project add`, switch PHP/DB/search per project, enable/disable without data loss
- **Smart start** — `bin/mage up` reads `projects.json`, starts only the services your projects need, auto-loads the right compose override files
- **Multi-PHP** — PHP 7.0–8.4 simultaneously (PPA for 7.4+, compiled from source for 7.0–7.3)
- **Multi-database** — MySQL 8.4, MySQL 8.0, MariaDB 11.4 on separate ports — all at once
- **Multi-search** — OpenSearch 2.x, OpenSearch 1.x, Elasticsearch 8.x, ES 7.x — all simultaneously
- **Multi-cache** — Redis 7.x and Redis 6.x on separate ports
- **Project-aware commands** — `bin/mage shell mysite.com` auto-resolves PHP version and `cd`'s into the project directory
- **Compose overrides** — clean core compose file (10 services); variants in `compose/*.yml` loaded on demand
- **Runtime DNS** — nginx starts even if a PHP backend is stopped (returns 502, doesn't crash)
- **Zero password warnings** — DB exports are clean SQL, no `Using a password` contamination

---

## Architecture

```
.
├── bin/mage              # CLI — single entrypoint for everything
├── projects.json         # Project registry
├── docker-compose.yml    # Core services (nginx, php81-84, mysql, opensearch, redis, rabbitmq, mailpit)
├── compose/              # Override files loaded on demand
│   ├── legacy.yml        #   PHP 7.0–7.4
│   ├── mysql80.yml       #   MySQL 8.0 (:3307)
│   ├── mariadb.yml       #   MariaDB 11.4 (:3308)
│   ├── opensearch1.yml   #   OpenSearch 1.3 (:9201)
│   ├── elasticsearch.yml #   Elasticsearch 8.x (:9202)
│   ├── elasticsearch7.yml#   Elasticsearch 7.x (:9203)
│   ├── redis6.yml        #   Redis 6.2 (:6380)
│   ├── debug.yml         #   phpMyAdmin + Redis Commander
│   ├── varnish.yml       #   Varnish 7.6
│   └── dashboards.yml    #   OpenSearch Dashboards
├── build/
│   ├── php/              # Dockerfile for PHP 7.4–8.4 (ondrej PPA)
│   └── php-legacy/       # Dockerfile for PHP 7.0–7.3 (compiled from source)
├── conf/                 # Nginx, PHP, MySQL, Varnish configs
├── data/                 # Persistent DB data
├── databases/import|export/
├── logs/nginx/
├── scripts/lib/          # Shared bash library
└── sources/              # Website source code
```

---

## Services & Port Map

### Core (always available)

| Service | Image | Port |
|---|---|---|
| nginx | `nginx:stable-alpine` | 80, 443 |
| php81–php84 | `build/php` | — |
| mysql | `mysql:8.4` | 3306 |
| opensearch | `opensearch:2.19.1` | 9200 |
| redis | `redis:7.4-alpine` | 6379 |
| rabbitmq | `rabbitmq:4.1` | 5672, 15672 |
| mailpit | `axllent/mailpit` | 1025, 8025 |

### Override files (`compose/*.yml` — loaded when projects need them)

| File | Service | Port | Use case |
|---|---|---|---|
| `legacy.yml` | php70–php74 | — | Magento 2.1–2.4.3 |
| `mysql80.yml` | MySQL 8.0 | 3307 | Legacy Magento |
| `mariadb.yml` | MariaDB 11.4 | 3308 | Magento 2.4.8 |
| `opensearch1.yml` | OpenSearch 1.3 | 9201 | Magento 2.4.4–2.4.5 |
| `elasticsearch.yml` | ES 8.17 | 9202 | Alternative to OpenSearch |
| `elasticsearch7.yml` | ES 7.17 | 9203 | Magento 2.3–2.4.5 |
| `redis6.yml` | Redis 6.2 | 6380 | Magento 2.4.0–2.4.5 |
| `debug.yml` | phpMyAdmin, Redis Commander | 8080, 8081 | DB/cache inspection |
| `varnish.yml` | Varnish 7.6 | 6081 | Full-page cache |
| `dashboards.yml` | OpenSearch Dashboards | 5601 | Search analytics |

> All ports are unique — every service variant can run simultaneously.

---

## Magento Compatibility Matrix

| Magento | PHP | Search | Database | Redis | Override files needed |
|---|---|---|---|---|---|
| 2.1–2.2 | 7.0, 7.1 | — | MySQL | 5.x | `legacy.yml` |
| 2.3.x | 7.2–7.4 | ES 7.x | MySQL 8.0 | 6.x | `legacy.yml` `elasticsearch7.yml` `mysql80.yml` `redis6.yml` |
| 2.4.0–2.4.3 | 7.4 | ES 7.x | MySQL 8.0 | 6.x | `legacy.yml` `elasticsearch7.yml` `mysql80.yml` `redis6.yml` |
| 2.4.4–2.4.5 | 8.1 | OS 1.x / ES 7.17 | MySQL 8.0 | 6.x+ | `opensearch1.yml` or `elasticsearch7.yml` `mysql80.yml` |
| 2.4.6 | 8.1, 8.2 | OS 2.x / ES 8.x | MySQL 8.0 | 7.x | *(core only)* |
| 2.4.7 | 8.2, 8.3 | OS 2.x / ES 8.x | MySQL 8.4 | 7.x | *(core only)* |
| 2.4.8 | 8.3, 8.4 | OS 2.x+ | MySQL 8.4 / MariaDB 11.4 | 7.x | `mariadb.yml` *(optional)* |

> `bin/mage up` auto-detects which override files to load from your `projects.json`.

---

## CLI Reference (`bin/mage`)

### Project Management

| Command | Description |
|---|---|
| `project list` | List all projects with status |
| `project add <domain>` | Register a project (interactive wizard) |
| `project remove <domain>` | Remove from registry (keeps source + data) |
| `project enable <domain>` | Enable project, create vhost |
| `project disable <domain>` | Disable project, remove vhost |
| `project info <domain>` | Show config + service status |
| `project switch-php <domain> <php>` | Change PHP version |
| `project switch-db <domain> <db>` | Change DB (mysql/mysql80/mariadb) |
| `project switch-search <domain> <s>` | Change search engine |
| `project set <domain> <field> <val>` | Set any field (php, app, db_service, db_name, search, redis, enabled) |

### Lifecycle

| Command | Description |
|---|---|
| `setup` | Interactive first-time setup |
| `doctor [fix]` | Check/fix system settings (sysctl, THP, Docker logs) |
| `build [php...]` | Build PHP images (or all). Use `--with=legacy` for PHP 7.x, `--no-cache` to rebuild |
| `up [services...]` | Smart start from projects.json (or explicit services) |
| `up --with=<override>` | Start with specific compose override |
| `down` | Stop & remove all containers |
| `stop` / `restart` | Stop or restart services |
| `status` | Show projects + running containers |
| `logs [service]` | Tail container logs |

### Development

| Command | Example |
|---|---|
| `shell <domain\|php>` | `bin/mage shell mysite.com` |
| `composer <domain\|php> [args]` | `bin/mage composer mysite.com install` |
| `magento <domain> [args]` | `bin/mage magento mysite.com cache:flush` |

### Database

| Command | Example |
|---|---|
| `db create <domain\|name>` | `bin/mage db create mysite.com` |
| `db import <domain> <file>` | `bin/mage db import mysite.com backup.sql` |
| `db export <domain\|name>` | `bin/mage db export mysite.com` |
| `db drop <domain\|name>` | `bin/mage db drop mysite.com` |
| `db list [--db-service=...]` | `bin/mage db list --db-service=mariadb` |

### Hosting & Tools

| Command | Example |
|---|---|
| `vhost <domain> <app> <php>` | `bin/mage vhost shop.test magento2 php83` |
| `ssl <domain>` | `bin/mage ssl shop.test` |
| `xdebug <on\|off\|status> <php>` | `bin/mage xdebug on php83` |
| `varnish <on\|off\|status> <domain>` | `bin/mage varnish on shop.test` |
| `install <ver> <ed> <domain> [php]` | `bin/mage install 2.4.7 community shop.test` |

---

## Examples

### Multi-project agency setup

A typical agency running 3 projects on different stacks:

```bash
# Modern Magento 2.4.7 project
bin/mage project add shop.test
# → php83, mysql, opensearch

# Legacy Magento 2.3 project
bin/mage project add legacy.test
# → php72, mysql80, elasticsearch7, redis6
bin/mage project set legacy.test redis redis6

# New Magento 2.4.8 on MariaDB
bin/mage project add new-shop.test
# → php84, mariadb, opensearch

# Start everything — smart start loads the right overrides
bin/mage up
# → Services: nginx php83 php72 php84 mysql mysql80 mariadb opensearch elasticsearch7 redis redis6 mailpit
# → Overrides: legacy mysql80 mariadb elasticsearch7 redis6

# Check what's running
bin/mage status
```

### Working with a project

```bash
# Shell into the right PHP container for your project
bin/mage shell shop.test
# → Opens bash in php83 container, cd's to /home/public_html/shop.test

# Run composer (auto-detects PHP version from project config)
bin/mage composer shop.test install
bin/mage composer shop.test require monolog/monolog

# Run Magento CLI commands
bin/mage magento shop.test cache:flush
bin/mage magento shop.test setup:upgrade
bin/mage magento shop.test indexer:reindex

# Database operations (auto-routes to the right DB service)
bin/mage db create shop.test        # Creates 'shop_test' on mysql
bin/mage db import shop.test bk.sql # Imports into the right DB
bin/mage db export shop.test        # Exports clean SQL (no password warnings)
```

### Switching services for a project

```bash
# Upgrade PHP
bin/mage project switch-php shop.test php84
# → Regenerates nginx vhost automatically

# Switch from MySQL to MariaDB
bin/mage project switch-db shop.test mariadb
# → Gives migration steps: export → create → import → update env.php

# Switch search engine
bin/mage project switch-search shop.test elasticsearch
# → Reminds to update Magento config

# Change any field
bin/mage project set shop.test db_name new_database
bin/mage project set shop.test redis redis6
```

### Enable/disable without data loss

```bash
# Taking a project offline temporarily
bin/mage project disable legacy.test
# → Removes nginx vhost (so nginx doesn't try to route to it)
# → Keeps source code, database, all data intact
# → Next 'bin/mage up' won't start php72/mysql80 if nothing else needs them

# Bring it back
bin/mage project enable legacy.test
# → Recreates nginx vhost
# → 'bin/mage up' will start the required services again
```

### Manual override (without projects.json)

```bash
# Start with specific overrides
bin/mage up --with=legacy --with=debug php72 nginx mysql

# Or use docker compose directly
docker compose -f docker-compose.yml -f compose/legacy.yml -f compose/mariadb.yml up -d

# Debug tools
docker compose -f docker-compose.yml -f compose/debug.yml up -d
# → phpMyAdmin at http://localhost:8080
# → Redis Commander at http://localhost:8081
```

### Fresh Magento install (one command)

```bash
# Magento 2.4.7 — auto-detects: php83 + mysql + opensearch
bin/mage install 2.4.7 community shop.test
# → Registers project in projects.json
# → Starts required services automatically
# → Creates database, vhost, downloads Magento, runs setup:install
# → Prints admin URL + credentials

# Magento 2.4.8 with explicit PHP
bin/mage install 2.4.8 enterprise new-shop.test php84

# Legacy Magento 2.3.7 — auto-detects: php74 + mysql80 + elasticsearch7
bin/mage install 2.3.7 community legacy.test
# → Loads compose/legacy.yml + compose/mysql80.yml + compose/elasticsearch7.yml

# What gets auto-detected per version:
#   2.4.8  → php84 + mysql    + opensearch
#   2.4.7  → php83 + mysql    + opensearch
#   2.4.6  → php82 + mysql    + opensearch
#   2.4.4  → php81 + mysql    + opensearch
#   2.4.0  → php74 + mysql80  + elasticsearch7
#   2.3.x  → php73 + mysql80  + elasticsearch7
#   2.2.x  → php71 + mysql80  + none
```

### SSL & Xdebug

```bash
# SSL
mkcert -install                  # One-time: install local CA
bin/mage ssl shop.test           # Generate cert + HTTPS vhost
# → https://shop.test now works

# Xdebug
bin/mage xdebug on php83         # Enable
bin/mage xdebug status php83     # Check
bin/mage xdebug off php83        # Disable
# PHPStorm: map /home/public_html/shop.test → sources/shop.test
```

### Varnish

```bash
bin/mage up --with=varnish
bin/mage varnish on shop.test     # Enable FPC + proxy
bin/mage varnish status shop.test # Check
bin/mage varnish off shop.test    # Disable
```

### Building images

```bash
# Build all PHP images (core PHP 8.1–8.4)
bin/mage build

# Build specific PHP version
bin/mage build php83

# Build including legacy PHP 7.x
bin/mage build --with=legacy

# Force rebuild without Docker cache
bin/mage build --no-cache
bin/mage build php83 --no-cache

# Pre-built images (nginx, mysql, redis, opensearch, etc.) are pulled
# automatically on 'bin/mage up' — no build needed:
bin/mage build nginx
# → ⚠ 'nginx' uses a pre-built image — nothing to build.
#   Run bin/mage up to pull and start.
```

### System tuning (first-time setup)

```bash
# Check what needs fixing
bin/mage doctor
# ✔ Docker Engine installed: v29.4.0
# ✔ vm.max_map_count = 262144
# ✖ vm.overcommit_memory = 0 (need 1 for Redis)
# ✖ THP enabled (causes Redis latency spikes)
# ✖ Docker log rotation NOT configured
# ✖ vm.swappiness = 60 (recommend ≤ 10)

# Auto-fix everything (persists across reboots)
bin/mage doctor fix
# ✔ vm.overcommit_memory = 1 (fixed, persistent)
# ✔ THP disabled (fixed, systemd service created)
# ✔ Docker log rotation enabled (daemon restarted)
# ✔ vm.swappiness = 10 (fixed, persistent)
```

---

## Compose Override Files

Core services live in `docker-compose.yml`. Non-default services live in `compose/*.yml`:

```bash
# bin/mage up auto-loads needed overrides from projects.json
bin/mage up

# Manual: specify which overrides to load
bin/mage up --with=legacy --with=mariadb
# equivalent to:
docker compose -f docker-compose.yml -f compose/legacy.yml -f compose/mariadb.yml up -d

# All overrides (everything)
docker compose \
  -f docker-compose.yml \
  -f compose/legacy.yml \
  -f compose/mysql80.yml \
  -f compose/mariadb.yml \
  -f compose/opensearch1.yml \
  -f compose/elasticsearch.yml \
  -f compose/elasticsearch7.yml \
  -f compose/redis6.yml \
  -f compose/debug.yml \
  -f compose/varnish.yml \
  up -d
```

### Port map (no conflicts)

```
Service           Port    Service              Port
────────────────  ────    ──────────────────   ────
mysql      8.4   3306    opensearch    2.x    9200
mysql80    8.0   3307    opensearch1   1.3    9201
mariadb   11.4   3308    elasticsearch 8.x    9202
redis      7.4   6379    elasticsearch7 7.x   9203
redis6     6.2   6380
```

---

## Environment Variables

Copy `env-example` to `.env`. Each service variant has its own version + port:

| Variable | Default | Description |
|---|---|---|
| `MYSQL_VERSION` | `8.4` | MySQL (core) |
| `MYSQL80_VERSION` | `8.0` | MySQL 8.0 (override) |
| `MARIADB_VERSION` | `11.4` | MariaDB (override) |
| `MYSQL_PORT` | `3306` | MySQL 8.4 port |
| `MYSQL80_PORT` | `3307` | MySQL 8.0 port |
| `MARIADB_PORT` | `3308` | MariaDB port |
| `OPENSEARCH_VERSION` | `2.19.1` | OpenSearch (core) |
| `OPENSEARCH1_VERSION` | `1.3.19` | OpenSearch 1.x (override) |
| `ELASTICSEARCH_VERSION` | `8.17.0` | ES 8.x (override) |
| `ELASTICSEARCH7_VERSION` | `7.17.27` | ES 7.x (override) |
| `REDIS_VERSION` | `7.4` | Redis (core) |
| `REDIS6_VERSION` | `6.2` | Redis 6.x (override) |
| `REDIS_PORT` | `6379` | Redis 7.x port |
| `REDIS6_PORT` | `6380` | Redis 6.x port |

See `env-example` for the full list including credentials and other ports.

---

## Prerequisites

- **Docker Engine 20.10+** with Compose plugin v2 (`docker compose`)
- **Linux** (Ubuntu 22.04+ recommended) — works on macOS/WSL2 with caveats
- **mkcert** — optional, for local SSL certificates

### System Health Check

Run the doctor command to verify your system is ready:

```bash
bin/mage doctor        # Check system settings
bin/mage doctor fix    # Auto-fix all issues (requires sudo)
```

The doctor checks and fixes:

| Check | Why | Required by |
|---|---|---|
| `vm.max_map_count ≥ 262144` | Memory-mapped files | OpenSearch, Elasticsearch |
| `vm.overcommit_memory = 1` | Prevent BGSAVE failures | Redis |
| Transparent Huge Pages disabled | Prevent latency spikes | Redis |
| `net.core.somaxconn ≥ 65535` | Connection backlog | Redis |
| Docker log rotation | Prevent disk fill | All containers |
| `vm.swappiness ≤ 10` | Reduce swap pressure | Databases |
| Disk space ≥ 20GB | Build + run images | Docker |
| User in `docker` group | No sudo for docker | Docker |

All fixes persist across reboots (written to `/etc/sysctl.conf` + systemd services).

---

## Migrating from v1

| Old | New |
|---|---|
| `php74-c2` | `php74` |
| `php81-c2` | `php81` |
| `mailhog` | `mailpit` (port 8025) |
| `elasticsearch` | `opensearch` (default) |
| `docker-compose up` | `docker compose up` |
| Per-version Dockerfiles | `build/php/` + `build/php-legacy/` |
| All services in one file | Core + `compose/*.yml` overrides |
| Manual service selection | `projects.json` + smart `bin/mage up` |

---

## Tests

```bash
bash tests/test-all.sh        # Full suite: ~160 assertions across 16 sections
bash tests/test-php-images.sh  # PHP image verification: 101 checks × 9 versions
bash tests/test-stacks.sh      # Magento stack connectivity: 6 scenarios
```

---

## WSL / Windows

If `bin/mage` fails with `bash\r: No such file or directory`, line endings are wrong:

```bash
# Fix existing clone (re-normalize line endings)
git rm --cached -r .
git reset --hard
```

For new clones, `.gitattributes` enforces LF automatically.

---

## License

MIT
