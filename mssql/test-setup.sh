#!/bin/bash
set -e
testdb="ssetest"
./runsql.sh -d master "drop database if exists [$testdb];"
./runsql.sh -d master "create database [$testdb];"
./runsql.sh -d "$testdb" -f test-db.sql
./runsql.sh -d "$testdb" -f test-db-ms_descriptions.sql
