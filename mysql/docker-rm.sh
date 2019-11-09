#!/bin/sh -v
# danger, deletes container+data
./docker-stop.sh
docker rm sse-mysql
