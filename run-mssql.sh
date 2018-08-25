#!/bin/sh -

# see all colours: https://stackoverflow.com/questions/5947742/how-to-change-the-output-color-of-echo-in-linux#comment32077818_5947788
# for (( i = 0; i < 17; i++ )); do echo "$(tput setaf $i)This is ($i) $(tput sgr0)"; done

go run sdv.go --display-name mssql-test --driver mssql --db "server=sdv-regression-test.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=sdv-regression-test" --listen-on-port 8084 --live 2>&1 | sed "s,.*,$(tput setaf 12)mssql-test &$(tput sgr0)," &
go run sdv.go --display-name mssql-test --driver mssql --db "server=sdv-adventureworks.database.windows.net;user id=sdvRO;password=Startups 4 the rest of us;database=AdventureWorksLT" --listen-on-port 8084 --live 2>&1 | sed "s,.*,$(tput setaf 13)mssql-aw &$(tput sgr0)," &

wait
