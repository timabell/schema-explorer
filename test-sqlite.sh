#!/bin/sh

echo "=================="
echo "sqlite"
echo "=================="

(cd sqlite/ && ./setup.sh)

# relative path hack with pwd, otherwise not resolved.
export schemaexplorer_driver=sqlite
export schemaexplorer_sqlite_file="`pwd`/sqlite/db/test.db"
go test sdv_test.go #-test.v
