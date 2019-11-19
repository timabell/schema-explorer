#!/bin/bash
set -e


DB_HOST=${MYSQL_HOST:-127.0.0.1}
DB_USER=${MYSQL_ROOT_USER:-root}
DB_PASSWD=${MYSQL_ROOT_PASSWORD:-supersecret}

echo MYSQL_ROOT_USER=$MYSQL_ROOT_USER
echo MYSQL_HOST=$MYSQL_HOST
echo MYSQL_OPTS=$MYSQL_OPTS

export DB_USER
export DB_PASSWD
export DB

echo Waiting for mysql...
dockerize -wait tcp://localhost:3306 -timeout 1m

cd testdata/container/mysql
chmod +x *.sh
./setup-mysql-user.sh
./setup-mysql-db.sh

SETUP_DB=ssetestusr ./setup-mysql-db.sh



# for f in ./testdata/container/mysql/*.sql; do
#     echo Running $f
#     mysql $MYSQL_OPTS -u $USER -p$PASSWD -h $HOST < $f
# done


