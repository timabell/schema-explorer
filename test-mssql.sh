#!/bin/sh

echo "=================="
echo "mssql"
echo "=================="

export schemaexplorer_driver=mssql
export schemaexplorer_mssql_connection_string="server=localhost;user id=sa;password=GithubIs2broken;database=sdv-regression-test"
go test sdv_test.go # -test.v
