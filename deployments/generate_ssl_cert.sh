#!/bin/bash
#
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
	-v $(pwd)/configs:/app/configs \
	-v $(pwd)/.letsencrypt:/app/data \
	-e "CONFIG_FILE=/app/configs/certs.yaml" \
	-e "OWNER_INFO_FILE=/app/data/owner.yaml" \
	-e "ACCOUNT_KEY_PATH=/app/data/account-key.pem" \
	-e "KEYS_DIR=/app/secrets" \
	-e "ROOT_DOMAIN=${ROOT_DOMAIN}" \
	${REGISTRY_ENDPOINT}/${STACK_NAME}/key_rotation:latest \
	/app/certs

set +o allexport

