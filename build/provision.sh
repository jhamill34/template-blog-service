#!/bin/bash

set -o allexport

source ./.env
source ./build/.env

stack_name=$(echo ${ROOT_DOMAIN} | sed 's/\./_/g')

echo "Creating ECR repositories for ${stack_name}..."

services=( app_service gateway_service auth_service mailer migrator database cache proxy key_rotation )

for service in "${services[@]}"; do 
	echo "${stack_name}/${service}"
	result=$(aws ecr create-repository --repository-name "${stack_name}/${service}" 2>/dev/null)
	ret=$?
	if [ $ret -eq 0 ]; then
		uri=$(echo $result | jq -r '.repository.repositoryUri')
		echo "Created ${uri}"
	else 
		uri=$(aws ecr describe-repositories --repository-names "${stack_name}/${service}" | jq -r '.repositories[0].repositoryUri')
		echo "Already exists ${uri}"
	fi
	echo ""
done


set +o allexport

