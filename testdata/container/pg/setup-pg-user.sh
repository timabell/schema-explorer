#!/bin/bash
set -e
echo ---- Starting $0

USER=${DB_USER:-postgres}
DB=${DB:-postgres}

DB_OPTS=${PG_OPTS:--d $PG_URL$DB}

SETUP_DB=${SETUP_DB:-ssetest}
SETUP_USER=${SETUP_USER:-ssetestusr}
SETUP_PASSWD=${SETUP_PASSWD:-ssetestusrpass}


echo SETUP_DB=$SETUP_DB
echo SETUP_USER=$SETUP_USER
echo SETUP_PASSWD=$SETUP_PASSWD
echo DB_OPTS=$DB_OPTS



psql -v ON_ERROR_STOP=1 $DB_OPTS <<-EOSQL
    DROP USER IF EXISTS $SETUP_USER;
    CREATE USER $SETUP_USER;
    ALTER USER $SETUP_USER with password '$SETUP_PASSWD';
EOSQL

echo ---- Done with $0

