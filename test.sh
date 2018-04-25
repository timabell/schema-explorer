#!/bin/sh -

echo "=================="
echo "sqlite"
echo "=================="

(cd sqlite/ && ./setup.sh)

# relative path hack with pwd, otherwise not resolved.
go test ./... -driver sqlite -db "`pwd`/sqlite/db/test.db" # -test.v

echo "=================="
echo "mssql"
echo "=================="

go test ./... -driver mssql -db "server=sdv-regression-test.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=sdv-regression-test"  #-test.v

echo "=================="
echo "postgres"
echo "=================="

(cd pg/ && ./setup.sh)

go test -driver pg -db "postgres://postgres:postgres@localhost/sse-test" # -test.v
