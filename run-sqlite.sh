#!/bin/sh -
# relative path hack with pwd, otherwise not resolved.
# create db first with sqlite/setup.sh

export schemaexplorer_driver=sqlite
export schemaexplorer_display_name=sqlite-chinook
export schemaexplorer_live=true
export schemaexplorer_listen_on_port=8081
export schemaexplorer_file="$HOME/Documents/projects/sql-data-viewer/Chinook_Sqlite_AutoIncrementPKs.sqlite"
go run sdv.go 2>&1 | sed "s,.*,$(tput setaf 11)sqlite &$(tput sgr0)," &
wait
