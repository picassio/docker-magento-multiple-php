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
	apt-get -y install php7.3 php7.3-apcu php7.3-fpm php7.3-common php7.3-gd php7.3-intl php7.3-soap php7.3-opcache php7.3-xml php7.3-mysql php7.3-mbstring php7.3-bcmath php7.3-json php7.3-curl php7.3-zip

RUN export DEBIAN_FRONTEND=noninteractive && apt-get install -y curl git sudo apt-utils vim zip unzip
RUN dpkg-reconfigure -f noninteractive tzdata

RUN curl https://getcomposer.org/download/1.10.19/composer.phar --output /usr/bin/composer && chmod +x /usr/bin/composer

RUN mkdir /home/nginx/.composer && chown -R nginx:nginx /home/nginx/.composer

ENTRYPOINT ["/tini", "--"]
CMD ["/usr/sbin/php-fpm7.3", "-F", "-O", "--fpm-config", "/etc/php/7.3/fpm/php-fpm.conf"]
