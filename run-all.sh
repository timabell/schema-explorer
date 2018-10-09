#!/bin/sh
# relative path hack with pwd, otherwise not resolved.
# create db first with sqlite/setup.sh

# see all colours: https://stackoverflow.com/questions/5947742/how-to-change-the-output-color-of-echo-in-linux#comment32077818_5947788
# for (( i = 0; i < 17; i++ )); do echo "$(tput setaf $i)This is ($i) $(tput sgr0)"; done

echo -n "$(tput setaf 2)"
echo "http://localhost:8081/ - sqlite chinook"
echo -n "$(tput setaf 9)"
echo "http://localhost:8082/ - sqlite test"
echo -n "$(tput setaf 14)"
echo "http://localhost:8085/ - pg test"
# echo -n "$(tput setaf 12)"
# echo "http://localhost:8083/ - mssql adventureworks"
# echo -n "$(tput setaf 13)"
# echo "http://localhost:8084/ - mssql test"
# echo "http://localhost:8086/ - wwi (broken)"
echo -n "$(tput sgr0)"
echo "Ctrl-C to tear them all down again."

./run-sqlite.sh &
sleep 0.5
./run-sqlite-test.sh &
sleep 0.5
./run-pg.sh &
# sleep 0.5
# ./run-mssql.sh &
# sleep 0.5
# ./run-mssql-test.sh &
wait
