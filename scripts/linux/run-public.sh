#!/bin/sh
./sdv-linux-x64 --driver mssql --mssql-connection-string "server=sdv-adventureworks.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=AdventureWorksLT" --listen-on-port 80 --listen-on-address 0.0.0.0
