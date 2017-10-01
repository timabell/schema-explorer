.@echo off
rem sdv doesn't currently discover the pipe name automatically so need to supply it
rem "SQLEXPRESS" is the instance name which can be changed to suit.
rem more info:
rem  https://technet.microsoft.com/en-us/library/ms189307(v=sql.105).aspx
rem  https://github.com/denisenkom/go-mssqldb/pull/250

.@echo on
sql-data-viewer.exe -driver mssql -db "server=np:\\.\pipe\MSSQL$SQLEXPRESS\sql\query;database=master"
