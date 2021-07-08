#!/bin/bash
set -e

# https://dev.to/nabbisen/microsoft-sql-server-client-on-debian-1p3o
# https://docs.microsoft.com/en-us/windows-server/administration/linux-package-repository-for-microsoft-software
curl https://packages.microsoft.com/keys/microsoft.asc | sudo apt-key add -
curl https://packages.microsoft.com/config/debian/10/prod.list | sudo tee /etc/apt/sources.list.d/msprod.list

echo "Installing mssql-tools ACCEPT_EULA=$ACCEPT_EULA"
sudo apt-get update
sudo ACCEPT_EULA=y DEBIAN_FRONTEND=noninteractive apt-get install -y mssql-tools locales

echo "en_US.UTF-8 UTF-8" | sudo tee -a /etc/locale.gen
sudo locale-gen

LOGINTIMEOUT=30
SQLCMD=/opt/mssql-tools/bin/sqlcmd
USER=sa
PASSWD=${SA_PASSWORD:-GithubIs2broken}


echo Waiting for mssql...
dockerize -wait tcp://localhost:1433 -timeout 1m

# Do the actual work

function runsql() {
    script=$1
    db=$2
    echo "$(date --rfc-3339=seconds) - running $script using sqlcmd on database $db"
    $SQLCMD -l $LOGINTIMEOUT -S localhost -U sa -P GithubIs2broken -d $db -i $script
    echo "$(date --rfc-3339=seconds) - done with $script"
}

runsql "testdata/container/mssql/database.sql" "master"
runsql "testdata/container/mssql/test-db.sql" "ssetest"
runsql "testdata/container/mssql/test-db-ms_descriptions.sql" "ssetest"



echo Done!