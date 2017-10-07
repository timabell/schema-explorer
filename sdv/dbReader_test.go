package sdv

import (
	"flag"
	"fmt"
	"testing"

	"bitbucket.org/timabell/sql-data-viewer/schema"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/simnalamburt/go-mssqldb"
)

var testDb string
var testDbDriver string

func init() {
	flag.StringVar(&testDbDriver, "driver", "", "Driver to use (mssql or sqlite)")
	flag.StringVar(&testDb, "db", "", "connection string for mssql / filename for sqlite")
	flag.Parse()
	if testDbDriver == "" {
		flag.Usage()
		panic("Driver argument required.")
	}
	if testDb == "" {
		flag.Usage()
		panic("Db argument required.")
	}
}

func Test_CheckConnection(t *testing.T) {
	reader := getDbReader(testDbDriver, testDb)
	err := reader.CheckConnection()
	if err != nil {
		t.Error(err)
	}
}

func Test_GetTables(t *testing.T) {
	reader := getDbReader(testDbDriver, testDb)
	tables, err := reader.GetTables()
	if err != nil {
		t.Error(err)
	}
	expectedCount := 1
	if len(tables) != expectedCount {
		t.Error(fmt.Sprintf("Expected %d tables, found %d", expectedCount, len(tables)))
	}
	table := tables[0]
	expectedName := "foo"
	if table.Name != expectedName {
		t.Error(fmt.Sprintf("Expected table '%s' found '%s'", expectedName, table.Name))
	}
}

func Test_GetColumns(t *testing.T) {
	reader := getDbReader(testDbDriver, testDb)
	table := schema.Table{Name: "foo"}
	columns, err := reader.GetColumns(table)
	if err != nil {
		t.Error(err)
	}
	expectedCount := 2
	if len(columns) != expectedCount {
		t.Error(fmt.Sprintf("Expected %d columns, found %d", expectedCount, len(columns)))
	}
	col0 := columns[0]
	expectedName := "id"
	if col0.Name != expectedName {
		t.Error(fmt.Sprintf("Expected column '%s' found '%s'", expectedName, col0))
	}
}
