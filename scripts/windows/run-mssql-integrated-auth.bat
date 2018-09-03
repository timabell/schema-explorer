rem connect to local sql server using integrated auth (by omitting user id)
rem The server must be listening on tcp/ip, which is *not* the default
sql-data-viewer.exe --driver mssql --mssql-connection-string "server=localhost;database=master"
pause
