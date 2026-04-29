# ==============================================================================
# Unified PHP Dockerfile for Magento Development
# Supports PHP 7.0 – 8.4 via build arguments
# ==============================================================================

# --- Build arguments ---------------------------------------------------------
ARG BASE_IMAGE=ubuntu:24.04
FROM ${BASE_IMAGE}

ARG PHP_VERSION=8.3
ARG COMPOSER_VERSION=2.8.4
ARG NODE_MAJOR=20
ARG XDEBUG_VERSION=3.3.2

# Persist PHP_VERSION as env so the entrypoint script can read it at runtime
ENV PHP_VERSION=${PHP_VERSION}

# Prevent interactive prompts during build
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=UTC

# --- System setup: directories, user, base packages -------------------------
RUN mkdir -p /home/public_html /run/php /var/lib/php/session && \
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
        tini \
        tzdata \
        gnupg \
        ca-certificates && \
    dpkg-reconfigure -f noninteractive tzdata && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# --- PHP repository + PHP installation --------------------------------------
RUN LC_ALL=C.UTF-8 add-apt-repository ppa:ondrej/php && \
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

# --- mhsendmail (Mailpit-compatible) ----------------------------------------
RUN ARCH=$(dpkg --print-architecture) && \
    curl -fsSL "https://github.com/axllent/mailpit/releases/latest/download/mhsendmail_linux_${ARCH}" \
        --output /usr/local/bin/mhsendmail && \
    chmod +x /usr/local/bin/mhsendmail

# --- Composer directory for nginx user ---------------------------------------
RUN mkdir -p /home/nginx/.composer && \
    chown -R nginx:nginx /home/nginx/.composer

# --- sudo for nginx user (passwordless) -------------------------------------
RUN echo "nginx ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/nginx && \
    chmod 0440 /etc/sudoers.d/nginx

# --- Entrypoint --------------------------------------------------------------
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["/usr/local/bin/docker-entrypoint.sh"]
