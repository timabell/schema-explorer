#!/bin/bash
set -e

echo POSTGRES_USER: $POSTGRES_USER
echo POSTGRES_DB: $POSTGRES_DB
echo PG_URL: $PG_URL

echo Waiting for postgres...
dockerize -wait tcp://localhost:5432 -timeout 1m

cd testdata/container/pg
chmod +x *.sh
./setup-pg-user.sh

./setup-pg-db.sh
SETUP_DB=ssetestusr ./setup-pg-db.sh