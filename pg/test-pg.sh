#!/bin/sh
set -e

echo "=================="
echo "postgres"
echo "=================="

./setup-ssetest.sh
cd ..

export schemaexplorer_driver=pg
export schemaexplorer_pg_connection_string="postgres://ssetestusr:ssetestusr@localhost/ssetest?sslmode=disable"
go clean -testcache
go test sse_test.go # -test.v
