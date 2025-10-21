# Docker Compose Modular Configuration

This directory contains modular Docker Compose files that can be combined to create custom service configurations.

## Overview

Instead of a single monolithic `docker-compose.yml`, services are organized into logical groups that can be mixed and matched based on your needs.

**Benefits**:
- Start only the services you need
- Reduce resource usage
- Faster startup times
- Easy to customize for different projects
- Clear separation of concerns

## Directory Structure

```
compose/
├── core/                   # Essential services
│   ├── web.yml            # Nginx web server
│   ├── database.yml       # MySQL database
│   └── mail.yml           # Mailhog mail catcher
├── php/                    # PHP-FPM versions
│   ├── modern.yml         # PHP 8.2, 8.1-c2, 7.4-c2 (Magento 2.4+)
│   ├── standard.yml       # PHP 7.4, 7.3, 7.2 (Magento 2.3)
│   └── php82.yml          # PHP 8.2 only (minimal)
├── optional/               # Optional services
│   ├── elasticsearch.yml  # Elasticsearch + Kibana
│   ├── cache.yml          # Redis + Varnish
│   ├── queue.yml          # RabbitMQ
│   └── admin.yml          # phpMyAdmin + phpRedmin
├── presets/                # Pre-configured combinations
│   ├── magento-2.3.yml    # Magento 2.3.x stack
│   ├── magento-2.4.yml    # Magento 2.4.x stack
│   └── minimal.yml        # Minimal setup
└── volumes.yml             # Shared volume definitions
```

## Service Groups

### Core Services (Always Required)

**web.yml** - Nginx Web Server
- Service: `nginx`
- Ports: 80, 443
- Configuration: `/etc/nginx/sites-available/`, `/etc/nginx/sites-enabled/`

**database.yml** - MySQL Database
- Service: `mysql`
- Ports: 3306
- Version: Configurable via `MYSQL_VERSION` env var
- Authentication: `mysql_native_password`

**mail.yml** - Mailhog Mail Catcher
- Service: `mailhog`
- Ports: 1025 (SMTP), 8025 (Web UI)
- Use for testing email in development

### PHP Versions

**modern.yml** - Modern PHP Stack (Magento 2.4+)
- Services: `php82`, `php81-c2`, `php74-c2`
- Features: Composer 2, Xdebug 3.x
- Best for: Magento 2.4.0+, modern applications

**standard.yml** - Standard PHP Stack (Magento 2.3)
- Services: `php74`, `php73`, `php72`
- Features: Composer 1.x/2.x hybrid
- Best for: Magento 2.3.x, legacy applications

**php82.yml** - Single PHP 8.2
- Service: `php82`
- Use for: Minimal setups, single-version projects

### Optional Services

**elasticsearch.yml** - Search Engine
- Services: `elasticsearch`, `kibana`
- Ports: 9200 (ES), 5601 (Kibana)
- Required for: Magento 2.4+ catalog search
- Version: Configurable via `ELASTICSEARCH_VERSION`

**cache.yml** - Performance Cache
- Services: `redis`, `varnish`
- Ports: 6379 (Redis), 6081 (Varnish), 6082 (Varnish Admin)
- Use for: Production-like performance testing

**queue.yml** - Message Queue
- Service: `rabbitmq`
- Ports: 5672 (AMQP), 15672 (Management UI)
- Required for: Magento Commerce, async processing

**admin.yml** - Admin Tools
- Services: `phpmyadmin`, `phpredmin`
- Ports: 8080 (phpMyAdmin), 8081 (phpRedmin)
- Use for: Database and cache inspection

## Using the Service Manager

The easiest way to manage services is using the `scripts/services` script:

### Quick Start with Presets

```bash
# Magento 2.4 stack (PHP 8.2/8.1/7.4 + Elasticsearch + Cache)
./scripts/services preset magento-2.4

# Magento 2.3 stack (PHP 7.4/7.3/7.2)
./scripts/services preset magento-2.3

# Minimal setup (PHP 8.2 only)
./scripts/services preset minimal
```

### Custom Service Selection

```bash
# Core + modern PHP + Elasticsearch
./scripts/services start core php-modern elasticsearch

# Core + PHP 8.2 + cache + admin tools
./scripts/services start core php82 cache admin

# Full stack
./scripts/services start core php-modern elasticsearch cache queue admin
```

### Interactive Mode

```bash
./scripts/services interactive
```

This will guide you through selecting:
- PHP versions (modern/standard/minimal)
- Elasticsearch (yes/no)
- Cache layer (yes/no)
- Message queue (yes/no)
- Admin tools (yes/no)

### Other Commands

```bash
# List available services and presets
./scripts/services list

# Show current configuration
./scripts/services status

# Generate docker-compose.yml without starting
./scripts/services generate core php82

# Stop all services
./scripts/services stop

# Reset to original docker-compose.yml
./scripts/services reset
```

## Manual Usage (Advanced)

If you prefer to manually create your `docker-compose.yml`:

```yaml
# docker-compose.yml
include:
  - compose/core/web.yml
  - compose/core/database.yml
  - compose/core/mail.yml
  - compose/php/modern.yml
  - compose/optional/elasticsearch.yml
  - compose/volumes.yml
```

Then start services:

```bash
docker-compose up -d
```

## Preset Configurations

### magento-2.3.yml

**Included Services**:
- Core: Nginx, MySQL, Mailhog
- PHP: 7.4, 7.3, 7.2
- Total: 6 containers

**Use Cases**:
- Magento 2.3.x development
- Legacy PHP projects
- Multi-version PHP testing

**Resource Usage**: Medium

### magento-2.4.yml

**Included Services**:
- Core: Nginx, MySQL, Mailhog
- PHP: 8.2, 8.1-c2, 7.4-c2
- Optional: Elasticsearch, Kibana, Redis, Varnish
- Total: 10 containers

**Use Cases**:
- Magento 2.4.0+ development
- Full-featured development environment
- Performance testing with cache

**Resource Usage**: High

### minimal.yml

**Included Services**:
- Core: Nginx, MySQL, Mailhog
- PHP: 8.2
- Total: 4 containers

**Use Cases**:
- Single-version projects
- Resource-constrained environments
- Quick testing

**Resource Usage**: Low

## Volume Management

All persistent data is stored in Docker volumes defined in `volumes.yml`:

- `composer_cache` - Composer package cache (shared across PHP containers)
- `elasticsearch-data` - Elasticsearch indices
- `redis_data` - Redis persistence
- `rabbitmq-data` - RabbitMQ messages and config

Volumes persist even when containers are stopped.

## Environment Variables

Configure services via `.env` file:

```bash
# PHP versions
PHP_VERSION_82=8.2-fpm-bullseye
PHP_VERSION_81=8.1-fpm-bullseye
PHP_VERSION_74=7.4-fpm-buster

# MySQL
MYSQL_VERSION=8.0

# Elasticsearch
ELASTICSEARCH_VERSION=7.17.1

# Redis
REDIS_VERSION=6.0-alpine

# Varnish
VARNISH_VERSION=6.0
```

## Networking

All services communicate via the `magento-network` bridge network.

**Service Names** (use these for connections):
- Database: `mysql`
- Web server: `nginx`
- PHP containers: `php82`, `php81-c2`, `php74`, etc.
- Elasticsearch: `elasticsearch`
- Redis: `redis`
- Varnish: `varnish`
- RabbitMQ: `rabbitmq`

**Example Magento env.php**:

```php
'db' => [
    'connection' => [
        'default' => [
            'host' => 'mysql',
            'dbname' => 'magento',
            'username' => 'magento',
            'password' => 'magento',
        ],
    ],
],
'cache' => [
    'frontend' => [
        'default' => [
            'backend' => 'Cm_Cache_Backend_Redis',
            'backend_options' => [
                'server' => 'redis',
                'port' => '6379',
            ],
        ],
    ],
],
```

## Troubleshooting

### Service won't start

```bash
# Check service status
./scripts/services status

# View logs
docker-compose logs -f <service-name>

# Recreate containers
docker-compose down
./scripts/services preset magento-2.4
```

### Port conflicts

If you get "port already in use" errors:

1. Check what's using the port: `lsof -i :80`
2. Stop the conflicting service
3. Or modify ports in compose files

### Volume permission issues

```bash
# Fix ownership of project files
./scripts/fixowner
```

### Reset everything

```bash
# Stop all services
./scripts/services stop

# Remove volumes (WARNING: deletes all data)
docker-compose down -v

# Start fresh
./scripts/services preset magento-2.4
```

## Best Practices

1. **Start with presets** - Use `magento-2.4` or `magento-2.3` presets as starting points
2. **Use minimal for testing** - Quick service validation with minimal resources
3. **Add services incrementally** - Start with core, add optional services as needed
4. **Check status regularly** - Use `./scripts/services status` to verify configuration
5. **Backup before changes** - Original config saved as `docker-compose.yml.backup`

## Migration from Monolithic Config

If you have an existing `docker-compose.yml`:

```bash
# Backup is created automatically
./scripts/services preset magento-2.4

# To restore original
./scripts/services reset
```

Your original configuration is preserved as `docker-compose.yml.backup`.

## Performance Tips

**Minimize running containers**:
- Use `php82.yml` instead of `modern.yml` if you don't need multiple PHP versions
- Skip `admin.yml` unless actively debugging
- Disable `queue.yml` for Community Edition

**Resource allocation**:
- Elasticsearch needs 2GB+ RAM
- Assign Docker at least 4GB for Magento 2.4
- Assign Docker at least 2GB for minimal setup

## Contributing

When adding new services:

1. Create a new `.yml` file in appropriate directory (`core/`, `optional/`, `php/`)
2. Define service with clear naming
3. Add to `scripts/services` service groups
4. Document in this README
5. Create or update relevant preset

## See Also

- [Scripts Documentation](../scripts/README.md)
- [Main README](../README.md)
- [Docker Compose Include Docs](https://docs.docker.com/compose/compose-file/14-include/)
