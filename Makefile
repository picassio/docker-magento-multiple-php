# =============================================================================
# Docker Magento – Makefile convenience targets
# =============================================================================
# Usage:
#   make up                    # Start default stack
#   make up ARGS="php83 nginx" # Start specific services
#   make shell PHP=php83       # Open shell in PHP container
#   make logs ARGS=nginx       # Tail nginx logs
#   make help                  # Show all commands
# =============================================================================

.PHONY: up down restart stop status logs shell setup help

up:
	@bin/mage up $(ARGS)

down:
	@bin/mage down

restart:
	@bin/mage restart $(ARGS)

stop:
	@bin/mage stop

status:
	@bin/mage status

logs:
	@bin/mage logs $(ARGS)

shell:
	@bin/mage shell $(PHP)

setup:
	@bin/mage setup

help:
	@bin/mage help
