FROM ubuntu:focal

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

RUN mkdir -p /home/public_html /run/php /var/lib/php/session &&\
	groupadd -g 1000 nginx && useradd -u 1000 -g nginx -m -d /home/nginx -s /bin/bash nginx &&\
	chown -R nginx:nginx /home/public_html/	&&\
	chown -R nginx:nginx /var/lib/php/

WORKDIR /home/public_html

RUN export DEBIAN_FRONTEND=noninteractive &&\
	apt-get update -y &&\
	apt-get install apt-utils sudo vim git zip unzip -y &&\
	apt-get -y install php7.2 php7.2-dev php7.2-apcu php7.2-fpm php7.2-common php7.2-gd php7.2-intl php7.2-soap php7.2-opcache php7.2-xml php7.2-mysql php7.2-mbstring php7.2-bcmath php7.2-json php7.2-curl php7.2-zip

RUN update-alternatives --set php /usr/bin/php7.2 \
	&& update-alternatives --set phar /usr/bin/phar7.2 \
	&& update-alternatives --set phar.phar /usr/bin/phar.phar7.2 \
	&& update-alternatives --set phpize /usr/bin/phpize7.2 \
	&& update-alternatives --set php-config /usr/bin/php-config7.2
	
RUN ln -fs /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime && dpkg-reconfigure -f noninteractive tzdata

RUN apt-get install -y curl
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

ENTRYPOINT ["/tini", "--"]
CMD ["/usr/sbin/php-fpm7.2", "-F", "-O", "--fpm-config", "/etc/php/7.2/fpm/php-fpm.conf"]
