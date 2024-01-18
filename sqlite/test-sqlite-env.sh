#!/bin/sh
set -e

echo "=================="
echo "sqlite"
echo "=================="

./setup.sh

# relative path hack with pwd, otherwise not resolved.
export schemaexplorer_driver=sqlite
export schemaexplorer_live=false
export schemaexplorer_sqlite_file="`pwd`/db/test.db"

cd ..
go clean -testcache
go test sse_test.go #-test.v
