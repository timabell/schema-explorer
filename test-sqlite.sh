#!/bin/sh -

echo "=================="
echo "sqlite"
echo "=================="

(cd sqlite/ && ./setup.sh)

# relative path hack with pwd, otherwise not resolved.
go test ./... -driver sqlite -db "`pwd`/sqlite/db/test.db" # -test.v
