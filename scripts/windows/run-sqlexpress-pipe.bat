.@echo off

rem schema explorer's sql server driver doesn't currently discover the pipe name automatically so need to supply it
rem "SQLEXPRESS" is the instance name which can be changed to suit.
rem You can find the instance name by looking at the details of your sql server [express] service in control panel > services

rem more info:
rem  https://technet.microsoft.com/en-us/library/ms189307(v=sql.105).aspx
rem  https://github.com/denisenkom/go-mssqldb/pull/250

.@echo on
sql-data-viewer.exe --driver mssql --mssql-connection-string "server=np:\\.\pipe\MSSQL$SQLEXPRESS\sql\query;database=master"
