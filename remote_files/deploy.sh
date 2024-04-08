#!/bin/bash

cd /minitwit || exit

touch .env
echo "DBUSER=$1">.env #only one '>' = overwrite .env if existing
echo "DBPASS=$2">>.env


docker compose -f docker-compose.yml pull
docker compose -f docker-compose.yml up -d --remove-orphans
#docker pull $DOCKER_USERNAME/flagtoolimage:latest