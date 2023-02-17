#! /bin/bash
set -e

RUN sed -i -e "s/REPLACE_WITH_REAL_KEY/${NEW_RELIC_LICENSE_KEY}/" \
    -e "s/newrelic.appname[[:space:]]=[[:space:]].*/newrelic.appname=\"${NEW_RELIC_APPNAME}\"/" \
    -e '$anewrelic.distributed_tracing_enabled=true' \
    -e '$anewrelic.daemon.address="newrelic:31339"' \
    /etc/php/7.3/fpm/conf.d/newrelic.ini

RUN sed -i -e "s/REPLACE_WITH_REAL_KEY/${NEW_RELIC_LICENSE_KEY}/" \
    -e "s/newrelic.appname[[:space:]]=[[:space:]].*/newrelic.appname=\"${NEW_RELIC_APPNAME}\"/" \
    -e '$anewrelic.distributed_tracing_enabled=true' \
    -e '$anewrelic.daemon.address="newrelic:31339"' \
    /etc/php/7.3/cli/conf.d/newrelic.ini
    
exec $@