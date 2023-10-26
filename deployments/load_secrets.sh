#/bin/bash 

set -o allexport 

source ./secrets/env set

echo $DB_ROOT_PASSWORD | docker secret create db_root_password -
echo $DB_AUTH_PASSWORD | docker secret create db_auth_password -
echo $DB_APP_PASSWORD | docker secret create db_app_password -
echo $DB_MIGRATOR_PASSWORD | docker secret create db_migrator_password -
echo $CACHE_PASSWORD | docker secret create cache_password -
echo $DEFAULT_OAUTH_CLIENT_ID | docker secret create oauth_client_id -
echo $DEFAULT_OAUTH_CLIENT_SECRET | docker secret create oauth_client_secret -
echo $AUTH_TOKEN_SECRET | docker secret create auth_token_secret -
echo $AUTH_SESSION_SECRET | docker secret create auth_session_secret -
echo $APP_SESSION_SECRET | docker secret create app_session_secret -

set +o allexport 

docker secret create access_token_public_key ./secrets/token.pem
docker secret create access_token_private_key ./secrets/token-key.pem

docker secret create dkim_private_key ./secrets/dkim-key.pem

docker secret create ssl_public_key ./secrets/sslcert.pem
docker secret create ssl_private_key ./secrets/sslkey.pem

