#!/bin/bash
set -e

# chomp off switches
if [ "-d" == "$1" ];
then
    shift
    database="$1"
    shift
else
    database="ssetest"
fi

if [ "-f" == "$1" ];
then
    shift
    file="$1"
    shift
else
    query="$1"
    if [ -z "$query" ];
    then
       query="select @@version;"
    fi
fi

if [ -z "$file" ];
then
#    echo "running mssql query on db $database:"
    docker exec mssql1 opt/mssql-tools/bin/sqlcmd -W -U sa -P GithubIs2broken -d "$database" -Q "$query"
else
#    echo "running file $file on db $database"
    docker exec mssql1 opt/mssql-tools/bin/sqlcmd -W -U sa -P GithubIs2broken -d "$database" -i "/src/$file"
fi
