FROM ubuntu:bionic

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

RUN mkdir -p /home/public_html /run/php /var/lib/php/session &&\
	groupadd -g 1000 nginx && useradd -u 1000 -g nginx -M -d /home/public_html -s /bin/bash nginx &&\
	chown -R nginx:nginx /home/public_html/	&&\
	chown -R nginx:nginx /var/lib/php/

WORKDIR /home/public_html

RUN export DEBIAN_FRONTEND=noninteractive &&\
	apt-get update -y &&\
	apt-get install apt-utils sudo vim  zip unzip -y &&\
	apt-get -y install php7.2 php7.2-apcu php7.2-fpm php7.2-common php7.2-gd php7.2-intl php7.2-soap php7.2-opcache php7.2-xml php7.2-mysql php7.2-mbstring php7.2-bcmath php7.2-json php7.2-curl php7.2-zip

RUN ln -fs /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime && dpkg-reconfigure -f noninteractive tzdata

RUN apt-get install -y curl
RUN curl https://getcomposer.org/download/1.10.19/composer.phar --output /usr/bin/composer && chmod +x /usr/bin/composer
#RUN wget https://getcomposer.org/download/1.10.19/composer.phar -O /usr/bin/composer && chmod +x /usr/bin/composer

ENTRYPOINT ["/tini", "--"]
CMD ["/usr/sbin/php-fpm7.2", "-F", "-O", "--fpm-config", "/etc/php/7.2/fpm/php-fpm.conf"]
