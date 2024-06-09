#!/bin/bash

touch .env
echo "DBUSER=$1">.env
echo "DBPASS=$2">>.env

#setting the env
if [ -f ./.env ]; then
    set -a
    source ./.env
    set +a
else
    echo ".env file not found!"
    exit 1
fi

# stacking the compose in the swarm
docker stack deploy -c docker-compose.yml llama_service

#rolling update image of minitwit
docker service update --image mihr/minitwitimage:latest llama_service_minitwit 