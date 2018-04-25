#!/bin/sh -

echo "=================="
echo "postgres"
echo "=================="

(cd pg/ && ./setup.sh)

go test ./... -driver pg -db "postgres://postgres:postgres@localhost/sse-test" # -test.v
