#!/bin/sh -

echo "=================="
echo "sqlite"
echo "=================="


# relative path hack with pwd, otherwise not resolved.
# create db first with sqlite/setup.sh
go test ./... -driver sqlite -db "`pwd`/sqlite/db/test.db" -test.v

echo "=================="
echo "mssql"
echo "=================="


go test ./... -driver mssql -db "server=sdv-regression-test.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=sdv-regression-test" -test.v
