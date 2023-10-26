#!/bin/bash

set -o allexport

source ./.env
source ./deployments/.env

export STACK_NAME=$(echo ${ROOT_DOMAIN} | sed 's/\./_/g')

echo "Logging into ECR..."
export REGISTRY_ENDPOINT=$(aws ecr get-authorization-token | jq -r '.authorizationData[0].proxyEndpoint' | sed 's/https:\/\///')
aws ecr get-login-password | docker login --username AWS --password-stdin $REGISTRY_ENDPOINT

docker run -it --rm \
	--name key_rotation \
	-v $(pwd)/secrets:/app/secrets \
	-e "KEYS_DIR=/app/secrets" \
	${REGISTRY_ENDPOINT}/${STACK_NAME}/key_rotation:latest \
	/app/generate $1

set +o allexport
