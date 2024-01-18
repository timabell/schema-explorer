#!/bin/sh
set -e
# relative path hack with pwd, otherwise not resolved.
# create db first with sqlite/setup.sh
cd ..

export schemaexplorer_driver=sqlite
export schemaexplorer_display_name=sqlite-test
export schemaexplorer_live=true
export schemaexplorer_listen_on_port=8801
export schemaexplorer_sqlite_file="`pwd`/sqlite/db/test.db"
go run sse.go 2>&1 | sed "s,.*,$(tput setaf 9)sqlite-test &$(tput sgr0)," &
wait
