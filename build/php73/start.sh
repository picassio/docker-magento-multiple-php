#! /bin/bash
set -e 

sed -i -e "s/REPLACE_WITH_REAL_KEY/${NEW_RELIC_LICENSE_KEY}/" \
    -e "s/newrelic.appname[[:space:]]=[[:space:]].*/newrelic.appname=\"${NEW_RELIC_APPNAME}\"/" \
    -e '$anewrelic.distributed_tracing_enabled=true' \
    -e '$anewrelic.daemon.address="newrelic:31339"' \
    /etc/php/7.3/fpm/conf.d/newrelic.ini

sed -i -e "s/REPLACE_WITH_REAL_KEY/${NEW_RELIC_LICENSE_KEY}/" \
    -e "s/newrelic.appname[[:space:]]=[[:space:]].*/newrelic.appname=\"${NEW_RELIC_APPNAME}\"/" \
    -e '$anewrelic.distributed_tracing_enabled=true' \
    -e '$anewrelic.daemon.address="newrelic:31339"' \
    /etc/php/7.3/cli/conf.d/newrelic.ini
    
/usr/sbin/php-fpm7.3 -F -O --fpm-config /etc/php/7.3/fpm/php-fpm.conf