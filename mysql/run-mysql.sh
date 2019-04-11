#!/bin/sh
cd ..

export schemaexplorer_driver=mysql
export schemaexplorer_display_name=mysql-test
export schemaexplorer_live=false
export schemaexplorer_listen_on_port=8831
export schemaexplorer_mysql_database=ssetest
export schemaexplorer_mysql_user=ssetestusr
export schemaexplorer_mysql_password=ssetestusrpass
go run sse.go
