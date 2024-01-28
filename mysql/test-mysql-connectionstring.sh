#!/bin/sh
set -e

echo "======================"
echo "mysql connectionstring"
echo "======================"

./setup.sh

cd ..
export schemaexplorer_driver=mysql
export schemaexplorer_live=false
export schemaexplorer_mysql_connection_string="ssetestusr:ssetestusrpass@tcp(localhost:3306)/ssetest"
go clean -testcache
go test sse_test.go #-test.v
