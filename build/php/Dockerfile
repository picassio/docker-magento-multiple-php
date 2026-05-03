# ==============================================================================
# Unified PHP Dockerfile for Magento Development
# Supports PHP 7.0 – 8.4 via build arguments
# ==============================================================================

# --- Build arguments ---------------------------------------------------------
ARG BASE_IMAGE=ubuntu:24.04
FROM ${BASE_IMAGE}

ARG PHP_VERSION=8.3
ARG PHP_EXTENSIONS=""
ARG COMPOSER_VERSION=2.8.4
ARG NODE_MAJOR=20
ARG XDEBUG_VERSION=3.3.2

# Persist PHP_VERSION as env so the entrypoint script can read it at runtime
ENV PHP_VERSION=${PHP_VERSION}

# Prevent interactive prompts during build
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=UTC

# --- System setup: directories, user, base packages -------------------------
# Ubuntu 24.04 ships with uid/gid 1000 taken by 'ubuntu' user — remove it first
RUN if getent passwd 1000 >/dev/null 2>&1; then userdel -r $(getent passwd 1000 | cut -d: -f1) 2>/dev/null || true; fi && \
    if getent group 1000 >/dev/null 2>&1; then groupdel $(getent group 1000 | cut -d: -f1) 2>/dev/null || true; fi && \
    mkdir -p /home/public_html /run/php /var/lib/php/session && \
    groupadd -g 1000 nginx && \
    useradd -u 1000 -g nginx -m -d /home/nginx -s /bin/bash nginx && \
    chown -R nginx:nginx /home/public_html/ && \
    chown -R nginx:nginx /var/lib/php/

WORKDIR /home/public_html

RUN apt-get update -y && \
    apt-get install -y \
        software-properties-common \
        curl \
        git \
        sudo \
        apt-utils \
        vim \
        zip \
        unzip \
        tzdata \
        gnupg \
        ca-certificates && \
    # tini: available in noble (24.04) repos, needs manual install on bionic (18.04)
    (apt-get install -y tini 2>/dev/null || \
        (curl -fsSL https://github.com/krallin/tini/releases/download/v0.19.0/tini -o /usr/bin/tini && chmod +x /usr/bin/tini)) && \
    dpkg-reconfigure -f noninteractive tzdata && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# --- PHP repository + PHP installation --------------------------------------
# We avoid `add-apt-repository ppa:ondrej/php` because it calls the Launchpad
# REST API (getSigningKeyData) which frequently returns 504 Gateway Timeout.
# Instead we fetch the key from keyserver.ubuntu.com (reliable) and add the
# PPA source manually.
#
# WSL2 NOTE: Run `bin/mage doctor fix` first — it sets Docker MTU to match
# WSL2's NIC (1280), which is required for containers to reach Launchpad.
# See: https://github.com/microsoft/WSL/issues/5491
RUN curl -fsSL 'https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x14AA40EC0831756756D7F66C4F4EA0AAE5267A6C' \
        | gpg --dearmor -o /usr/share/keyrings/ondrej-php.gpg && \
    . /etc/os-release && \
    echo "deb [signed-by=/usr/share/keyrings/ondrej-php.gpg] https://ppa.launchpadcontent.net/ondrej/php/ubuntu ${VERSION_CODENAME} main" \
        > /etc/apt/sources.list.d/ondrej-php.list && \
    apt-get update -y

# Core PHP extensions (version-agnostic list)
# PHP 7.x needs the separate json package; PHP 8.x has it built-in.
RUN apt-get install -y \
        php${PHP_VERSION} \
        php${PHP_VERSION}-fpm \
        php${PHP_VERSION}-common \
        php${PHP_VERSION}-dev \
        php${PHP_VERSION}-apcu \
        php${PHP_VERSION}-gd \
        php${PHP_VERSION}-intl \
        php${PHP_VERSION}-soap \
        php${PHP_VERSION}-opcache \
        php${PHP_VERSION}-xml \
        php${PHP_VERSION}-mysql \
        php${PHP_VERSION}-mbstring \
        php${PHP_VERSION}-bcmath \
        php${PHP_VERSION}-curl \
        php${PHP_VERSION}-zip && \
    # Install json extension only for PHP 7.x (built-in on 8.x)
    PHP_MAJOR=$(echo ${PHP_VERSION} | cut -d. -f1) && \
    if [ "${PHP_MAJOR}" = "7" ]; then \
        apt-get install -y php${PHP_VERSION}-json; \
    fi && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Set PHP alternatives
RUN update-alternatives --set php /usr/bin/php${PHP_VERSION} && \
    update-alternatives --set phar /usr/bin/phar${PHP_VERSION} && \
    update-alternatives --set phar.phar /usr/bin/phar.phar${PHP_VERSION} && \
    update-alternatives --set phpize /usr/bin/phpize${PHP_VERSION} && \
    update-alternatives --set php-config /usr/bin/php-config${PHP_VERSION}

# --- Composer ----------------------------------------------------------------
RUN curl -fsSL https://getcomposer.org/download/${COMPOSER_VERSION}/composer.phar \
        --output /usr/bin/composer && \
    chmod +x /usr/bin/composer

# --- n98-magerun2 ------------------------------------------------------------
RUN curl -fsSL https://files.magerun.net/n98-magerun2.phar \
        --output /usr/bin/n98-magerun2.phar && \
    chmod +x /usr/bin/n98-magerun2.phar

# --- Node.js (via nodesource) + Grunt CLI ------------------------------------
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_MAJOR}.x -o /tmp/nodesource_setup.sh && \
    bash /tmp/nodesource_setup.sh && \
    apt-get install -y nodejs && \
    npm install -g grunt-cli && \
    rm -f /tmp/nodesource_setup.sh && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# --- NVM for nginx user ------------------------------------------------------
RUN su - nginx -c "touch /home/nginx/.bash_profile" && \
    su - nginx -c "curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/master/install.sh | bash" && \
    chmod +x /home/nginx/.nvm/nvm.sh && \
    su - nginx -c "echo '[ -s /home/nginx/.nvm/nvm.sh ] && . /home/nginx/.nvm/nvm.sh # This loads NVM' > .profile"

# --- mcrypt (PHP <= 8.0 only) -----------------------------------------------
# PHP 7.0/7.1 have php-mcrypt as a distro package; 7.2–8.0 need pecl.
RUN PHP_MAJOR=$(echo ${PHP_VERSION} | cut -d. -f1) && \
    PHP_MINOR=$(echo ${PHP_VERSION} | cut -d. -f2) && \
    if [ "${PHP_MAJOR}" = "7" ] && [ "${PHP_MINOR}" -le 1 ]; then \
        apt-get update -y && \
        apt-get install -y libmcrypt-dev php${PHP_VERSION}-mcrypt && \
        apt-get clean && rm -rf /var/lib/apt/lists/*; \
    elif [ "${PHP_MAJOR}" = "7" ] || ([ "${PHP_MAJOR}" = "8" ] && [ "${PHP_MINOR}" = "0" ]); then \
        apt-get update -y && \
        apt-get install -y libmcrypt-dev && \
        pecl install mcrypt && \
        echo "extension=mcrypt.so" > /etc/php/${PHP_VERSION}/mods-available/mcrypt.ini && \
        phpenmod -v ${PHP_VERSION} mcrypt && \
        apt-get clean && rm -rf /var/lib/apt/lists/*; \
    fi

# --- Xdebug (installed via pecl, starts disabled) ---------------------------
RUN pecl channel-update pecl.php.net && \
    pecl install -f xdebug-${XDEBUG_VERSION} && \
    # Create the ini file but don't enable it by default
    echo "zend_extension=xdebug.so" > /etc/php/${PHP_VERSION}/mods-available/xdebug.ini && \
    # Remove any auto-enabled symlinks so xdebug starts disabled
    rm -f /etc/php/${PHP_VERSION}/cli/conf.d/*-xdebug.ini \
          /etc/php/${PHP_VERSION}/fpm/conf.d/*-xdebug.ini || true

# --- mhsendmail (MailHog binary, compatible with Mailpit SMTP) ---------------
# The MailHog mhsendmail binary speaks standard SMTP and works with Mailpit.
# Only amd64 binary available; for other archs we create a simple wrapper script.
RUN ARCH=$(dpkg --print-architecture) && \
    if [ "${ARCH}" = "amd64" ]; then \
        curl -fsSL "https://github.com/mailhog/mhsendmail/releases/download/v0.2.0/mhsendmail_linux_amd64" \
            --output /usr/local/bin/mhsendmail; \
    else \
        printf '#!/bin/sh\nexec /usr/sbin/sendmail "$@"\n' > /usr/local/bin/mhsendmail; \
    fi && \
    chmod +x /usr/local/bin/mhsendmail

# --- Composer directory for nginx user ---------------------------------------
RUN mkdir -p /home/nginx/.composer && \
    chown -R nginx:nginx /home/nginx/.composer

# --- sudo for nginx user (passwordless) -------------------------------------
RUN echo "nginx ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/nginx && \
    chmod 0440 /etc/sudoers.d/nginx

# --- Extension installer helper -----------------------------------------------
COPY php-ext-install /usr/local/bin/php-ext-install
RUN chmod +x /usr/local/bin/php-ext-install

# --- User-requested extensions (via PHP_EXTENSIONS build arg) -----------------
# Usage in docker-compose.yml:
#   args:
#     PHP_EXTENSIONS: "redis imagick mongodb"
RUN if [ -n "${PHP_EXTENSIONS}" ]; then \
        php-ext-install ${PHP_EXTENSIONS}; \
    fi

# --- Entrypoint --------------------------------------------------------------
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/usr/local/bin/docker-entrypoint.sh"]
