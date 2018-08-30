#!/bin/sh

echo "=================="
echo "mssql"
echo "=================="

export schemaexplorer_driver=mssql
export schemaexplorer_mssql_connection_string="server=sdv-regression-test.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=sdv-regression-test"
go test sdv_test.go # -test.v
