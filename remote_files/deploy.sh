#!/bin/bash

# Changes the working directory to /minitwit. If the directory doesn't exist, it exits the script.
cd /minitwit || exit

# Creates a .env file and writes the provided database credentials (DBUSER and DBPASS) to it.
touch .env
echo "DBUSER=$1">.env
echo "DBPASS=$2">>.env

# Stops and removes the containers defined in the docker-compose.yml file using the docker compose down command.
echo "docker compose down call"
docker compose -f docker-compose.yml down

# Wait for 10 seconds (sleep 10s) to allow time for the containers to stop gracefully.
sleep 10s

# Removes the Docker image mihr/minitwitimage using the docker image rm command.
echo "removing mihr/minitwitimage"
docker image rm mihr/minitwitimage

# Pulls the latest version of the mihr/minitwitimage Docker image from the Docker registry using the docker pull command.
echo "pulling minitwit new image"
docker pull mihr/minitwitimage

# Starts the Docker services defined in the docker-compose.yml file in detached mode (docker compose -f docker-compose.yml up -d --remove-orphans), 
# - removing any orphaned containers.
echo "starting the docker compose"
docker compose -f docker-compose.yml up -d --remove-orphans