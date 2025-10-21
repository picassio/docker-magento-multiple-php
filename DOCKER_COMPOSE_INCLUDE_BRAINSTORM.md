# Docker Compose Include Strategy - Brainstorming

## Current State Analysis

**Problem**: The current `docker-compose.yml` has significant repetition:
- 8 PHP service definitions (php70-php82, php74-c2, php81-c2) - almost identical
- Mixed concerns: core services, optional services, monitoring tools
- 237 lines with ~70% repetition in PHP services
- Difficult to add/remove PHP versions
- Hard to enable/disable optional stacks

## Docker Compose Include Feature

Available in Docker Compose v2.20+ (Compose Specification):
```yaml
include:
  - path: ./compose-files/php-services.yml
  - path: ./compose-files/optional-services.yml
    env_file: .env.optional
```

## 🎯 Strategy 1: Service Grouping (Recommended)

### Structure
```
docker-compose.yml              # Main orchestration file
compose/
├── core/
│   ├── web.yml                # Nginx
│   ├── database.yml           # MySQL
│   └── mail.yml               # Mailhog
├── php/
│   ├── php70.yml              # Individual PHP versions
│   ├── php71.yml
│   ├── php72.yml
│   ├── php73.yml
│   ├── php74.yml
│   ├── php74-c2.yml
│   ├── php81-c2.yml
│   └── php82.yml
├── optional/
│   ├── elasticsearch-stack.yml  # Elasticsearch + Kibana
│   ├── cache.yml                # Redis + Varnish
│   ├── queue.yml                # RabbitMQ
│   └── admin-tools.yml          # phpMyAdmin, phpRedmin
└── shared/
    └── volumes.yml              # Shared volume definitions
```

### Main docker-compose.yml
```yaml
# Main orchestration file
include:
  # Core services (always included)
  - compose/core/web.yml
  - compose/core/database.yml
  - compose/core/mail.yml

  # PHP versions (include only what you need)
  - compose/php/php74.yml
  - compose/php/php81-c2.yml
  - compose/php/php82.yml

  # Optional services (comment out if not needed)
  - compose/optional/elasticsearch-stack.yml
  - compose/optional/cache.yml
  # - compose/optional/queue.yml          # Disabled by default
  # - compose/optional/admin-tools.yml    # Disabled by default

  # Shared resources
  - compose/shared/volumes.yml
```

### Benefits
✅ Easy to enable/disable entire stacks (comment one line)
✅ Clear separation of concerns
✅ Easy to understand what's running
✅ Can version control which services are enabled per environment

---

## 🎯 Strategy 2: Template-Based PHP Services

### Structure
```
docker-compose.yml
compose/
├── templates/
│   └── php-template.yml       # PHP service template with variables
└── php-services.yml            # Includes template multiple times
```

### PHP Template (php-template.yml)
```yaml
services:
  ${PHP_SERVICE_NAME}:
    build:
      context: ./build/${PHP_SERVICE_NAME}
    image: ${PHP_IMAGE_TAG}
    hostname: ${PHP_SERVICE_NAME}
    extra_hosts:
      - "host.docker.internal:host-gateway"
    volumes:
      - ./sources:/home/public_html
      - composer_cache:/home/nginx/.composer
      - ./conf/php/${PHP_SERVICE_NAME}/magento.conf:/etc/php/${PHP_VERSION}/fpm/pool.d/www.conf
      - ./conf/php/${PHP_SERVICE_NAME}/php.ini:/etc/php/${PHP_VERSION}/fpm/php.ini
      - ./conf/php/${PHP_SERVICE_NAME}/php.ini:/etc/php/${PHP_VERSION}/cli/php.ini
```

### Limitations
⚠️ Docker Compose doesn't support variable expansion in service names in include
❌ This approach won't work without custom tooling
💡 Better to use Strategy 1 or 3

---

## 🎯 Strategy 3: Profile-Based Activation

### Structure
```yaml
# docker-compose.yml
services:
  # Core services (no profiles)
  nginx:
    # ... nginx config

  mysql:
    # ... mysql config

  # PHP 7.0 - Legacy profile
  php70:
    profiles: ["php70", "legacy", "all"]
    # ... config

  # PHP 7.4 - Standard profile
  php74:
    profiles: ["php74", "standard", "all"]
    # ... config

  # PHP 8.1 - Modern profile
  php81-c2:
    profiles: ["php81", "modern", "all"]
    # ... config

  # PHP 8.2 - Latest profile
  php82:
    profiles: ["php82", "latest", "all"]
    # ... config

  # Optional services with profiles
  elasticsearch:
    profiles: ["magento24", "search", "all"]
    # ... config

  rabbitmq:
    profiles: ["queue", "enterprise", "all"]
    # ... config
```

### Usage
```bash
# Start only PHP 7.4 and PHP 8.1
docker-compose --profile php74 --profile php81 up -d

# Start all modern PHP versions
docker-compose --profile modern up -d

# Start everything
docker-compose --profile all up -d

# Start Magento 2.4 stack (with Elasticsearch)
docker-compose --profile magento24 --profile php81 up -d
```

### Benefits
✅ Single docker-compose.yml file
✅ Flexible service activation
✅ Can combine profiles
✅ No file management

### Drawbacks
⚠️ All definitions in one file (less modular)
⚠️ Profiles need to be specified each time

---

## 🎯 Strategy 4: Hybrid Approach (Best of All)

### Structure
```
docker-compose.yml              # Main file with core services
compose/
├── php/
│   ├── legacy.yml             # PHP 7.0, 7.1
│   ├── standard.yml           # PHP 7.2, 7.3, 7.4
│   ├── modern.yml             # PHP 8.1, 8.2
│   └── all.yml                # All PHP versions
├── stacks/
│   ├── magento-2.3.yml        # For Magento 2.3 (PHP 7.2-7.4, MySQL, Nginx)
│   ├── magento-2.4.yml        # For Magento 2.4 (PHP 7.4-8.2, ES, MySQL, Nginx)
│   └── full-stack.yml         # Everything including monitoring
└── environments/
    ├── development.yml        # Dev overrides (debug on, etc.)
    └── production.yml         # Prod overrides (optimizations)
```

### Usage Scenarios

**Scenario 1: Magento 2.4 Development**
```yaml
# docker-compose.yml
include:
  - compose/stacks/magento-2.4.yml
  - compose/environments/development.yml
```

**Scenario 2: Multi-version Testing**
```yaml
# docker-compose.yml
include:
  - compose/php/all.yml
  - compose/core/web.yml
  - compose/core/database.yml
```

**Scenario 3: Production-like Environment**
```yaml
# docker-compose.yml
include:
  - compose/stacks/magento-2.4.yml
  - compose/environments/production.yml
  - compose/optional/cache.yml
```

---

## 🎯 Strategy 5: Environment-Specific Includes

### Structure
```
docker-compose.yml                    # Symlink to environment-specific file
docker-compose.dev.yml               # Development configuration
docker-compose.staging.yml           # Staging configuration
docker-compose.test.yml              # Testing configuration

compose/
├── base/
│   └── services.yml                 # Base service definitions
└── overrides/
    ├── dev-overrides.yml
    ├── staging-overrides.yml
    └── test-overrides.yml
```

### Switching Environments
```bash
# Link to dev environment
ln -sf docker-compose.dev.yml docker-compose.yml

# Or specify explicitly
docker-compose -f docker-compose.dev.yml up -d
```

---

## 📊 Comparison Matrix

| Strategy | Modularity | Flexibility | Complexity | Maintenance | Best For |
|----------|------------|-------------|------------|-------------|----------|
| **1. Service Grouping** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | Production teams |
| **2. Templates** | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐ | Not recommended |
| **3. Profiles** | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | Single developers |
| **4. Hybrid** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | **RECOMMENDED** |
| **5. Environment-Specific** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | Multi-environment |

---

## 🏆 Recommended Implementation

### Phase 1: Immediate Wins (Service Grouping)

```
compose/
├── core.yml              # nginx, mysql, mailhog
├── php-legacy.yml        # php70, php71
├── php-standard.yml      # php72, php73, php74
├── php-modern.yml        # php74-c2, php81-c2, php82
├── elasticsearch.yml     # elasticsearch + kibana
├── cache.yml            # redis + varnish
├── queue.yml            # rabbitmq
├── admin.yml            # phpmyadmin, phpredmin
└── volumes.yml          # shared volumes
```

**Main docker-compose.yml**:
```yaml
include:
  - compose/core.yml
  - compose/php-modern.yml        # Most commonly used
  # - compose/php-standard.yml    # Uncomment if needed
  # - compose/php-legacy.yml      # Uncomment if needed
  - compose/elasticsearch.yml     # For Magento 2.4+
  - compose/cache.yml
  # - compose/queue.yml           # Uncomment for enterprise
  # - compose/admin.yml           # Uncomment for debugging
  - compose/volumes.yml
```

### Phase 2: Quick Start Presets

Create preset files for common scenarios:

```
presets/
├── magento-2.3.yml      # Include: core, php-standard
├── magento-2.4.yml      # Include: core, php-modern, elasticsearch
├── full-stack.yml       # Include: everything
└── minimal.yml          # Include: core, php82
```

Usage:
```bash
# Quick start with Magento 2.4
ln -sf presets/magento-2.4.yml docker-compose.yml
docker-compose up -d

# Or explicitly
docker-compose -f presets/magento-2.4.yml up -d
```

### Phase 3: Script Enhancement

Update management scripts to be aware of includes:

```bash
# scripts/init-magento
# Auto-detect required services and suggest includes

if [[ $MAGENTO_VERSION == 2.4* ]]; then
    echo "📋 Magento 2.4 detected"
    echo "✓ Required: core.yml, php-modern.yml, elasticsearch.yml"
    echo "Suggested: cache.yml for better performance"

    # Auto-generate compose file
    cat > docker-compose.yml <<EOF
include:
  - compose/core.yml
  - compose/php-modern.yml
  - compose/elasticsearch.yml
  - compose/cache.yml
  - compose/volumes.yml
EOF
fi
```

---

## 💡 Additional Ideas

### 1. Resource Profiles
```yaml
# compose/resources/development.yml
services:
  php82:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G

# compose/resources/production.yml
services:
  php82:
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 4G
```

### 2. Feature Toggles
```yaml
# compose/features/xdebug.yml
services:
  php82:
    environment:
      - XDEBUG_MODE=debug
      - XDEBUG_CONFIG=client_host=host.docker.internal
```

### 3. Version Matrix Testing
```yaml
# compose/testing/php-matrix.yml
# Include all PHP versions for CI/CD testing
include:
  - ../php/php74.yml
  - ../php/php81-c2.yml
  - ../php/php82.yml
```

### 4. Network Isolation
```yaml
# compose/networks/isolated.yml
# Separate networks for security testing
networks:
  frontend:
  backend:
  database:
```

---

## 🚀 Migration Path

### Step 1: Extract Core Services
1. Create `compose/core.yml` with nginx, mysql, mailhog
2. Test: `docker-compose -f compose/core.yml config`
3. Update main file to include it

### Step 2: Extract PHP Services
1. Create `compose/php-modern.yml` (most used)
2. Test independently
3. Update main file

### Step 3: Extract Optional Services
1. Move elasticsearch, rabbitmq, etc. to separate files
2. Make them opt-in via includes

### Step 4: Create Presets
1. Build preset files for common scenarios
2. Document usage
3. Update scripts to generate/recommend presets

### Step 5: Documentation
1. Update README with include strategy
2. Add examples for common use cases
3. Create troubleshooting guide

---

## ⚠️ Gotchas & Considerations

### 1. Merge Behavior
- Included files are merged, not overridden
- Later files take precedence for conflicts
- Use carefully with environment variables

### 2. Path Resolution
- Paths in included files are relative to that file
- `build: context: ../build/nginx` might need adjustment

### 3. Variable Scope
- `.env` file is read from main directory
- Can specify `env_file` per include

### 4. Dependency Management
- `depends_on` across includes works
- But harder to visualize dependencies

### 5. Volume Sharing
- Define volumes in one place (volumes.yml)
- Reference from multiple services

---

## 🎬 Example: Complete Modular Setup

### Directory Structure
```
.
├── docker-compose.yml
├── compose/
│   ├── core.yml
│   ├── php-modern.yml
│   ├── elasticsearch.yml
│   ├── cache.yml
│   └── volumes.yml
├── presets/
│   ├── magento-2.4.yml
│   └── full-stack.yml
└── scripts/
    └── init-magento  (updated to use includes)
```

### Workflow
```bash
# 1. Quick start with preset
docker-compose -f presets/magento-2.4.yml up -d

# 2. Custom configuration
# Edit docker-compose.yml to include desired services

# 3. Verify configuration
docker-compose config

# 4. Start services
docker-compose up -d

# 5. Check what's running
docker-compose ps
```

---

## 📝 Summary & Recommendation

**RECOMMENDED APPROACH**: **Strategy 4 (Hybrid)**

**Immediate Actions**:
1. ✅ Create `compose/` directory structure
2. ✅ Extract core services (nginx, mysql, mailhog)
3. ✅ Group PHP services by era (legacy/standard/modern)
4. ✅ Separate optional services (elasticsearch, cache, queue, admin)
5. ✅ Create preset files for common scenarios
6. ✅ Update scripts to work with includes
7. ✅ Document the new structure

**Benefits**:
- 🎯 Reduce 237 lines to ~150 with better organization
- 🎯 Easy to enable/disable entire stacks
- 🎯 Quick start with presets
- 🎯 Clear separation of concerns
- 🎯 Easier maintenance and updates
- 🎯 Better for CI/CD pipelines

**Next Steps**:
Should I implement this modular structure for you?
