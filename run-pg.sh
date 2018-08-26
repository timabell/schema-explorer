#!/bin/sh -

export schemaexplorer_driver=pg
export schemaexplorer_display_name=pg-test
export schemaexplorer_live=true
export schemaexplorer_listen_on_port=8085
export schemaexplorer_pg_db="postgres://ssetestusr:ssetestusr@localhost/ssetest" 
go run sdv.go 2>&1 | sed "s,.*,$(tput setaf 15)pg-test &$(tput sgr0)," &
wait
