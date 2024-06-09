#!/bin/bash

cd /minitwit || exit

touch .env
echo "DBUSER=$1">.env
echo "DBPASS=$2">>.env

echo "docker compose down call"
docker compose -f docker-compose.yml down
sleep 10s

echo "removing mihr/minitwitimage"
docker image rm mihr/minitwitimage

echo "pulling minitwit new image"
docker pull mihr/minitwitimage

echo "starting the docker compose"
docker compose -f docker-compose.yml up -d --remove-orphans