FROM ubuntu:bionic

ENV DEBIAN_FRONTEND=noninteractive
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

RUN apt-get update -y &&\
	apt-get install apt-utils tzdata curl gnupg apt-transport-https -y
RUN apt-get install debian-archive-keyring -y

RUN curl -s https://packagecloud.io/install/repositories/varnishcache/varnish60lts/script.deb.sh | bash

RUN apt-get update -y

RUN apt-get install varnish -y

ENTRYPOINT ["/tini", "--"]
CMD ["/usr/sbin/varnishd", "-F", "-f", "/etc/varnish/default.vcl", "-a", "0.0.0.0:6081", "-p", "feature=+http2", "-s", "malloc,256m"]