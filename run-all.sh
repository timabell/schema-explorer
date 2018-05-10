#!/bin/sh -
# relative path hack with pwd, otherwise not resolved.
# create db first with sqlite/setup.sh
go run sdv.go -driver sqlite -db "`pwd`/sqlite/db/test.db" -port 8081 -live &
sleep 1
./run-sqlite.sh &
sleep 1
go run sdv.go -driver mssql -db "server=sdv-regression-test.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=sdv-regression-test" -port 8083 -live &
sleep 1
go run sdv.go -driver mssql -db "server=sdv-adventureworks.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=AdventureWorksLT" -port 8084 -live &
sleep 1
go run sdv.go -driver mssql -db "server=sdv-wwi.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=WideWorldImporters" -port 8085 -live &
sleep 1
./run-pg.sh &
wait
