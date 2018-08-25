#!/bin/sh -

echo "=================="
echo "postgres"
echo "=================="

(cd pg/ && ./setup-ssetest.sh)

schemaexplorer_driver=pg schemaexplorer_db="postgres://ssetestusr:ssetestusr@localhost/ssetest" go test sdv_test.go # -test.v
