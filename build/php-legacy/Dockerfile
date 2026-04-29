# ==============================================================================
# Legacy PHP Dockerfile — Builds PHP 7.0–7.3 from source
# For versions no longer available in ondrej/php PPA
# ==============================================================================

ARG BASE_IMAGE=ubuntu:focal
FROM ${BASE_IMAGE}

ARG PHP_VERSION=7.2
ARG PHP_PATCH_VERSION=7.2.34
ARG COMPOSER_VERSION=1.10.19
ARG NODE_MAJOR=18
ARG XDEBUG_VERSION=2.9.8

ENV PHP_VERSION=${PHP_VERSION}
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=UTC

# PHP install paths (match ondrej/php layout so our configs/scripts work unchanged)
ENV PHP_PREFIX=/usr
ENV PHP_INI_DIR=/etc/php/${PHP_VERSION}
ENV PHP_FPM_CONF_DIR=/etc/php/${PHP_VERSION}/fpm
ENV PHP_CLI_CONF_DIR=/etc/php/${PHP_VERSION}/cli

# ── User & directory setup ────────────────────────────────────────────────────
RUN if getent passwd 1000 >/dev/null 2>&1; then userdel -r $(getent passwd 1000 | cut -d: -f1) 2>/dev/null || true; fi && \
    if getent group 1000 >/dev/null 2>&1; then groupdel $(getent group 1000 | cut -d: -f1) 2>/dev/null || true; fi && \
    mkdir -p /home/public_html /run/php /var/lib/php/session /var/log/php-fpm && \
    groupadd -g 1000 nginx && \
    useradd -u 1000 -g nginx -m -d /home/nginx -s /bin/bash nginx && \
    chown -R nginx:nginx /home/public_html/ && \
    chown -R nginx:nginx /var/lib/php/

WORKDIR /home/public_html

# ── Base packages ─────────────────────────────────────────────────────────────
RUN apt-get update -y && \
    apt-get install -y \
        curl git sudo apt-utils vim zip unzip tzdata gnupg ca-certificates \
        software-properties-common \
        # Build tools for compiling PHP
        autoconf dpkg-dev file g++ gcc libc-dev make pkg-config re2c bison \
        # PHP dependency libraries
        libxml2-dev libcurl4-openssl-dev libssl-dev libreadline-dev \
        libsqlite3-dev zlib1g-dev libpng-dev libjpeg-dev libfreetype6-dev \
        libicu-dev libxslt1-dev libzip-dev libonig-dev libmcrypt-dev \
        libsodium-dev libbz2-dev libgmp-dev libc-client-dev \
        libkrb5-dev libwebp-dev xz-utils \
    && \
    # tini (may not be in older repos)
    # These packages may not exist on all distros — install if available
    apt-get install -y libargon2-dev libfreetype-dev 2>/dev/null || true && \
    (apt-get install -y tini 2>/dev/null || \
        (curl -fsSL https://github.com/krallin/tini/releases/download/v0.19.0/tini -o /usr/bin/tini && chmod +x /usr/bin/tini)) && \
    dpkg-reconfigure -f noninteractive tzdata && \
    # PHP 7.x configure needs legacy *-config scripts removed in newer distros.
    # Create shims that use pkg-config instead.
    if ! command -v freetype-config >/dev/null 2>&1; then \
        printf '#!/bin/sh\ncase "$1" in\n  --cflags) pkg-config --cflags freetype2;;\n  --libs) pkg-config --libs freetype2;;\n  --prefix) pkg-config --variable=prefix freetype2;;\n  --version) pkg-config --modversion freetype2;;\n  *) pkg-config "$@" freetype2;;\nesac\n' > /usr/local/bin/freetype-config && \
        chmod +x /usr/local/bin/freetype-config; \
    fi && \
    if ! command -v icu-config >/dev/null 2>&1; then \
        printf '#!/bin/sh\ncase "$1" in\n  --cflags|--cppflags) pkg-config --cflags icu-uc icu-io;;\n  --ldflags|--ldflags-searchpath) pkg-config --libs-only-L icu-uc;;\n  --libs) echo "-licui18n -licuuc -licudata";;\n  --prefix) pkg-config --variable=prefix icu-uc;;\n  --version) pkg-config --modversion icu-uc;;\n  --detect-prefix) true;;\n  *) echo "";;\nesac\n' > /usr/local/bin/icu-config && \
        chmod +x /usr/local/bin/icu-config; \
    fi && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# ── Download & compile PHP from source ────────────────────────────────────────
RUN mkdir -p /usr/src && cd /usr/src && \
    curl -fsSL "https://www.php.net/distributions/php-${PHP_PATCH_VERSION}.tar.xz" -o php.tar.xz && \
    tar xf php.tar.xz && \
    cd "php-${PHP_PATCH_VERSION}" && \
    \
    gnuArch="$(dpkg-architecture --query DEB_BUILD_GNU_TYPE)" && \
    debMultiarch="$(dpkg-architecture --query DEB_BUILD_MULTIARCH)" && \
    \
    # Build configure flags dynamically based on PHP version
    PHP_MAJOR=$(echo ${PHP_VERSION} | cut -d. -f1) && \
    PHP_MINOR=$(echo ${PHP_VERSION} | cut -d. -f2) && \
    \
    CONFIGURE_FLAGS="\
        --build=$gnuArch \
        --prefix=${PHP_PREFIX} \
        --sbindir=${PHP_PREFIX}/sbin \
        --sysconfdir=${PHP_INI_DIR} \
        --with-config-file-path=${PHP_INI_DIR}/cli \
        --with-config-file-scan-dir=${PHP_INI_DIR}/cli/conf.d \
        --with-libdir=lib/$debMultiarch \
        --enable-fpm --with-fpm-user=nginx --with-fpm-group=nginx \
        --enable-mbstring --enable-mysqlnd --enable-opcache \
        --enable-intl --enable-bcmath --enable-soap \
        --enable-zip --enable-exif --enable-sockets \
        --with-curl --with-gd --with-jpeg-dir=/usr --with-png-dir=/usr \
        --with-freetype-dir=/usr --with-iconv --with-openssl \
        --with-readline --with-zlib --with-pdo-mysql=mysqlnd \
        --with-mysqli=mysqlnd --with-mysql-sock --with-xsl \
        --with-mhash --with-mcrypt --with-bz2 --with-gmp \
        --with-pic --with-pear --disable-cgi --disable-phpdbg" && \
    \
    # --with-webp-dir: PHP 7.0 doesn't have it
    if [ "$PHP_MINOR" -ge 1 ]; then \
        CONFIGURE_FLAGS="$CONFIGURE_FLAGS --with-webp-dir=/usr"; \
    fi && \
    \
    # --with-xmlrpc: may fail on some setups, add if available
    CONFIGURE_FLAGS="$CONFIGURE_FLAGS --with-xmlrpc" && \
    \
    ./configure $CONFIGURE_FLAGS \
    && \
    make -j "$(nproc)" && \
    make install && \
    make clean && \
    \
    # Clean up source
    cd / && rm -rf /usr/src/php* && \
    \
    # Smoke test
    php -v

# ── Set up PHP directory structure (match ondrej/php layout) ──────────────────
RUN mkdir -p \
        "${PHP_FPM_CONF_DIR}/pool.d" \
        "${PHP_FPM_CONF_DIR}/conf.d" \
        "${PHP_CLI_CONF_DIR}/conf.d" \
        "${PHP_INI_DIR}/mods-available" && \
    \
    # Create FPM config if not exists
    if [ ! -f "${PHP_FPM_CONF_DIR}/php-fpm.conf" ]; then \
        cp "${PHP_INI_DIR}/php-fpm.conf.default" "${PHP_FPM_CONF_DIR}/php-fpm.conf" 2>/dev/null || true; \
        # Point FPM to our pool directory
        echo "include=${PHP_FPM_CONF_DIR}/pool.d/*.conf" >> "${PHP_FPM_CONF_DIR}/php-fpm.conf"; \
    fi && \
    \
    # Create symlinks so /usr/sbin/php-fpmX.Y works (matching ondrej convention)
    ln -sf /usr/sbin/php-fpm "/usr/sbin/php-fpm${PHP_VERSION}" && \
    \
    # Create symlinks for php binary versioned names
    ln -sf /usr/bin/php "/usr/bin/php${PHP_VERSION}" && \
    ln -sf /usr/bin/phpize "/usr/bin/phpize${PHP_VERSION}" && \
    ln -sf /usr/bin/php-config "/usr/bin/php-config${PHP_VERSION}" && \
    ln -sf /usr/bin/phar.phar "/usr/bin/phar${PHP_VERSION}" && \
    ln -sf /usr/bin/phar.phar "/usr/bin/phar.phar${PHP_VERSION}"

# ── Composer ──────────────────────────────────────────────────────────────────
RUN curl -fsSL "https://getcomposer.org/download/${COMPOSER_VERSION}/composer.phar" \
        --output /usr/bin/composer && \
    chmod +x /usr/bin/composer

# ── n98-magerun2 ──────────────────────────────────────────────────────────────
RUN curl -fsSL https://files.magerun.net/n98-magerun2.phar \
        --output /usr/bin/n98-magerun2.phar && \
    chmod +x /usr/bin/n98-magerun2.phar

# ── Node.js + Grunt (best-effort for legacy distros) ──────────────────────────
# Nodesource may not support old distros. Fall back to distro nodejs.
RUN ( \
        curl -fsSL "https://deb.nodesource.com/setup_${NODE_MAJOR}.x" -o /tmp/nodesource_setup.sh && \
        bash /tmp/nodesource_setup.sh && \
        apt-get install -y nodejs && \
        rm -f /tmp/nodesource_setup.sh \
    ) || ( \
        echo "Nodesource ${NODE_MAJOR}.x not available, installing distro nodejs..." && \
        apt-get update -y && \
        apt-get install -y nodejs npm 2>/dev/null || true \
    ) && \
    (npm install -g grunt-cli 2>/dev/null || true) && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# ── NVM for nginx user ───────────────────────────────────────────────────────
RUN su - nginx -c "touch /home/nginx/.bash_profile" && \
    su - nginx -c "curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/master/install.sh | bash" && \
    chmod +x /home/nginx/.nvm/nvm.sh && \
    su - nginx -c "echo '[ -s /home/nginx/.nvm/nvm.sh ] && . /home/nginx/.nvm/nvm.sh' > .profile"

# ── Xdebug ────────────────────────────────────────────────────────────────────
RUN pecl channel-update pecl.php.net && \
    pecl install -f "xdebug-${XDEBUG_VERSION}" && \
    # Don't enable by default
    rm -f "${PHP_CLI_CONF_DIR}/conf.d/"*xdebug* \
          "${PHP_FPM_CONF_DIR}/conf.d/"*xdebug* 2>/dev/null || true

# ── APCu ──────────────────────────────────────────────────────────────────────
RUN pecl install -f apcu && \
    echo "extension=apcu.so" > "${PHP_INI_DIR}/mods-available/apcu.ini" && \
    ln -sf "${PHP_INI_DIR}/mods-available/apcu.ini" "${PHP_CLI_CONF_DIR}/conf.d/20-apcu.ini" && \
    ln -sf "${PHP_INI_DIR}/mods-available/apcu.ini" "${PHP_FPM_CONF_DIR}/conf.d/20-apcu.ini"

# ── mhsendmail (Mailpit-compatible) ──────────────────────────────────────────
RUN ARCH=$(dpkg --print-architecture) && \
    if [ "${ARCH}" = "amd64" ]; then \
        curl -fsSL "https://github.com/mailhog/mhsendmail/releases/download/v0.2.0/mhsendmail_linux_amd64" \
            --output /usr/local/bin/mhsendmail; \
    else \
        printf '#!/bin/sh\nexec /usr/sbin/sendmail "$@"\n' > /usr/local/bin/mhsendmail; \
    fi && \
    chmod +x /usr/local/bin/mhsendmail

# ── nginx user setup ─────────────────────────────────────────────────────────
RUN mkdir -p /home/nginx/.composer && \
    chown -R nginx:nginx /home/nginx/.composer && \
    echo "nginx ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/nginx && \
    chmod 0440 /etc/sudoers.d/nginx

# ── Entrypoint ────────────────────────────────────────────────────────────────
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/usr/local/bin/docker-entrypoint.sh"]
