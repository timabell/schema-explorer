#!/bin/bash
set -e

echo "=================="
echo "mssql"
echo "=================="

echo "running mssql/test-setup.sh..."
./test-setup.sh

cd ..
export schemaexplorer_driver=mssql
export schemaexplorer_mssql_connection_string="server=localhost;user id=sa;password=GithubIs2broken;database=ssetest"
go clean -testcache
go test sse_test.go # -test.v
