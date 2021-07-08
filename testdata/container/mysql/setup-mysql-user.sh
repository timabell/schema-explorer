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

echo SETUP_DB=$SETUP_DB
echo SETUP_USER=$SETUP_USER
echo SETUP_PASSWD=$SETUP_PASSWD
echo DB_OPTS=$DB_OPTS

mysql $DB_OPTS <<-EOSQL
    drop user if exists '$SETUP_USER'@'%';
    CREATE USER '$SETUP_USER'@'%' IDENTIFIED BY '$SETUP_PASSWD';
    GRANT ALL PRIVILEGES ON *.* TO '$SETUP_USER'@'%' WITH GRANT OPTION;
    GRANT RELOAD,PROCESS ON *.* TO '$SETUP_USER'@'%';
    FLUSH PRIVILEGES;
    select user, host from mysql.user where user = '$SETUP_USER';
EOSQL

echo ---- Done with $0

