#!/bin/sh -v

go test ./... -driver sqlite -db ~/Documents/projects/sql-data-viewer/Chinook_Sqlite_AutoIncrementPKs.sqlite -test.v

go test ./... -driver mssql -db "server=sdv-adventureworks.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=AdventureWorksLT" -test.v
