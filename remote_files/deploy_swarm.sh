#!/bin/bash

# Ensure you are using bash
cd ..
cd go-minitwit

# Load environment variables
set -a
source .env
set +a

# Export required variables
export DIGITALOCEAN_PRIVATE_NETWORKING=true
export DROPLETS_API="https://api.digitalocean.com/v2/droplets"
export BEARER_AUTH_TOKEN="Authorization: Bearer $DIGITAL_OCEAN_TOKEN"
export JSON_CONTENT="Content-Type: application/json"

# Define your JSON configuration
CONFIG='{"name":"swarm-manager-test","tags":["demo"],
	"size":"s-2vcpu-4gb", "image":"docker-20-04",
	"ssh_keys":["e4:dd:bf:3e:07:e4:a8:be:0f:2a:b1:5e:9b:ae:43:dd"]}'

# Perform the API request and capture the response
RESPONSE=$(curl -s -X POST "$DROPLETS_API" \
    -d "$CONFIG" \
    -H "$BEARER_AUTH_TOKEN" \
    -H "$JSON_CONTENT")

# Extract droplet ID
SWARM_MANAGER_ID=$(echo "$RESPONSE" | jq -r '.droplet.id')

# Check for null response and provide feedback
if [[ "$SWARM_MANAGER_ID" == "null" ]]; then
    ERROR_MSG=$(echo "$RESPONSE" | jq -r '.message // "Unknown error"')
    echo "Error: $ERROR_MSG"
else
    sleep 5
    echo "Swarm Manager ID: $SWARM_MANAGER_ID"
fi

export JQFILTER='.droplets | .[] | select (.name == "swarm-manager-test") 
	| .networks.v4 | .[]| select (.type == "public") | .ip_address'

sleep 90

SWARM_MANAGER_IP=$(curl -s GET "$DROPLETS_API"\
    -H "$BEARER_AUTH_TOKEN" -H "$JSON_CONTENT"\
    | jq -r "$JQFILTER") && echo "SWARM_MANAGER_IP=$SWARM_MANAGER_IP"
