set -o allexport 

source ./.env
source ./deployments/.env

export STACK_NAME=$(echo ${ROOT_DOMAIN} | sed 's/\./_/g')

echo "Logging into ECR..."
export REGISTRY_ENDPOINT=$(aws ecr get-authorization-token | jq -r '.authorizationData[0].proxyEndpoint' | sed 's/https:\/\///')
aws ecr get-login-password | docker login --username AWS --password-stdin $REGISTRY_ENDPOINT

docker stack deploy -c ./deployments/docker-compose.yaml $STACK_NAME --with-registry-auth

set +o allexport

