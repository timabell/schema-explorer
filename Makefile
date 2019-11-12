# Makefile for schema-explorer
# Needs python3 IF DOCKER_HOST environment variable is set


# make some commands work on windows as well as *nix
ifdef SystemRoot
	RM = del /q
	RMDIR = rmdir /s /q
	CAT = type
	FixPath = $(subst /,\,$1)
else
	RM = rm -f
	RMDIR = rm -Rf
	FixPath = $1
	CAT = cat
endif

HOSTNAME:=localhost

# Find out DOCKER_HOST address to use as hostname for testing database servers
# Used if your docker_host is not same as localhost
ifdef DOCKER_HOST
	SCRIPT:=python3 -c "from urllib.parse import urlparse; data = urlparse('${DOCKER_HOST}'); print(data.netloc.replace(':' + str(data.port), ''))"
	HOSTNAME:=$(shell $(SCRIPT))
endif

COMPOSEFILE:=test/test-compose.yml
SQLITEDB:=$(call FixPath,db/test.db)
SQLITESCRIPT:=$(call FixPath,sqlite/test-db.sql)

TESTS = testsqlite testmssql

.PHONY: test $(TESTS)

test: $(TESTS)

testmssql:
	go clean -testcache
	go test sse_test.go -driver mssql -mssql-connection-string "sqlserver://sa:GithubIs2broken@${HOSTNAME}?database=ssetest"

testsqlite:
	-$(RMDIR) $(call FixPath,./db)
	mkdir db
	$(CAT) $(SQLITESCRIPT) | sqlite3 $(SQLITEDB)
	go clean -testcache
	go test sse_test.go -driver sqlite -sqlite-file $(SQLITEDB)

testmysql:
	go clean -testcache
	go test sse_test.go -driver mysql -mysql-connection-string "ssetestusr:ssetestusrpass@tcp(${HOSTNAME}:3306)/ssetest"
	
compose-up:
	docker-compose -f $(COMPOSEFILE) up -d

compose-rm: compose-kill
	docker-compose -f $(COMPOSEFILE) rm --force -v

compose-kill:
	docker-compose -f $(COMPOSEFILE) kill

compose-logs:
	docker-compose -f $(COMPOSEFILE) logs

compose-logs-f:
	docker-compose -f $(COMPOSEFILE) logs -f

compose-build:
	docker-compose -f $(COMPOSEFILE) build

clean:
	-$(RMDIR) $(call FixPath,./bin)
	-$(RMDIR) $(call FixPath,./package)
	-$(RMDIR) $(call FixPath,./db)
