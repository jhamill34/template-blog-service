set -o allexport 

source ./deployments/.env

stack_name=$(echo ${ROOT_DOMAIN} | sed 's/\./_/g')
docker stack deploy -c ./deployments/docker-compose.yaml $stack_name

set +o allexport

