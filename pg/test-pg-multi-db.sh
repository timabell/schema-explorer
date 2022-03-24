#!/bin/sh

echo "=================="
echo "postgres multi-db"
echo "=================="

./setup-ssetest.sh
cd ..

export schemaexplorer_driver=pg
# connect with socket and no pre-specified database
export schemaexplorer_pg_host=localhost
export schemaexplorer_pg_user=postgres
export schemaexplorer_pg_password=postgres
export schemaexplorer_pg_ssl_mode=disable
go clean -testcache
go test sse_test.go # -test.v
