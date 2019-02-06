#!/bin/bash

echo "=================="
echo "mssql"
echo "=================="

echo "running mssql/test-setup.sh..."
pushd . > /dev/null
cd mssql
./test-setup.sh
popd > /dev/null

export schemaexplorer_driver=mssql
export schemaexplorer_mssql_connection_string="server=localhost;user id=sa;password=GithubIs2broken;database=sse-regression-test"
go clean -testcache
go test sse_test.go # -test.v
