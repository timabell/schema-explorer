#!/bin/sh
./sdv-linux-x64 -driver mssql -db "server=sdv-adventureworks.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=AdventureWorksLT" -port 80 -listenOn 0.0.0.0
