ARG NEW_RELIC_AGENT_VERSION
ARG NEW_RELIC_LICENSE_KEY
ARG NEW_RELIC_APPNAME

FROM ubuntu:bionic

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

RUN mkdir -p /home/public_html /run/php /var/lib/php/session &&\
	groupadd -g 1000 nginx && useradd -u 1000 -g nginx -m -d /home/nginx -s /bin/bash nginx &&\
	chown -R nginx:nginx /home/public_html/	&&\
	chown -R nginx:nginx /var/lib/php/

WORKDIR /home/public_html

RUN export DEBIAN_FRONTEND=noninteractive && apt-get update -y
RUN export DEBIAN_FRONTEND=noninteractive && apt-get install -y software-properties-common
RUN export DEBIAN_FRONTEND=noninteractive && LC_ALL=C.UTF-8 add-apt-repository ppa:ondrej/php && apt-get update -y
RUN export DEBIAN_FRONTEND=noninteractive &&\
	apt-get -y install php7.3 php7.3-dev php7.3-apcu php7.3-fpm php7.3-common php7.3-gd php7.3-intl php7.3-soap php7.3-opcache php7.3-xml php7.3-mysql php7.3-mbstring php7.3-bcmath php7.3-json php7.3-curl php7.3-zip

RUN update-alternatives --set php /usr/bin/php7.3 \
	&& update-alternatives --set phar /usr/bin/phar7.3 \
	&& update-alternatives --set phar.phar /usr/bin/phar.phar7.3 \
	&& update-alternatives --set phpize /usr/bin/phpize7.3 \
	&& update-alternatives --set php-config /usr/bin/php-config7.3
	
RUN export DEBIAN_FRONTEND=noninteractive && apt-get install -y curl git sudo apt-utils vim zip unzip
RUN dpkg-reconfigure -f noninteractive tzdata

RUN curl https://getcomposer.org/download/1.10.19/composer.phar --output /usr/bin/composer && chmod +x /usr/bin/composer
RUN curl https://files.magerun.net/n98-magerun2.phar --output /usr/bin/n98-magerun2.phar && chmod +x /usr/bin/n98-magerun2.phar

RUN export DEBIAN_FRONTEND=noninteractive \ 
	&& curl -sL https://deb.nodesource.com/setup_14.x -o nodesource_setup.sh \
    && sudo bash nodesource_setup.sh \
    && apt-get install -y nodejs \
    && apt-get update \
    && npm install -g grunt-cli

RUN su - nginx -c "touch /home/nginx/.bash_profile"
RUN su - nginx -c "curl https://raw.githubusercontent.com/nvm-sh/nvm/master/install.sh | bash"
RUN chmod +x /home/nginx/.nvm/nvm.sh
RUN su - nginx -c "echo '[ -s /home/nginx/.nvm/nvm.sh ] && . /home/nginx/.nvm/nvm.sh # This loads NVM' > .profile"

RUN pecl channel-update pecl.php.net
RUN export DEBIAN_FRONTEND=noninteractive\ 
		&& apt-get install libmcrypt-dev -y
RUN pecl install mcrypt
RUN pecl install -f xdebug-2.9.8

RUN curl -L https://github.com/mailhog/mhsendmail/releases/download/v0.2.0/mhsendmail_linux_amd64 --output /usr/local/bin/mhsendmail \
	&& chmod +x /usr/local/bin/mhsendmail

RUN mkdir /home/nginx/.composer && chown -R nginx:nginx /home/nginx/.composer

RUN echo "nginx ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/nginx

ENV NEW_RELIC_AGENT_VERSION ${NEW_RELIC_AGENT_VERSION}
ENV NEW_RELIC_LICENSE_KEY ${NEW_RELIC_LICENSE_KEY}
ENV NEW_RELIC_APPNAME ${NEW_RELIC_APPNAME}

RUN curl -L https://download.newrelic.com/php_agent/archive/${NEW_RELIC_AGENT_VERSION}/newrelic-php5-${NEW_RELIC_AGENT_VERSION}-linux.tar.gz | tar -C /tmp -zx \
    && export NR_INSTALL_USE_CP_NOT_LN=1 \
    && export NR_INSTALL_SILENT=1 \
    && /tmp/newrelic-php5-${NEW_RELIC_AGENT_VERSION}-linux/newrelic-install install \
    && rm -rf /tmp/newrelic-php5-* /tmp/nrinstall*

RUN sed -i -e "s/REPLACE_WITH_REAL_KEY/${NEW_RELIC_LICENSE_KEY}/" \
    -e "s/newrelic.appname[[:space:]]=[[:space:]].*/newrelic.appname=\"${NEW_RELIC_APPNAME}\"/" \
    -e '$anewrelic.distributed_tracing_enabled=true' \
    -e '$anewrelic.daemon.address="newrelic:31339"' \
    /etc/php/7.3/fpm/conf.d/newrelic.ini

ENTRYPOINT ["/tini", "--"]
CMD ["/usr/sbin/php-fpm7.3", "-F", "-O", "--fpm-config", "/etc/php/7.3/fpm/php-fpm.conf"]
