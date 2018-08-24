#!/bin/sh -

echo "=================="
echo "sqlite"
echo "=================="

(cd sqlite/ && ./setup.sh)

# relative path hack with pwd, otherwise not resolved.
schemaexplorer_driver=sqlite schemaexplorer_file="`pwd`/sqlite/db/test.db" go test sdv_test.go # -test.v
