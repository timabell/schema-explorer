Testing Schema Explorer
=======================
The following instructions assumes that you have `sqlite3`, `docker` 
and `docker-compose` installed and on your path. 

If you are using a remote docker host you will
need `python3` as well to parse the uri in the `DOCKER_HOST` environment 
variable.

The database scripts and other initialization for each container (sqlite 
is run locally) can be found under `test/container/<driver>`.
There is a docker-compose file in `test/test-compose.yml` that will be
used with these containers.

# How to test
```shell
# Prepare for testing by setting up containters for mssql, mysql and postgres
make containers

# Run all tests
make test

# Run separate tests
make testsqlite
make testpg
make testmssql
make testmysql

# Create containers and run tests in one sweep
make test-containers

# cleanup
make clean

# stop and remove all containers
make clean-containers
```

# Build and manage images
If you change anything in the `test/container/<driver>/` files you will have
to rebuild that container.
There is a somewhat brutal way of doing it by calling `make container-build`
that simply rebuilds all images defined `test/test-compose.yml` files
regardles if they have been changed or not.
This process takes some time the first time you do it since all base images
needs to be downloaded if missing.
On all following occations this will happen a lot quicker since it's only
a few small files that are added to initialize them correctly.

This rebuild will happen on...
* `make build`
* `make containers`
* `make test-containers`

This can leave behind a lot of unused images which if you are careful and
know what you are doing kan be removed with `docker image prune`.
