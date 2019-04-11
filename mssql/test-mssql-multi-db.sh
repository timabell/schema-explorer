#!/bin/bash

echo "=================="
echo "mssql"
echo "=================="

echo "running mssql/test-setup.sh..."
./test-setup.sh

cd ..
export schemaexplorer_driver=mssql
export schemaexplorer_mssql_host=localhost
export schemaexplorer_mssql_user=sa
export schemaexplorer_mssql_password=GithubIs2broken
go clean -testcache
go test sse_test.go # -test.v
