#!/bin/sh

export schemaexplorer_driver=mssql
export schemaexplorer_display_name=mssql-test
export schemaexplorer_live=true
export schemaexplorer_listen_on_port=8084
export schemaexplorer_mssql_connection_string="server=localhost;user id=sa;password=GithubIs2broken;database=sse-regression-test"
go run sse.go 2>&1 | sed "s,.*,$(tput setaf 13)mssql-test &$(tput sgr0)," &
wait
