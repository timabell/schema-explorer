#!/bin/sh -v
docker pull mysql:latest
docker run --name sse-mysql -e MYSQL_ROOT_PASSWORD=omgroot -p 3306:3306 -d mysql:latest
sleep 25 # wait for container to come up
./setup-once.sh
