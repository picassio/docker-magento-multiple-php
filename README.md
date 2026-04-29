# 🐘 Docker Magento Multi-PHP

> Multi-PHP Docker development stack for Magento (and WordPress, Laravel)

**PHP 7.0 – 8.4** · **Magento 2.1 – 2.4.8** · **Linux-focused** · **Docker Compose v2**

---

## Quick Start

```bash
git clone https://github.com/picassio/docker-magento-multiple-php.git ~/docker-magento
cd ~/docker-magento
cp env-example .env
docker compose up -d nginx php83 mysql opensearch redis mailpit
# Create a vhost:
bin/mage vhost mysite.com magento2 php83
```

Your site is now available at `http://mysite.com` (add it to `/etc/hosts` → `127.0.0.1`).

---

## Features

- **Multi-PHP** — PHP 7.0 through 8.4, all from a single unified Dockerfile
- **Auto virtual hosts** — create Nginx vhosts for Magento 1/2, WordPress, Laravel, or generic PHP
- **SSL certificates** — one-command local HTTPS via `mkcert`
- **Xdebug toggle** — enable/disable per PHP version without rebuilding
- **Database management** — create, drop, import, export MySQL/MariaDB databases
- **Email catching** — Mailpit catches all outgoing mail (web UI on port 8025)
- **Search engines** — OpenSearch 2.x (default) or Elasticsearch 8.x
- **Redis** — session/cache backend with optional Redis Commander UI
- **RabbitMQ** — message queue with management UI
- **Varnish** — full-page cache reverse proxy (opt-in)
- **Composer auth** — one-time setup for `repo.magento.com`
- **Fresh install** — download and install Magento from scratch

---

## Architecture

```
.
├── bin/mage              # CLI wrapper (routes to scripts/)
├── build/php/            # Unified Dockerfile (all PHP versions)
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
| **mariadb** | `mariadb:11.4` | — | `mariadb` | 3306 |
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

| Command | Description | Example |
|---|---|---|
| `vhost <domain> <type> <php>` | Create Nginx virtual host | `bin/mage vhost shop.test magento2 php83` |
| `ssl <domain> <php>` | Generate SSL cert & HTTPS vhost | `bin/mage ssl shop.test php83` |
| `db create <name>` | Create a MySQL database | `bin/mage db create mydb` |
| `db drop <name>` | Drop a MySQL database | `bin/mage db drop mydb` |
| `db import <name> <file>` | Import SQL dump into database | `bin/mage db import mydb dump.sql.gz` |
| `db export <name>` | Export database to `databases/export/` | `bin/mage db export mydb` |
| `db list` | List all databases | `bin/mage db list` |
| `xdebug on <php>` | Enable Xdebug | `bin/mage xdebug on php83` |
| `xdebug off <php>` | Disable Xdebug | `bin/mage xdebug off php83` |
| `xdebug status <php>` | Check Xdebug status | `bin/mage xdebug status php83` |
| `shell <php>` | Open bash shell in PHP container | `bin/mage shell php83` |
| `mysql` | Open MySQL CLI as root | `bin/mage mysql` |
| `composer-auth` | Setup Magento Composer credentials | `bin/mage composer-auth` |
| `install <domain> <php> <ver>` | Install fresh Magento | `bin/mage install shop.test php83 2.4.7` |
| `fixowner` | Fix file ownership in `sources/` | `bin/mage fixowner` |
| `varnish enable <domain>` | Enable Varnish for a domain | `bin/mage varnish enable shop.test` |
| `varnish disable <domain>` | Disable Varnish for a domain | `bin/mage varnish disable shop.test` |
| `services` | List running services | `bin/mage services` |

---

## Common Workflows

### Setting Up an Existing Magento Project

```bash
# 1. Clone your project into sources/
git clone git@github.com:you/your-magento.git sources/yoursite

# 2. Start the stack with the PHP version you need
docker compose up -d nginx php83 mysql opensearch redis mailpit

# 3. Create the vhost
bin/mage vhost yoursite.test magento2 php83

# 4. Import the database
bin/mage db create yoursite
bin/mage db import yoursite databases/import/yoursite.sql.gz

# 5. Update app/etc/env.php with DB credentials (host: mysql, user: magento, pass: magento)

# 6. Add to /etc/hosts
echo "127.0.0.1 yoursite.test" | sudo tee -a /etc/hosts
```

### Installing Fresh Magento

```bash
# 1. Setup Composer auth (one-time)
bin/mage composer-auth

# 2. Install Magento 2.4.7 with PHP 8.3
bin/mage install shop.test php83 2.4.7

# 3. Access at http://shop.test
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

Services are organized into profiles to keep the default stack lean:

```bash
# Default stack (Nginx + PHP 8.1-8.4 + MySQL + OpenSearch + Redis + RabbitMQ + Mailpit)
docker compose up -d

# Add legacy PHP 7.x support
docker compose --profile legacy up -d

# Add debug tools (phpMyAdmin + Redis Commander)
docker compose --profile debug up -d

# Use MariaDB instead of MySQL
docker compose --profile mariadb up -d

# Enable Varnish cache
docker compose --profile varnish up -d

# Use Elasticsearch instead of OpenSearch
docker compose --profile elasticsearch up -d

# Everything at once
docker compose --profile legacy --profile debug --profile varnish up -d
```

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
| `MARIADB_VERSION` | `11.4` | MariaDB image tag (when using mariadb profile) |

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
| `MYSQL_PORT` | `3306` | MySQL |
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
