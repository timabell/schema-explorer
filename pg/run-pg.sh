#!/bin/sh
set -e
cd ..

export schemaexplorer_driver=pg
export schemaexplorer_display_name=pg-test
export schemaexplorer_live=true
export schemaexplorer_listen_on_port=8811
export schemaexplorer_pg_connection_string="postgres://ssetestusr:ssetestusr@localhost/ssetest?sslmode=disable"
go run sse.go 2>&1 | sed "s,.*,$(tput setaf 14)pg-test &$(tput sgr0)," &
wait
