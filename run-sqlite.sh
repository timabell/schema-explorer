#!/bin/sh -
# relative path hack with pwd, otherwise not resolved.
# create db first with sqlite/setup.sh

#go run sdv.go -name sqlite-chinook -driver sqlite -db "$HOME/Documents/projects/sql-data-viewer/Chinook_Sqlite_AutoIncrementPKs.sqlite" -port 8082 -live

schemaexplorer_driver=sqlite schemaexplorer_name=sqlite-chinook schemaexplorer_live=true schemaexplorer_listenonport=8082 schemaexplorer_file="$HOME/Documents/projects/sql-data-viewer/Chinook_Sqlite_AutoIncrementPKs.sqlite" go run sdv.go
