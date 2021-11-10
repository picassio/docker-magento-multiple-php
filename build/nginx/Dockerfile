FROM ubuntu:bionic

ENV DEBIAN_FRONTEND noninteractive

# nginx.conf
#error_log  stderr warn;
#access_log  /dev/stdout main;

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

RUN mkdir -p /home/public_html /run/php &&\
	groupadd -g 1000 nginx && useradd -u 1000 -g nginx -M -d /home/nginx -s /bin/false nginx &&\
	chown -R nginx:nginx /home/public_html/

WORKDIR /home/public_html

RUN export DEBIAN_FRONTEND=noninteractive
RUN apt-get update -y &&\
	apt-get install apt-utils tzdata -y

RUN apt-get -y install nginx

RUN ln -fs /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime && dpkg-reconfigure -f noninteractive tzdata

ADD conf/nginx.conf /etc/nginx/nginx.conf

ENTRYPOINT ["/tini", "--"]
CMD ["nginx", "-g", "daemon off;"]