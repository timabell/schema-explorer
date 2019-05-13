#!/bin/sh
# relative path hack with pwd, otherwise not resolved.
# create db first with sqlite/setup.sh
cd ..

go run sse.go \
--driver=sqlite \
--display-name=sqlite-chinook \
--live=true \
--listen-on-port=8802 \
--sqlite-file="$HOME/Documents/projects/sql-data-viewer/Chinook_Sqlite_AutoIncrementPKs.sqlite"
