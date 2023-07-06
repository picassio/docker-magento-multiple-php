FROM ubuntu:focal

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
	apt-get -y install php8.1 php8.1-dev php8.1-apcu php8.1-fpm php8.1-common php8.1-gd php8.1-intl php8.1-soap php8.1-opcache php8.1-xml php8.1-mysql php8.1-mbstring php8.1-bcmath php8.1-curl php8.1-zip php8.1-xdebug

RUN update-alternatives --set php /usr/bin/php8.1 \
	&& update-alternatives --set phar /usr/bin/phar8.1 \
	&& update-alternatives --set phar.phar /usr/bin/phar.phar8.1 \
	&& update-alternatives --set phpize /usr/bin/phpize8.1 \
	&& update-alternatives --set php-config /usr/bin/php-config8.1
	
RUN export DEBIAN_FRONTEND=noninteractive && apt-get install -y curl git sudo apt-utils vim zip unzip
RUN dpkg-reconfigure -f noninteractive tzdata

RUN curl https://getcomposer.org/download/2.1.12/composer.phar --output /usr/bin/composer && chmod +x /usr/bin/composer
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
RUN unlink /etc/php/8.1/cli/conf.d/20-xdebug.ini \
	&& unlink /etc/php/8.1/fpm/conf.d/20-xdebug.ini

RUN pecl channel-update pecl.php.net
RUN export DEBIAN_FRONTEND=noninteractive\ 
 		&& apt-get install libmcrypt-dev -y
RUN pecl install mcrypt
RUN pecl install -f xdebug-3.1.5

RUN curl -L https://github.com/mailhog/mhsendmail/releases/download/v0.2.0/mhsendmail_linux_amd64 --output /usr/local/bin/mhsendmail \
	&& chmod +x /usr/local/bin/mhsendmail

RUN mkdir /home/nginx/.composer && chown -R nginx:nginx /home/nginx/.composer

RUN echo "nginx ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/nginx

ENTRYPOINT ["/tini", "--"]
CMD ["/usr/sbin/php-fpm8.1", "-F", "-O", "--fpm-config", "/etc/php/8.1/fpm/php-fpm.conf"]
