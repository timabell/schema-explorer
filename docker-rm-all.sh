#!/bin/sh
# https://stackoverflow.com/a/30311197/10245
docker rm -f $(docker ps -a -q)
docker rmi -f  $(docker images -q)
