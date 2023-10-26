#!/bin/bash

docker run -it --rm \
	--name key_rotation \
	-v $(pwd)/secrets:/app/secrets \
	-e "KEYS_DIR=/app/secrets" \
	key_rotation:latest \
	/app/generate $1

