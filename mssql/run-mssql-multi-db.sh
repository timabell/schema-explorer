#!/bin/sh
cd ..

export schemaexplorer_driver=mssql
export schemaexplorer_live=true
export schemaexplorer_listen_on_port=8822
export schemaexplorer_mssql_host=localhost
export schemaexplorer_mssql_user=sa
export schemaexplorer_mssql_password=GithubIs2broken
go run sse.go 2>&1 | sed "s,.*,$(tput setaf 13)mssql-test &$(tput sgr0)," &
wait
