#!/bin/bash

cd /minitwit || exit

touch .env
echo "DBUSER=$1">.env
echo "DBPASS=$2">>.env

docker compose -f docker-compose.yml down
sleep 10s
docker compose -f docker-compose.yml pull
docker compose -f docker-compose.yml up -d --remove-orphans
