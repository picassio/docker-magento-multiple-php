#!/bin/bash
[ -z "$1" ] && echo "Please specify a PHP container to go into (ex. php74)" && exit
docker-compose exec --user nginx "$@" bash