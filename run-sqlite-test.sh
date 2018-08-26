#!/bin/sh -
# relative path hack with pwd, otherwise not resolved.
# create db first with sqlite/setup.sh

export schemaexplorer_driver=sqlite
export schemaexplorer_display_name=sqlite-test
export schemaexplorer_live=true
export schemaexplorer_listen_on_port=8082
export schemaexplorer_file="`pwd`/sqlite/db/test.db"
go run sdv.go 2>&1 | sed "s,.*,$(tput setaf 10)sqlite-test &$(tput sgr0)," &
wait
