#!/bin/sh

echo "=================="
echo "mysql"
echo "=================="

export schemaexplorer_driver=mysql
export schemaexplorer_live=false
export schemaexplorer_mysql_database=ssetest
export schemaexplorer_mysql_user=ssetestusr
export schemaexplorer_mysql_password=ssetestusrpass
go run sse.go
