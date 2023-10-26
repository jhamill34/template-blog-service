#!/bin/bash 

docker build . -f ./build/service/Dockerfile --build-arg "SERVICE=app" -t app_service:latest 
docker build . -f ./build/service/Dockerfile --build-arg "SERVICE=gateway" -t gateway_service:latest 
docker build . -f ./build/service/Dockerfile --build-arg "SERVICE=auth" -t auth_service:latest 
docker build . -f ./build/service/Dockerfile --build-arg "SERVICE=mail" -t mailer:latest 
docker build . -f ./build/service/Dockerfile --build-arg "SERVICE=migrator" -t migrator:latest 

docker build . -f ./build/database/Dockerfile -t database:latest
docker build . -f ./build/cache/Dockerfile -t cache:latest
docker build . -f ./build/proxy/Dockerfile -t proxy:latest

docker build . -f ./build/key_rotation/Dockerfile -t key_rotation:latest

