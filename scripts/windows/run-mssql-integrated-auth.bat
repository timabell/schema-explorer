rem connect to local sql express server using integrated auth (by omitting user id)
sql-data-viewer.exe -driver mssql -db "server=localhost\SQLEXPRESS;database=exampledb" -port 8080
pause
