#!/bin/sh -

echo "=================="
echo "mssql"
echo "=================="

go test ./... -driver mssql -db "server=sdv-regression-test.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=sdv-regression-test"  #-test.v
