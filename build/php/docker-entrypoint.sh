#!/bin/sh
set -e

# If called with no arguments, start php-fpm with the correct version
if [ $# -eq 0 ]; then
    exec /usr/sbin/php-fpm${PHP_VERSION} -F -O \
        --fpm-config /etc/php/${PHP_VERSION}/fpm/php-fpm.conf
fi

# Otherwise, run whatever command was passed
exec "$@"
