#!/bin/sh
# relative path hack with pwd, otherwise not resolved.
# create db first with sqlite/setup.sh

# see all colours: https://stackoverflow.com/questions/5947742/how-to-change-the-output-color-of-echo-in-linux#comment32077818_5947788
# for (( i = 0; i < 17; i++ )); do echo "$(tput setaf $i)This is ($i) $(tput sgr0)"; done

echo -n "$(tput setaf 2)"
echo "http://localhost:8801/ - sqlite"
echo "http://localhost:8802/ - sqlite chinook"

echo -n "$(tput setaf 14)"
echo "http://localhost:8811/ - pg"
echo "http://localhost:8812/ - pg multi-db"

echo -n "$(tput setaf 13)"
echo "http://localhost:8821/ - mssql"
echo "http://localhost:8822/ - mssql multidb"

echo -n "$(tput setaf 9)"
echo "http://localhost:8831/ - mysql"

echo -n "$(tput sgr0)"
echo "Ctrl-C to tear them all down again."

cd sqlite
./run-sqlite.sh &
sleep 0.5
./run-sqlite-test.sh &
sleep 0.5

cd ../pg
./run-pg.sh &
./run-pg-multi-db.sh &
sleep 0.5

cd ../mssql
./run-mssql.sh &
./run-mssql-multi-db.sh &
sleep 0.5

cd ../mysql
./run-mysql.sh &

wait
