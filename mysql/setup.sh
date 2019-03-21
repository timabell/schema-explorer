#!/bin/sh
echo "drop database if exists ssetest;" | mysql -u ssetestusr -pssetestusrpass
echo "create database ssetest;" | mysql -u ssetestusr -pssetestusrpass
mysql -u ssetestusr -pssetestusrpass ssetest < test-db.sql

