#!/bin/bash

LOGINTIMEOUT=30

echo "$(date --rfc-3339=seconds) - waiting 4 x 10 seconds for SQL Server to come up"
for i in {1..4}
do
    echo "period $i/4 waiting for the SQL Server to come up"
    sleep 10s
done

function runsql() {
    script=$1
    db=$2
    echo "$(date --rfc-3339=seconds) - running $script using sqlcmd on database $db"
    /opt/mssql-tools/bin/sqlcmd -l $LOGINTIMEOUT -S localhost -U sa -P GithubIs2broken -d $db -i $script
    echo "$(date --rfc-3339=seconds) - done with $script"
}

runsql "database.sql" "master"
runsql "test-db.sql" "ssetest"
runsql "test-db-ms_descriptions.sql" "ssetest"

echo "$(date --rfc-3339=seconds) - import-data done"
