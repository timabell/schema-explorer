#!/bin/sh -v

# https://docs.microsoft.com/en-us/sql/linux/quickstart-install-connect-docker?view=sql-server-2017

# map current folder in so that we can run sql files with sqlcmd

docker run -e 'ACCEPT_EULA=Y' -e 'SA_PASSWORD=GithubIs2broken' \
   -p 1433:1433 --name mssql1 \
   -v `pwd`:/src \
   -d mcr.microsoft.com/mssql/server:2017-latest
