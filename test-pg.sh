#!/bin/sh -

echo "=================="
echo "postgres"
echo "=================="

(cd pg/ && ./setup-ssetest.sh)

go test ./... -driver pg -db "postgres://ssetestusr:ssetestusr@localhost/ssetest" # -test.v
