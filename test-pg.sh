#!/bin/sh

echo "=================="
echo "postgres"
echo "=================="

(cd pg/ && ./setup-ssetest.sh)

export schemaexplorer_driver=pg
export schemaexplorer_pg_connection_string="postgres://ssetestusr:ssetestusr@localhost/ssetest"
go test sdv_test.go # -test.v
