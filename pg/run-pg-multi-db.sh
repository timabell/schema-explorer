#!/bin/sh
cd ..

export schemaexplorer_driver=pg
export schemaexplorer_live=true
export schemaexplorer_listen_on_port=8812
export schemaexplorer_pg_host=/var/run/postgresql/
go run sse.go 2>&1 | sed "s,.*,$(tput setaf 14)pg-test &$(tput sgr0)," &
wait
