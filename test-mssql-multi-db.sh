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
export schemaexplorer_mssql_host=localhost
export schemaexplorer_mssql_user=sa
export schemaexplorer_mssql_password=GithubIs2broken
go clean -testcache
go test sse_test.go # -test.v
