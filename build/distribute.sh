#!/bin/bash

set -o allexport

source ./.env
source ./build/.env

echo "Logging into ECR..."
registryEndpoint=$(aws ecr get-authorization-token | jq -r '.authorizationData[0].proxyEndpoint')
aws ecr get-login-password | docker login --username AWS --password-stdin $registryEndpoint

stack_name=$(echo ${ROOT_DOMAIN} | sed 's/\./_/g')

services=( app_service gateway_service auth_service mailer migrator database cache proxy key_rotation )

for service in "${services[@]}"; do 
	result=$(aws ecr describe-repositories --repository-names "${stack_name}/${service}" 2> /dev/null)
	ret=$?
	if [ $ret -eq 0 ]; then
		uri=$(echo $result | jq -r '.repositories[0].repositoryUri')

		echo "Tagging ${service} with ${uri}:latest ..."
		docker tag "${service}:latest" "${uri}:latest"

		echo "Pushing to ${uri}:latest ..."
		docker push "${uri}:latest"
	else
		echo "ECR Repository doesn't exist"
	fi
done
