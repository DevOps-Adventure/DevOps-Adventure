#!/bin/bash

touch .env
echo "DBUSER=$1">.env
echo "DBPASS=$2">>.env

#setting the env
# Checks if the .env file exists.
if [ -f ./.env ]; then
    # Sets the -a option in set, which means that subsequent variable assignments (source ./.env) will be exported automatically to the environment of subsequently executed commands.
    set -a 
    source ./.env
    # Disables automatic exporting of variables.
    set +a
else
    echo ".env file not found!"
    exit 1
fi

# Deploys the Docker stack named llama_service using the specified Compose file (docker-compose.yml)
docker stack deploy -c docker-compose.yml llama_service

# Updates the Docker service named llama_service_minitwit to use the latest version of the mihr/minitwitimage Docker image.
docker service update --image mihr/minitwitimage:latest llama_service_minitwit 