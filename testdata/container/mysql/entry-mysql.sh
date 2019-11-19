#!/bin/bash
set -e

DB_USER=${MYSQL_ROOT_USER:-root}
DB_PASSWD=${DB_PASSWD:-$MYSQL_ROOT_PASSWORD}
DB=${MYSQL_DATABASE:-mysql}

export DB_USER
export DB_PASSWD
export DB

cd /usr/src/app/
./setup-mysql-user.sh
./setup-mysql-db.sh

SETUP_DB=ssetestusr ./setup-mysql-db.sh