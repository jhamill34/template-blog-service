#!/bin/sh

DB_APP_PASSWORD=$(cat $DB_APP_PASSWORD_FILE)
mysql -u root -p$MYSQL_ROOT_PASSWORD <<EOF
create user '${DB_APP_USER}'@'%' identified by '${DB_APP_PASSWORD}';
EOF

DB_AUTH_PASSWORD=$(cat $DB_AUTH_PASSWORD_FILE)
mysql -u root -p$MYSQL_ROOT_PASSWORD <<EOF
create user '${DB_AUTH_USER}'@'%' identified by '${DB_AUTH_PASSWORD}';
EOF

DB_MIGRATOR_PASSWORD=$(cat $DB_MIGRATOR_PASSWORD_FILE)
mysql -u root -p$MYSQL_ROOT_PASSWORD <<EOF
create user '${DB_MIGRATOR_USER}'@'%' identified by '${DB_MIGRATOR_PASSWORD}';

grant all on ${MYSQL_DATABASE}.* to '${DB_MIGRATOR_USER}' with grant option;

flush privileges;
EOF


