Testing Schema Explorer
=======================
The following instructions assumes that you have `sqlite3`, `docker` 
and `docker-compose` installed and on your path. 
If you do not have access to docker you will not be able to run 
integration tests unless you provide for an environment with postgres, mysql
and mssql setup and configured in a propper way. You will be able to run some
tests even though.

If you are using a remote docker host you will
need `python3` as well to parse the uri in the `DOCKER_HOST` environment 
variable.

The database scripts and other initialization for each container (sqlite 
is run locally) can be found under `testdata/container/<driver>`.
There is a docker-compose file in `testdata/docker-compose.yml` that will be
used with these containers.

# How to build
This is a hidden gem but the makefile does support building the binaries
as well as testing.
Try out `make build` or `make package`.


# How to test
```shell
# Clean up any old and prepare for testing by setting up 
# docker for mssql, mysql and postgres
make docker

# Run all tests. 
# This nees a working environment with database servers to run successfully
make test

# the one above is basicall the same as
make test-units test-sqlite test-pg test-mysql test-mssql

# Test sqlite separately. Does not need a server

make test-sqlite
# These separate tests need database server
make test-pg
make test-mssql
make test-mysql

# Create containers and run tests in one sweep
make test-docker

# test-docker is basically the same as
make docker test

# cleanup
make clean

# stop and remove all containers
make clean-docker

# to remove everything built.
make clean-all


# Run all at once
make containers test clean clean-containers
```

# Build and manage images
If you change anything in the `testdata/container/<driver>/` files you will have
to rebuild that container.
There is a somewhat brutal way of doing it by calling `make docker-build`
that simply rebuilds all images defined `testdata/docker-compose.yml` files
regardles if they have been changed or not.
This process takes some time the first time you do it since all base images
needs to be downloaded if missing.
On all following occations this will happen a lot quicker since it's only
a few small files that are added to initialize them correctly.

This rebuild will happen on...
* `make docker`
* `make docker-build`
* `make test-docker`

This can leave behind a lot of unused images which if you are careful and
know what you are doing kan be removed with `docker image prune`.
