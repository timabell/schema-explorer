#!/bin/bash
set -e
echo ---- Starting $0

DB_USER=${DB_USER:-root}
DB_PASSWD=${DB_PASSWD:-omgroot}
DB=${DB:-mysql}
DB_OPTS=${DB_OPTS:--u $DB_USER -p$DB_PASSWD $MYSQL_OPTS}

SETUP_DB=${SETUP_DB:-ssetest}
SETUP_USER=${SETUP_USER:-ssetestusr}
SETUP_PASSWD=${SETUP_PASSWD:-ssetestusrpass}
SETUP_DB_OPTS=${SETUP_DB_OPTS:--u $SETUP_USER -p$SETUP_PASSWD $MYSQL_OPTS $SETUP_DB}
SETUP_DB_SCRIPT=${SETUP_DB_SCRIPT:-test-db.sql}

echo SETUP_DB=$SETUP_DB
echo SETUP_USER=$SETUP_USER
echo SETUP_PASSWD=$SETUP_PASSWD
echo SETUP_DB_SCRIPT=$SETUP_DB_SCRIPT
echo DB_OPTS=$DB_OPTS

mysql $DB_OPTS <<-EOSQL
    drop database if exists $SETUP_DB;
    create database $SETUP_DB;
EOSQL

mysql $SETUP_DB_OPTS < $SETUP_DB_SCRIPT

echo ---- Done with $0