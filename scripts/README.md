# Docker Magento Multiple PHP - Scripts

Management scripts for Docker-based Magento development environment with multiple PHP versions.

## Directory Structure

```
scripts/
├── lib/
│   └── common.sh              # Shared utility library
├── templates/
│   └── nginx/                 # Nginx configuration templates
│       ├── magento1.conf.template
│       ├── magento2.conf.template
│       ├── wordpress.conf.template
│       ├── laravel.conf.template
│       └── default.conf.template
├── create-vhost               # Create virtual hosts
├── database                   # Database management
├── fixowner                   # Fix file ownership
├── init-magento              # Automated Magento installation
├── list-services             # List running services
├── mysql                     # MySQL CLI access
├── setup-composer            # Configure Composer auth
├── shell                     # Access PHP containers
├── ssl                       # SSL certificate management
├── varnish                   # Varnish cache management
└── xdebug                    # Xdebug configuration
```

## Common Library

The `lib/common.sh` library provides shared functionality:

### Output Functions
- `_success()`, `_error()`, `_warning()`, `_arrow()` - Colored output
- `_header()`, `_note()`, `_bold()`, `_underline()` - Formatting

### Validation Functions
- `_typeExists()` - Check if command exists
- `_isOs*()` - OS detection helpers
- `_checkRootUser()` - Verify root privileges

### Docker Functions
- `getRunningServices()` - List running containers
- `isServiceRunning()` - Check specific service
- `getMysqlInformation()` - Get MySQL credentials
- `reloadNginx()` - Reload Nginx config

### User Interaction
- `askYesOrNo()` - Interactive yes/no prompts
- `_seekConfirmation()` - Confirmation dialogs
- `initYesNoPrompt()` - Initialize prompt variables

## Template System

Nginx configurations are stored as templates in `templates/nginx/`. Variables are replaced at runtime:
- `__DOMAIN__` - Domain name
- `__ROOT_DIR__` - Document root directory
- `__PHP_VERSION__` - PHP container version

## Usage Examples

### Create Virtual Host
```bash
./scripts/create-vhost --domain=test.local --app=magento2 --root-dir=mysite --php-version=php74
```

### Database Operations
```bash
./scripts/database list
./scripts/database create --database-name=mydb
./scripts/database export --database-name=mydb
./scripts/database import --source=backup.sql --target=mydb
```

### Install Magento
```bash
./scripts/init-magento \
  --domain=magento.local \
  --magento-version=2.4.6 \
  --magento-edition=community \
  --php-version=php81-c2
```

### SSL Certificates
```bash
./scripts/ssl --domain=test.local
```

### Xdebug Management
```bash
./scripts/xdebug enable --php-version=php74
./scripts/xdebug disable --php-version=php74
./scripts/xdebug status --php-version=php74
```

### Varnish Cache
```bash
./scripts/varnish enable --domain=test.local
./scripts/varnish disable --domain=test.local
./scripts/varnish status --domain=test.local
```

## Supported PHP Versions

- php70, php71, php72, php73
- php74, php74-c2 (Composer 2)
- php81-c2, php82

## Supported Applications

- **magento1** - Magento 1.x
- **magento2** - Magento 2.x
- **wordpress** - WordPress
- **laravel** - Laravel Framework
- **default** - Generic PHP application

## Error Handling

All scripts use:
- `set -e` for automatic error exit
- Consistent error messages via `_error()`
- Color-coded output for visibility
- Detailed validation before operations

## Development

### Adding New Scripts

1. Create script in `scripts/` directory
2. Source common library: `source "${SCRIPT_DIR}/lib/common.sh"`
3. Follow existing structure:
   ```bash
   #!/usr/bin/env bash
   set -e

   SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
   source "${SCRIPT_DIR}/lib/common.sh"

   # Your code here
   ```

### Adding New Templates

1. Create template in `scripts/templates/nginx/`
2. Use placeholders: `__DOMAIN__`, `__ROOT_DIR__`, `__PHP_VERSION__`
3. Update `VALID_APP_TYPES` in `create-vhost`

## Notes

- All scripts include `--help` flag for usage information
- Scripts use colored output for better readability
- Common operations are centralized in `lib/common.sh`
- Templates allow easy customization of configurations
