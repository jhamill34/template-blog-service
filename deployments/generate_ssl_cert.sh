#!/bin/bash
#
set -o allexport

source ./deployments/.env

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
	key_rotation:latest \
	/app/certs

set +o allexport

