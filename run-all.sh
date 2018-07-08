#!/bin/sh -
# relative path hack with pwd, otherwise not resolved.
# create db first with sqlite/setup.sh

# see all colours: https://stackoverflow.com/questions/5947742/how-to-change-the-output-color-of-echo-in-linux#comment32077818_5947788
# for (( i = 0; i < 17; i++ )); do echo "$(tput setaf $i)This is ($i) $(tput sgr0)"; done

echo "http://localhost:8081/ - sqlite test"
echo "http://localhost:8082/ - sqlite chinook"
echo "http://localhost:8083/ - mssql test"
echo "http://localhost:8084/ - adventureworks"
echo "http://localhost:8085/ - wwi (broken)"
echo "http://localhost:8086/ - pg test"
echo "Ctrl-C to tear them all down again."

go run sdv.go -name sqlite-test -driver sqlite -db "`pwd`/sqlite/db/test.db" -port 8081 -live 2>&1 | sed "s,.*,$(tput setaf 10)sqlite-test &$(tput sgr0)," &

./run-sqlite.sh 2>&1 | sed "s,.*,$(tput setaf 11)sqlite &$(tput sgr0)," &

go run sdv.go -name mssql-test -driver mssql -db "server=sdv-regression-test.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=sdv-regression-test" -port 8083 -live 2>&1 | sed "s,.*,$(tput setaf 12)mssql-test &$(tput sgr0)," &

go run sdv.go -name mssql-adventure-works -driver mssql -db "server=sdv-adventureworks.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=AdventureWorksLT" -port 8084 -live 2>&1 | sed "s,.*,$(tput setaf 13)mssql-aw &$(tput sgr0)," &

go run sdv.go -name mssql-wwi -driver mssql -db "server=sdv-wwi.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=WideWorldImporters" -port 8085 -live 2>&1 | sed "s,.*,$(tput setaf 14)mssql-wwi &$(tput sgr0)," &

./run-pg.sh 2>&1 | sed "s,.*,$(tput setaf 15)pg &$(tput sgr0)," &

wait
