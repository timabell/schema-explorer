# Makefile for schema-explorer
# Needs python3 IF DOCKER_HOST environment variable is set
# https://github.com/gotestyourself/gotestsum/releases


.DEFAULT_GOAL := usage

# GCC on windows for using with CGO
# TDM-GCC-64 - http://tdm-gcc.tdragon.net/download
# MSYS2 - http://www.msys2.org/

# try to be os agnostic
ifeq ($(OS),Windows_NT)
	# aliases
	RM = del /q
	RMDIR = rmdir /s /q
	DEVNUL := NUL
	WHICH := where
	CAT = type
	MV = move
	CPR = xcopy /s /e /y /i
	MKDIR = mkdir
	CP = copy
	ZIP = @echo Windows should learn how to zip
	NEWLINE = echo.
	# functions
	FixPath = $(subst /,\,$1)
	Sleep = ping 192.0.2.0 -n 1 -w $1000

	# variables
	CGO_windows ?= 1
	CGO_linux ?= 0
	CGO_darwin ?= 0
else
	# aliases
	RM = rm -f
	RMDIR = rm -Rf
	CAT = cat
	DEVNUL = /dev/null
	WHICH = which
	MV = mv
	MKDIR = mkdir -p
	CPR = cp -r
	CP = cp
	ZIP = zip -rq 
	NEWLINE = printf "\n"
	# functions
	FixPath = $1
	Sleep = sleep $1

	# os specific
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		CGO_windows ?= 1
		CGO_linux ?= 1
		CGO_darwin ?= 0
	endif
	ifeq ($(UNAME_S),Darwin)
        CGO_windows ?= 1
		CGO_linux ?= 1
		CGO_darwin ?= 1
    endif
endif

CC_windows ?= x86_64-w64-mingw32-gcc
CC_linux := $(CC)
CC_darwin := $(CC)

EXT_windows = ".exe"

TESTDIR ?= ./build/tests
TESTFILE=$(call FixPath,${TESTDIR}/gotestsum-$@-report.xml)
GOTEST=gotestsum --junitfile ${TESTFILE} --

ifeq ($(shell ${WHICH} gotestsum 2>${DEVNUL}),)
$(info You don't have 'gotestsum' on your PATH. Will fall back to using 'go test')
$(info Go to https://github.com/gotestyourself/gotestsum for instructions)
GOTEST:=go test
endif

GIT_VERSION = $(shell git rev-parse HEAD)

DC = docker-compose -p sse -f testdata/docker-compose.yml
HOSTNAME:=127.0.0.1

# Find out DOCKER_HOST address to use as hostname for testing database servers
# Used if your docker_host is not same as localhost.
# This will happen for example if you are using `docker-machine`
ifdef DOCKER_HOST
	SCRIPT:=python3 -c "from urllib.parse import urlparse; data = urlparse('${DOCKER_HOST}'); print(data.netloc.replace(':' + str(data.port), ''))"
	HOSTNAME:=$(shell $(SCRIPT))
endif

TEST_USR=ssetestusr
TEST_PWD=ssetestusrpass
TEST_DB=ssetest

SQLITE_DB:=$(call FixPath,./db/test.db)
SQLITE_SCRIPT:=$(call FixPath,sqlite/test-db.sql)
MSSQL_CNN:=sqlserver://sa:GithubIs2broken@${HOSTNAME}?database=${TEST_DB}
PG_CNN:=postgres://${TEST_USR}:${TEST_PWD}@${HOSTNAME}/${TEST_DB}?sslmode=disable
PG_CNN_MULTI:=postgres://${TEST_USR}:${TEST_PWD}@${HOSTNAME}/?sslmode=disable
MYSQL_CNN:="${TEST_USR}:${TEST_PWD}@tcp(${HOSTNAME}:3306)/${TEST_DB}"


PLATFORMS = windows darwin linux
BUILDS = $(addprefix build-, $(PLATFORMS))
FILES = $(addprefix files-, $(PLATFORMS))

DRIVERS = sqlite mssql pg mysql

TESTS = $(addprefix test-, $(DRIVERS))
.SUFFIXES:
.PHONY: test $(TESTS)

ifeq ($(shell ${WHICH} gcc --version 2>${DEVNUL}),)
$(info You don't have 'gcc' on your PATH. Please install first.)
$(info On windows you could use http://tdm-gcc.tdragon.net/download)
$(info On Debian/Ubuntu try sudo apt-get install build-essential)
$(error Can not continue without a working installation of gcc)
endif

$(TESTDIR):
	$(MKDIR) $(call FixPath,$@)

usage:
	@echo ---- usage ----
	@echo OS=$(OS) UNAME_S=$(UNAME_S) GOTEST=$(GOTEST)
	@$(NEWLINE)
	@echo Check the contents of TESTING.md for usage
	@echo Or try "make clean", "make test", or "make build"
	@$(NEWLINE)

build: $(BUILDS)

package: build files
	$(CP) README.md $(call FixPath,build/sse/)
	cd build; $(ZIP) sse.zip sse

files: $(FILES)

test: test-units $(TESTS)

test-docker: docker sleep30 test

clean:
	go clean
	-$(RMDIR) $(call FixPath,./build)
	-$(RMDIR) $(call FixPath,./db)

clean-all: clean docker-clean

## Below are "extra" hackery stuff that everything above depends on

build-%: export GOOS=$*
build-%: export CGO_ENABLED=$(CGO_$*)
build-%: export CC=$(CC_$*)
build-%: 
	@$(NEWLINE)
	@echo Will build for '$*' platform on $(OS) using CGO_ENABLED=$(CGO_ENABLED) and CC=$(CC)
	go build -ldflags "-X github.com/timabell/schema-explorer/about.gitVersion=$(GIT_VERSION)" -o build/sse/$*/schemaexplorer${EXT_$*} sse.go

files-%:
	-$(MKDIR) $(call FixPath,build/sse/$*)
	$(CPR) templates $(call FixPath,build/sse/$*/templates/)
	$(CPR) static $(call FixPath,build/sse/$*/static/) 
	$(CPR) config $(call FixPath,build/sse/$*/config/)


test-units: $(TESTDIR)
	$(GOTEST) ./... -tags=unit
	

#### Sqlite

sqlitedb:
	-$(RMDIR) $(call FixPath,./db)
	$(MKDIR) db
	$(CAT) $(SQLITE_SCRIPT) | sqlite3 $(SQLITE_DB)


test-sqlite: test-sqlite-flags test-sqlite-env


test-sqlite-flags: sqlitedb $(TESTDIR)
	@echo $@
	-$(RMDIR) $(call FixPath,./db)
	$(MKDIR) db
	$(CAT) $(SQLITE_SCRIPT) | sqlite3 $(SQLITE_DB)
	go clean -testcache
	$(GOTEST) . --driver sqlite --sqlite-file $(SQLITE_DB) \
		--display-name=testing-flags \
		--live=true \
		--listen-on-port=9999 \

test-sqlite-env: export schemaexplorer_driver=sqlite
test-sqlite-env: export schemaexplorer_live=false
test-sqlite-env: export schemaexplorer_sqlite_file=${SQLITE_DB}
test-sqlite-env: sqlitedb $(TESTDIR)
	@echo $@
	-$(RMDIR) $(call FixPath,./db)
	$(MKDIR) db
	$(CAT) $(SQLITE_SCRIPT) | sqlite3 $(SQLITE_DB)
	go clean -testcache
	$(GOTEST) .

#### Microsoft SQL

test-mssql: $(TESTDIR)
	go clean -testcache
	$(GOTEST) . -driver mssql -mssql-connection-string ${MSSQL_CNN}

#### Postgres

test-pg: test-pg-flags test-pg-multi

test-pg-flags: $(TESTDIR)
	go clean -testcache
	$(GOTEST) . -driver pg -pg-connection-string ${PG_CNN}

test-pg-multi: $(TESTDIR)
	go clean -testcache
	$(GOTEST) . -driver pg -pg-connection-string ${PG_CNN_MULTI}

#### Mysql

test-mysql: test-mysql-flags test-mysql-flags-cnn test-mysql-env test-mysql-env-cnn

test-mysql-flags: $(TESTDIR)
	$(GOTEST) . --live=true -driver mysql \
		--mysql-host ${HOSTNAME} \
		-mysql-user ${TEST_USR} \
		-mysql-password ${TEST_PWD}

test-mysql-flags-cnn: $(TESTDIR)
	go clean -testcache
	$(GOTEST) . --live=true -driver mysql --mysql-connection-string ${MYSQL_CNN}
	

test-mysql-env: export schemaexplorer_driver = mysql
test-mysql-env: export schemaexplorer_live = false
test-mysql-env: export schemaexplorer_mysql_host = ${HOSTNAME}
test-mysql-env: export schemaexplorer_mysql_port = 3306
test-mysql-env: export schemaexplorer_mysql_user = ${TEST_USR}
test-mysql-env: export schemaexplorer_mysql_password = ${TEST_PWD}
test-mysql-env: $(TESTDIR) 
	go clean -testcache
	 $(GOTEST) . 


test-mysql-env-cnn: export schemaexplorer_mysql_host =
test-mysql-env-cnn: export schemaexplorer_mysql_port = 
test-mysql-env-cnn: export schemaexplorer_mysql_user = 
test-mysql-env-cnn: export schemaexplorer_mysql_password = 
test-mysql-env-cnn: export schemaexplorer_driver = mysql
test-mysql-env-cnn: export schemaexplorer_live = false
test-mysql-env-cnn: export schemaexplorer_mysql_connection_string=${MYSQL_CNN}
test-mysql-env-cnn: $(TESTDIR) 
	@echo schemaexplorer_mysql_connection_string=${schemaexplorer_mysql_connection_string}
	@echo schemaexplorer_mysql_host=${schemaexplorer_mysql_host}
	go clean -testcache
	 $(GOTEST) . 


run-mysql:
	go run sse.go -driver mysql -mysql-connection-string ${MYSQL_CNN}


#### sleep
sleep%:
	@echo Waiting for $* seconds
	-$(call Sleep,$*)


#### docker

docker: docker-clean docker-build docker-up
	@echo "waiting for containters to hopefully wake up"
	@echo "run `make docker-logs` to verify everything is up and running"

docker-up:
	$(DC) up -d ${CONTAINER}

docker-kill:
	$(DC) kill ${CONTAINER}

docker-logs:
	@echo Stop following logs with ctrl+c
	$(DC) logs -f ${CONTAINER}

docker-build:
	$(DC) build ${CONTAINER}

docker-clean: docker-kill
	$(DC) rm --force -v


	