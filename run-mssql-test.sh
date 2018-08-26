#!/bin/sh -

export schemaexplorer_driver=mssql
export schemaexplorer_display_name=mssql-test
export schemaexplorer_live=true
export schemaexplorer_listen_on_port=8084
export schemaexplorer_mssql_db="server=sdv-regression-test.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=sdv-regression-test" 
go run sdv.go 2>&1 | sed "s,.*,$(tput setaf 12)mssql-test &$(tput sgr0)," &
wait
