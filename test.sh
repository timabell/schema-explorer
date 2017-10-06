#!/bin/sh -v

# relative path hack with pwd, otherwise not resolved.
# create db first with sqlite/setup.sh
go test ./... -driver sqlite -db "`pwd`/sqlite/db/test.db" -test.v

go test ./... -driver mssql -db "server=sdv-adventureworks.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=AdventureWorksLT" -test.v
