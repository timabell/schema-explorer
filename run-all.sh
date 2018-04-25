#!/bin/sh -
# relative path hack with pwd, otherwise not resolved.
# create db first with sqlite/setup.sh
go run sdv.go -driver sqlite -db "`pwd`/sqlite/db/test.db" -port 8081 -live &
go run sdv.go -driver sqlite -db "$HOME/Documents/projects/sql-data-viewer/Chinook_Sqlite_AutoIncrementPKs.sqlite" -port 8082 -live &
go run sdv.go -driver mssql -db "server=sdv-regression-test.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=sdv-regression-test" -port 8083 -live &
go run sdv.go -driver mssql -db "server=sdv-adventureworks.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=AdventureWorksLT" -port 8084 -live &
go run sdv.go -driver mssql -db "server=sdv-wwi.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=WideWorldImporters" -port 8085 -live &
go run sdv.go -driver pg -db "postgres://ssetest:ssetest@localhost/sse-test" -port 8086 -live &
wait
