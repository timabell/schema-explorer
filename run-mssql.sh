#!/bin/sh -

export schemaexplorer_driver=mssql
export schemaexplorer_display_name=mssql-adventureworks
export schemaexplorer_live=true
export schemaexplorer_listen_on_port=8083
export schemaexplorer_mssql_db="server=sdv-adventureworks.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=AdventureWorksLT" 
go run sdv.go 2>&1 | sed "s,.*,$(tput setaf 13)mssql-aw &$(tput sgr0)," &
wait
