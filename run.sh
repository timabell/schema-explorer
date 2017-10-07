#!/bin/sh -
# relative path hack with pwd, otherwise not resolved.
# create db first with sqlite/setup.sh
go run sdv.go -driver sqlite -db "`pwd`/sqlite/db/test.db" -port 8088 &
go run sdv.go -driver mssql -db "server=sdv-regression-test.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=sdv-regression-test" -port 8089 &
wait
