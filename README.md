# 🐘 Docker Magento Multi-PHP

> Multi-PHP Docker development stack for Magento (and WordPress, Laravel)

**PHP 7.0 – 8.4** · **Magento 2.1 – 2.4.8** · **Linux-focused** · **Docker Compose v2**

---

## Quick Start

```bash
git clone https://github.com/picassio/docker-magento-multiple-php.git ~/docker-magento
cd ~/docker-magento
cp env-example .env

# Register your project
bin/mage project add mysite.com    # Interactive wizard: pick PHP, DB, search

# Start — automatically starts only what your project needs
bin/mage up

# Work with your project by domain name
bin/mage shell mysite.com           # PHP shell
bin/mage composer mysite.com install # Run composer
bin/mage db import mysite.com bk.sql # Import database
```

---

## Features

- **Project management** — register projects, enable/disable, switch PHP versions. `bin/mage up` starts only what your projects need.
- **Multi-PHP** — PHP 7.0–8.4 (PPA for 7.4+, compiled from source for 7.0–7.3)
- **Multi-database** — MySQL 8.4, MySQL 8.0, MariaDB 11.4 running simultaneously on different ports
- **Project-aware commands** — `bin/mage shell mysite.com` auto-resolves PHP version + directory
- **Auto virtual hosts** — Nginx vhosts with runtime DNS (nginx won’t crash if a PHP container is stopped)
- **SSL, Xdebug, Varnish** — one-command toggle
- **Database management** — project-aware create, drop, import, export — auto-routes to the right DB service
- **Email catching** — Mailpit (web UI on port 8025)
- **Search engines** — OpenSearch 2.x (default) or Elasticsearch 8.x
- **Redis, RabbitMQ** — with optional debug UIs via profiles

---

## Architecture

```
.
├── bin/mage              # CLI wrapper — single entrypoint for everything
├── projects.json         # Project registry (which sites need which services)
├── build/
│   ├── php/              # Dockerfile for PHP 7.4–8.4 (ondrej PPA)
│   └── php-legacy/       # Dockerfile for PHP 7.0–7.3 (compiled from source)
├── conf/
│   ├── nginx/            # Nginx configs, SSL certs
│   ├── php/php70-php84/  # Per-version PHP configs
│   ├── mysql/            # MySQL config
│   └── varnish/          # Varnish VCL
├── data/                 # Persistent data (MySQL, etc.)
├── databases/
│   ├── import/           # Drop .sql/.sql.gz here to import
│   └── export/           # Exported dumps land here
├── logs/nginx/           # Nginx access/error logs
├── scripts/              # Individual command scripts
│   └── lib/              # Shared bash library
├── sources/              # Website source code (mount point)
├── docker-compose.yml    # Service definitions
└── env-example           # Template for .env
```

---

## Services

| Service | Image | Default | Profile | Ports |
|---|---|:---:|---|---|
| **nginx** | `nginx:stable-alpine` | ✅ | — | 80, 443 |
| **php81 – php84** | `build/php` | ✅ | — | — |
| **php70 – php74** | `build/php` | — | `legacy` | — |
| **mysql** | `mysql:8.4` | ✅ | — | 3306 |
| **mysql80** | `mysql:8.0` | — | `mysql80` | 3307 |
| **mariadb** | `mariadb:11.4` | — | `mariadb` | 3308 |
| **opensearch** | `opensearchproject/opensearch:2.19.1` | ✅ | — | 9200 |
| **elasticsearch** | `elasticsearch:8.17` | — | `elasticsearch` | 9200 |
| **redis** | `redis:7.4-alpine` | ✅ | — | 6379 |
| **rabbitmq** | `rabbitmq:4.1-management` | ✅ | — | 5672, 15672 |
| **mailpit** | `axllent/mailpit:latest` | ✅ | — | 1025, 8025 |
| **varnish** | `varnish:7.6-alpine` | — | `varnish` | 6081 |
| **phpmyadmin** | `phpmyadmin:latest` | — | `debug` | 8080 |
| **redis-commander** | `rediscommander/redis-commander` | — | `debug` | 8081 |
| **opensearch-dashboards** | `opensearchproject/opensearch-dashboards` | — | `dashboards` | 5601 |

---

## Magento Compatibility Matrix

| Magento | PHP | Search | Database | Profile |
|---|---|---|---|---|
| 2.1 – 2.2 | 7.0, 7.1 | — | MySQL 5.7 | `legacy` |
| 2.3.x | 7.2 – 7.4 | ES 7.x | MySQL 8.0 | `legacy` |
| 2.4.0 – 2.4.3 | 7.4 | ES 7.x | MySQL 8.0 | `legacy` |
| 2.4.4 – 2.4.5 | 8.1 | ES 7.17 / OS 1.2 | MySQL 8.0 | default |
| 2.4.6 | 8.1, 8.2 | ES 8.x / OS 2.x | MySQL 8.0 | default |
| 2.4.7 | 8.2, 8.3 | ES 8.x / OS 2.x | MySQL 8.4 | default |
| 2.4.8 | 8.3, 8.4 | OS 2.x / 3 | MySQL 8.4 / MariaDB 11.4 | default |

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
| `project switch-search <domain> <s>` | Change search (opensearch/elasticsearch/none) |
| `project set <domain> <field> <val>` | Set any project field |

### Lifecycle

| Command | Description |
|---|---|
| `up [services...]` | Smart start from projects.json (or manual) |
| `down` | Stop & remove all containers |
| `status` | Show projects, services, domains |
| `logs [service]` | Tail logs |

### Development (accepts domain — auto-resolves PHP)

| Command | Example |
|---|---|
| `shell <domain\|php>` | `bin/mage shell mysite.com` |
| `composer <domain\|php> [args]` | `bin/mage composer mysite.com install` |
| `magento <domain> [args]` | `bin/mage magento mysite.com cache:flush` |

### Database (project-aware — auto-routes to correct DB service)

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
| `install <ver> <ed> <domain> <php>` | `bin/mage install 2.4.7 community shop.test php83` |

---

## Common Workflows

### Setting Up an Existing Magento Project

```bash
# 1. Register the project
bin/mage project add mysite.com
# → Wizard asks: PHP version, app type, DB service, DB name, search engine
# → Creates source dir, offers to create vhost + database

# 2. Clone code
git clone git@github.com:you/your-magento.git sources/mysite.com

# 3. Start (auto-detects needed services)
bin/mage up

# 4. Import database (auto-routes to the right DB service)
bin/mage db import mysite.com backup.sql

# 5. Update env.php — DB host is the service name (mysql / mysql80 / mariadb)

# 6. Build
bin/mage composer mysite.com install
bin/mage magento mysite.com setup:upgrade
```

### Installing Fresh Magento

```bash
bin/mage install 2.4.7 community shop.test php83
# → Downloads, creates DB, configures vhost, runs setup:install
```

### Enabling SSL

```bash
# Requires mkcert (https://github.com/FiloSottile/mkcert)
mkcert -install   # one-time: install local CA

bin/mage ssl shop.test php83
# → generates cert, creates HTTPS vhost, reloads nginx
# → site available at https://shop.test
```

### Using Xdebug with PHPStorm

```bash
# Enable Xdebug for your PHP version
bin/mage xdebug on php83

# In PHPStorm:
# 1. Settings → PHP → Servers → add "shop.test" mapping /home/public_html → sources/
# 2. Start listening for connections (phone icon)
# 3. Set a breakpoint and reload the page
```

See `images/xdebug-phpstorm-01.png` for PHPStorm configuration screenshot.

### Using Varnish

```bash
# Start Varnish
docker compose --profile varnish up -d varnish

# Enable for a domain
bin/mage varnish enable shop.test

# Disable
bin/mage varnish disable shop.test
```

---

## Docker Compose Profiles

`bin/mage up` auto-detects needed profiles from your projects. Manual override:

```bash
bin/mage up                                    # Smart: from projects.json
docker compose --profile legacy up -d           # Legacy PHP 7.x
docker compose --profile mysql80 up -d mysql80  # MySQL 8.0 on port 3307
docker compose --profile mariadb up -d mariadb  # MariaDB on port 3308
docker compose --profile debug up -d            # phpMyAdmin + Redis Commander
```

| Profile | Services | Use Case |
|---|---|---|
| *(default)* | nginx, php81–84, mysql, opensearch, redis, rabbitmq, mailpit | Magento 2.4.4+ |
| `legacy` | php70–74 | Magento 2.1–2.4.3 |
| `mysql80` | MySQL 8.0 (port 3307) | Legacy projects |
| `mariadb` | MariaDB 11.4 (port 3308) | Magento 2.4.8 |
| `elasticsearch` | ES 8.17 | Alternative to OpenSearch |
| `varnish` | Varnish 7.6 | Full-page cache |
| `debug` | phpMyAdmin, Redis Commander | DB/cache inspection |

---

## Environment Variables

Copy `env-example` to `.env` and adjust as needed:

### Database

| Variable | Default | Description |
|---|---|---|
| `MYSQL_VERSION` | `8.4` | MySQL image tag |
| `MYSQL_DATABASE` | `magento` | Default database name |
| `MYSQL_USER` | `magento` | Database user |
| `MYSQL_PASSWORD` | `magento` | Database password |
| `MYSQL_ROOT_PASSWORD` | `root` | Root password |
| `MARIADB_VERSION` | `11.4` | MariaDB (profile: mariadb) |
| `MYSQL80_VERSION` | `8.0` | MySQL 8.0 (profile: mysql80) |

### Search

| Variable | Default | Description |
|---|---|---|
| `OPENSEARCH_VERSION` | `2.19.1` | OpenSearch image tag |
| `ELASTICSEARCH_VERSION` | `8.17.0` | Elasticsearch image tag (when using elasticsearch profile) |

### Services

| Variable | Default | Description |
|---|---|---|
| `REDIS_VERSION` | `7.4` | Redis image tag |
| `RABBITMQ_VERSION` | `4.1` | RabbitMQ image tag |
| `RABBITMQ_DEFAULT_USER` | `admin` | RabbitMQ admin user |
| `RABBITMQ_DEFAULT_PASS` | `admin` | RabbitMQ admin password |
| `VARNISH_VERSION` | `7.6` | Varnish image tag |

### Ports

| Variable | Default | Description |
|---|---|---|
| `HTTP_PORT` | `80` | Nginx HTTP |
| `HTTPS_PORT` | `443` | Nginx HTTPS |
| `MYSQL_PORT` | `3306` | MySQL 8.4 |
| `MYSQL80_PORT` | `3307` | MySQL 8.0 |
| `PHPMYADMIN_PORT` | `8080` | phpMyAdmin |
| `MAILPIT_PORT` | `8025` | Mailpit web UI |
| `RABBITMQ_MGMT_PORT` | `15672` | RabbitMQ management UI |
| `OPENSEARCH_PORT` | `9200` | OpenSearch API |
| `REDIS_PORT` | `6379` | Redis |
| `REDISCMD_PORT` | `8081` | Redis Commander |

---

## Prerequisites

- **Docker Engine 20.10+** with Compose plugin v2 (`docker compose`)
- **Linux** (Ubuntu 22.04+ recommended) — works on macOS/WSL2 with caveats
- **mkcert** — optional, for local SSL certificates

```bash
# Verify Docker
docker compose version   # needs v2.x

# Install mkcert (Ubuntu)
sudo apt install mkcert
mkcert -install
```

---

## Migrating from v1

If you used the previous version of this stack:

| Old | New |
|---|---|
| `php74-c2` | `php74` |
| `php81-c2` | `php81` |
| `mailhog` | `mailpit` (port 8025) |
| `elasticsearch` | `opensearch` (default; ES available via profile) |
| `docker-compose up` | `docker compose up` (Compose v2 plugin) |
| Per-version Dockerfiles | Single unified `build/php/Dockerfile` |

**Steps:**
1. Back up your `.env` and nginx vhost configs
2. Pull the latest changes
3. Copy `env-example` to `.env` and merge your settings
4. Update nginx vhost configs: change PHP upstream names (remove `-c2` suffix)
5. Rebuild: `docker compose build --no-cache`
6. Start: `docker compose up -d`

---

## Vietnamese Documentation

Tài liệu tiếng Việt: [README.vi.md](README.vi.md)

---

## License

MIT
