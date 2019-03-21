#!/bin/sh

echo "=================="
echo "mysql"
echo "=================="

(cd mysql/ && ./setup.sh)

# relative path hack with pwd, otherwise not resolved.
export schemaexplorer_driver=mysql
export schemaexplorer_live=false
export schemaexplorer_mysql_database=ssetest
go clean -testcache
go test sse_test.go #-test.v
