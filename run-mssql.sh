#!/bin/sh

export schemaexplorer_driver=mssql
export schemaexplorer_display_name=mssql-adventureworks
export schemaexplorer_live=true
export schemaexplorer_listen_on_port=8083
export schemaexplorer_mssql_connection_string="server=sse-adventureworks.database.windows.net;user id=sseRO;password=Startups 4 the rest of us;database=AdventureWorksLT"
go run sse.go 2>&1 | sed "s,.*,$(tput setaf 12)mssql-aw &$(tput sgr0)," &
wait
