# Makefile for schema-explorer
# Needs python3 IF DOCKER_HOST environment variable is set

# make some commands work on windows as well as *nix
ifdef SystemRoot
	RM = del /q
	RMDIR = rmdir /s /q
	CAT = type
	FixPath = $(subst /,\,$1)
	# based upon https://stackoverflow.com/questions/1672338/how-to-sleep-for-five-seconds-in-a-batch-file-cmd
	Sleep = ping 192.0.2.0 -n 1 -w $1000
else
	RM = rm -f
	RMDIR = rm -Rf
	CAT = cat
	FixPath = $1
	Sleep = sleep $1
endif

HOSTNAME:=localhost

# Find out DOCKER_HOST address to use as hostname for testing database servers
# Used if your docker_host is not same as localhost.
# This will happen for example if you are using `docker-machine`
ifdef DOCKER_HOST
	SCRIPT:=python3 -c "from urllib.parse import urlparse; data = urlparse('${DOCKER_HOST}'); print(data.netloc.replace(':' + str(data.port), ''))"
	HOSTNAME:=$(shell $(SCRIPT))
endif

COMPOSEFILE:=test/test-compose.yml

SQLITE_DB:=$(call FixPath,db/test.db)
SQLITE_SCRIPT:=$(call FixPath,sqlite/test-db.sql)
MSSQL_CNN:="sqlserver://sa:GithubIs2broken@${HOSTNAME}?database=ssetest"
PG_CNN:="postgres://ssetestusr:ssetestusrpass@${HOSTNAME}/ssetest?sslmode=disable"
MYSQL_CNN:="ssetestusr:ssetestusrpass@tcp(${HOSTNAME}:3306)/ssetest"

DRIVERS = sqlite mssql pg mysql

TESTS = $(addprefix test, $(DRIVERS))

.PHONY: test $(TESTS)

test: $(TESTS)

test-containers: containers test

testsqlite:
	-$(RMDIR) $(call FixPath,./db)
	mkdir db
	$(CAT) $(SQLITE_SCRIPT) | sqlite3 $(SQLITE_DB)
	go clean -testcache
	go test sse_test.go -driver sqlite -sqlite-file $(SQLITE_DB)

testmssql:
	go clean -testcache
	go test sse_test.go -driver mssql -mssql-connection-string ${MSSQL_CNN}

testpg:
	go clean -testcache
	go test sse_test.go -driver pg -pg-connection-string ${PG_CNN}

testmysql:
	go clean -testcache
	go test sse_test.go -driver mysql -mysql-connection-string ${MYSQL_CNN}


runmysql:
	go run sse.go -driver mysql -mysql-connection-string ${MYSQL_CNN}	


containers: clean-containers containers-build containers-up
	@echo "waiting for containters to hopefully wake up"
	-$(call Sleep,60)
	@echo "run `make containers-logs` to verify everything is up and running"


containers-up:
	docker-compose -f $(COMPOSEFILE) up -d

containers-kill:
	docker-compose -f $(COMPOSEFILE) kill

containers-logs:
	docker-compose -f $(COMPOSEFILE) logs

containers-logs-f:
	@echo Stop following logs with ctrl+c
	docker-compose -f $(COMPOSEFILE) logs -f

containers-build:
	docker-compose -f $(COMPOSEFILE) build

clean-containers: containers-kill
	docker-compose -f $(COMPOSEFILE) rm --force -v

clean:
	-$(RMDIR) $(call FixPath,./bin)
	-$(RMDIR) $(call FixPath,./package)
	-$(RMDIR) $(call FixPath,./db)
