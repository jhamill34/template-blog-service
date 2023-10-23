#!/bin/sh

mysql -u root -p$MYSQL_ROOT_PASSWORD <<EOF
create user '${DB_APP_USER}'@'%' identified by '${INITIAL_DB_APP_PASSWORD}';
EOF

mysql -u root -p$MYSQL_ROOT_PASSWORD <<EOF
create user '${DB_AUTH_USER}'@'%' identified by '${INITIAL_DB_AUTH_PASSWORD}';
EOF