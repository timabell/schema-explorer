#!/bin/sh -

echo "=================="
echo "postgres"
echo "=================="

(cd pg/ && ./setup-ssetest.sh)

export schemaexplorer_driver=pg
export schemaexplorer_pg_db="postgres://ssetestusr:ssetestusr@localhost/ssetest"
go test sdv_test.go # -test.v
