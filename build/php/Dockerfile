# ==============================================================================
# Unified PHP Dockerfile — Builds PHP from source
# Supports PHP 8.1 – 8.4 via build arguments
# No external PPA dependency — works on any network (WSL2, VPN, air-gapped)
# ==============================================================================

# --- Build arguments ---------------------------------------------------------
ARG BASE_IMAGE=ubuntu:24.04
FROM ${BASE_IMAGE}

ARG PHP_VERSION=8.3
ARG PHP_PATCH_VERSION=""
ARG PHP_EXTENSIONS=""
ARG COMPOSER_VERSION=2.8.4
ARG NODE_MAJOR=20
ARG XDEBUG_VERSION=3.4.7

# Persist as env for runtime use
ENV PHP_VERSION=${PHP_VERSION}

# Prevent interactive prompts during build
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=UTC

# PHP install paths (matches ondrej/php layout so existing configs work)
ENV PHP_PREFIX=/usr
ENV PHP_INI_DIR=/etc/php/${PHP_VERSION}
ENV PHP_FPM_CONF_DIR=/etc/php/${PHP_VERSION}/fpm
ENV PHP_CLI_CONF_DIR=/etc/php/${PHP_VERSION}/cli

# --- System setup: user, directories, base packages -------------------------
RUN if getent passwd 1000 >/dev/null 2>&1; then userdel -r $(getent passwd 1000 | cut -d: -f1) 2>/dev/null || true; fi && \
    if getent group 1000 >/dev/null 2>&1; then groupdel $(getent group 1000 | cut -d: -f1) 2>/dev/null || true; fi && \
    mkdir -p /home/public_html /run/php /var/lib/php/session /var/log/php-fpm && \
    groupadd -g 1000 nginx && \
    useradd -u 1000 -g nginx -m -d /home/nginx -s /bin/bash nginx && \
    chown -R nginx:nginx /home/public_html/ && \
    chown -R nginx:nginx /var/lib/php/

WORKDIR /home/public_html

# --- Build dependencies ------------------------------------------------------
RUN apt-get update -y && \
    apt-get install -y \
        # Essentials
        curl git sudo apt-utils vim zip unzip tzdata gnupg ca-certificates \
        software-properties-common \
        # Build tools
        autoconf dpkg-dev file g++ gcc libc-dev make pkg-config re2c bison \
        # PHP dependency libraries
        libxml2-dev libcurl4-openssl-dev libssl-dev libreadline-dev \
        libsqlite3-dev zlib1g-dev libpng-dev libjpeg-dev libfreetype6-dev \
        libicu-dev libxslt1-dev libzip-dev libonig-dev libsodium-dev \
        libbz2-dev libgmp-dev libc-client-dev libkrb5-dev libwebp-dev \
        libargon2-dev libavif-dev xz-utils \
    && \
    (apt-get install -y tini 2>/dev/null || \
        (curl -fsSL https://github.com/krallin/tini/releases/download/v0.19.0/tini -o /usr/bin/tini && chmod +x /usr/bin/tini)) && \
    dpkg-reconfigure -f noninteractive tzdata && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# --- Resolve PHP patch version -----------------------------------------------
# If PHP_PATCH_VERSION is not set, resolve the latest patch for the minor version
RUN if [ -z "${PHP_PATCH_VERSION}" ]; then \
        RESOLVED=$(curl -fsSL "https://www.php.net/releases/index.php?json&version=${PHP_VERSION}" | \
            grep -o '"version":"[^"]*"' | head -1 | cut -d'"' -f4); \
        if [ -z "$RESOLVED" ]; then \
            echo "ERROR: Could not resolve patch version for PHP ${PHP_VERSION}" && exit 1; \
        fi; \
        echo "$RESOLVED" > /tmp/php_patch_version; \
    else \
        echo "${PHP_PATCH_VERSION}" > /tmp/php_patch_version; \
    fi && \
    echo "Building PHP $(cat /tmp/php_patch_version)"

# --- Download & compile PHP from source --------------------------------------
RUN PHP_FULL=$(cat /tmp/php_patch_version) && \
    mkdir -p /usr/src && cd /usr/src && \
    curl -fsSL "https://www.php.net/distributions/php-${PHP_FULL}.tar.xz" -o php.tar.xz && \
    tar xf php.tar.xz && \
    cd "php-${PHP_FULL}" && \
    \
    gnuArch="$(dpkg-architecture --query DEB_BUILD_GNU_TYPE)" && \
    debMultiarch="$(dpkg-architecture --query DEB_BUILD_MULTIARCH)" && \
    \
    ./configure \
        --build="$gnuArch" \
        --prefix=${PHP_PREFIX} \
        --sbindir=${PHP_PREFIX}/sbin \
        --sysconfdir=${PHP_INI_DIR} \
        --with-config-file-path=${PHP_INI_DIR}/cli \
        --with-config-file-scan-dir=${PHP_INI_DIR}/cli/conf.d \
        --with-libdir="lib/$debMultiarch" \
        \
        # FPM
        --enable-fpm --with-fpm-user=nginx --with-fpm-group=nginx \
        \
        # Core extensions (built-in)
        --enable-mbstring \
        --enable-mysqlnd \
        --enable-opcache \
        --enable-intl \
        --enable-bcmath \
        --enable-soap \
        --enable-exif \
        --enable-sockets \
        --enable-pcntl \
        --enable-ftp \
        --enable-gd --with-jpeg --with-freetype --with-webp --with-avif \
        \
        # Shared libraries
        --with-curl \
        --with-zip \
        --with-iconv \
        --with-openssl \
        --with-readline \
        --with-zlib \
        --with-pdo-mysql=mysqlnd \
        --with-mysqli=mysqlnd \
        --with-mysql-sock \
        --with-xsl \
        --with-mhash \
        --with-bz2 \
        --with-gmp \
        --with-sodium \
        --with-password-argon2 \
        --with-pear \
        --with-pic \
        \
        --disable-cgi \
        --disable-phpdbg \
    && \
    make -j "$(nproc)" && \
    make install && \
    make clean && \
    \
    # Clean source
    cd / && rm -rf /usr/src/php* /tmp/php_patch_version && \
    \
    # Smoke test
    php -v

# --- Set up PHP directory structure (matches ondrej/php layout) ---------------
RUN mkdir -p \
        "${PHP_FPM_CONF_DIR}/pool.d" \
        "${PHP_FPM_CONF_DIR}/conf.d" \
        "${PHP_CLI_CONF_DIR}/conf.d" \
        "${PHP_INI_DIR}/mods-available" && \
    \
    # Create FPM config
    if [ -f "${PHP_INI_DIR}/php-fpm.conf.default" ]; then \
        cp "${PHP_INI_DIR}/php-fpm.conf.default" "${PHP_FPM_CONF_DIR}/php-fpm.conf"; \
        echo "include=${PHP_FPM_CONF_DIR}/pool.d/*.conf" >> "${PHP_FPM_CONF_DIR}/php-fpm.conf"; \
    fi && \
    \
    # Versioned symlinks (so php-fpm8.3, php8.3, etc. work)
    ln -sf /usr/sbin/php-fpm "/usr/sbin/php-fpm${PHP_VERSION}" && \
    ln -sf /usr/bin/php "/usr/bin/php${PHP_VERSION}" && \
    ln -sf /usr/bin/phpize "/usr/bin/phpize${PHP_VERSION}" && \
    ln -sf /usr/bin/php-config "/usr/bin/php-config${PHP_VERSION}" && \
    ln -sf /usr/bin/phar.phar "/usr/bin/phar${PHP_VERSION}" && \
    ln -sf /usr/bin/phar.phar "/usr/bin/phar.phar${PHP_VERSION}"

# --- Extension installer helper -----------------------------------------------
COPY php-ext-install /usr/local/bin/php-ext-install
RUN chmod +x /usr/local/bin/php-ext-install

# --- Default extensions via PECL (APCu + Xdebug) -----------------------------
RUN pecl channel-update pecl.php.net && \
    # APCu
    pecl install -f apcu && \
    echo "extension=apcu.so" > "${PHP_INI_DIR}/mods-available/apcu.ini" && \
    ln -sf "${PHP_INI_DIR}/mods-available/apcu.ini" "${PHP_CLI_CONF_DIR}/conf.d/20-apcu.ini" && \
    ln -sf "${PHP_INI_DIR}/mods-available/apcu.ini" "${PHP_FPM_CONF_DIR}/conf.d/20-apcu.ini" && \
    \
    # Xdebug (installed but disabled by default)
    pecl install -f "xdebug-${XDEBUG_VERSION}" && \
    echo "zend_extension=xdebug.so" > "${PHP_INI_DIR}/mods-available/xdebug.ini" && \
    rm -f "${PHP_CLI_CONF_DIR}/conf.d/"*xdebug* \
          "${PHP_FPM_CONF_DIR}/conf.d/"*xdebug* 2>/dev/null || true

# --- User-requested extensions (via PHP_EXTENSIONS build arg) -----------------
# Usage in docker-compose.yml:
#   args:
#     PHP_EXTENSIONS: "redis imagick mongodb"
RUN if [ -n "${PHP_EXTENSIONS}" ]; then \
        php-ext-install ${PHP_EXTENSIONS}; \
    fi

# --- Composer ----------------------------------------------------------------
RUN curl -fsSL "https://getcomposer.org/download/${COMPOSER_VERSION}/composer.phar" \
        --output /usr/bin/composer && \
    chmod +x /usr/bin/composer

# --- n98-magerun2 ------------------------------------------------------------
RUN curl -fsSL https://files.magerun.net/n98-magerun2.phar \
        --output /usr/bin/n98-magerun2.phar && \
    chmod +x /usr/bin/n98-magerun2.phar

# --- Node.js + Grunt CLI -----------------------------------------------------
RUN curl -fsSL "https://deb.nodesource.com/setup_${NODE_MAJOR}.x" -o /tmp/nodesource_setup.sh && \
    bash /tmp/nodesource_setup.sh && \
    apt-get install -y nodejs && \
    npm install -g grunt-cli && \
    rm -f /tmp/nodesource_setup.sh && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# --- NVM for nginx user ------------------------------------------------------
RUN su - nginx -c "touch /home/nginx/.bash_profile" && \
    su - nginx -c "curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/master/install.sh | bash" && \
    chmod +x /home/nginx/.nvm/nvm.sh && \
    su - nginx -c "echo '[ -s /home/nginx/.nvm/nvm.sh ] && . /home/nginx/.nvm/nvm.sh' > .profile"

# --- mhsendmail (Mailpit-compatible) -----------------------------------------
RUN ARCH=$(dpkg --print-architecture) && \
    if [ "${ARCH}" = "amd64" ]; then \
        curl -fsSL "https://github.com/mailhog/mhsendmail/releases/download/v0.2.0/mhsendmail_linux_amd64" \
            --output /usr/local/bin/mhsendmail; \
    else \
        printf '#!/bin/sh\nexec /usr/sbin/sendmail "$@"\n' > /usr/local/bin/mhsendmail; \
    fi && \
    chmod +x /usr/local/bin/mhsendmail

# --- nginx user permissions ---------------------------------------------------
RUN mkdir -p /home/nginx/.composer && \
    chown -R nginx:nginx /home/nginx/.composer && \
    echo "nginx ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/nginx && \
    chmod 0440 /etc/sudoers.d/nginx

# --- Entrypoint ---------------------------------------------------------------
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/usr/local/bin/docker-entrypoint.sh"]
