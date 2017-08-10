#!/bin/sh
# https://stackoverflow.com/a/30311197/10245
docker rm $(docker ps -a -q)
docker rmi $(docker images -q)
