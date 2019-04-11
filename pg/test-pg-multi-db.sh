#!/bin/sh

echo "=================="
echo "postgres multi-db"
echo "=================="

./setup-ssetest.sh
cd ..

export schemaexplorer_driver=pg
# connect with socket and no pre-specified database
export schemaexplorer_pg_host=/var/run/postgresql/
go clean -testcache
go test sse_test.go # -test.v
