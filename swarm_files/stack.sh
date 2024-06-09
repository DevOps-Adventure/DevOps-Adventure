#!/bin/bash

touch .env
echo "DBUSER=$1">.env
echo "DBPASS=$2">>.env

#setting the env
set -a; . ./.env; set +a

# stacking the compose in the swarm
docker stack deploy -c docker-compose.yml stack_name

#rolling update image of minitwit
docker service update --image mihr/minitwitimage:latest stack_name_minitwit 