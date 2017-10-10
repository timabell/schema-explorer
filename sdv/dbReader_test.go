package sdv

import (
	"flag"
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
		t.Fatal(err)
	}
}

func Test_GetTables(t *testing.T) {
	reader := getDbReader(testDbDriver, testDb)
	tables, err := reader.GetTables()
	if err != nil {
		t.Fatal(err)
	}
	expectedCount := 1
	if len(tables) != expectedCount {
		t.Fatalf("Expected %d tables, found %d", expectedCount, len(tables))
	}
	table := tables[0]
	expectedName := "foo"
	if table.Name != expectedName {
		t.Fatalf("Expected table '%s' found '%s'", expectedName, table.Name)
	}
}

func Test_GetColumns(t *testing.T) {
	reader := getDbReader(testDbDriver, testDb)
	table := schema.Table{Schema: "dbo", Name: "foo"}
	columns, err := reader.GetColumns(table)
	if err != nil {
		t.Fatal(err)
	}
	expectedCount := 3
	if len(columns) != expectedCount {
		t.Fatalf("Expected %d columns, found %d", expectedCount, len(columns))
	}
	col0 := columns[0]
	expectedName := "id"
	if col0.Name != expectedName {
		t.Fatalf("Expected column '%s' found '%s'", expectedName, col0)
	}
}

func Test_GetRows(t *testing.T) {
	reader := getDbReader(testDbDriver, testDb)
	rowCount := 1
	table := schema.Table{Schema: "dbo", Name: "foo"}
	columns, err := reader.GetColumns(table)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := GetRows(reader, nil, table, len(columns), rowCount+1)
	if len(rows) != rowCount {
		t.Errorf("Expected %d rows got %d", rowCount, len(rows))
	}
	if err != nil {
		t.Fatal(err)
	}
	rowData := rows[0]
	expectedId := "1"
	expectedName := "raaa"
	expectedColour := "blue"
	actualId := DbValueToString(rowData[0], columns[0].Type)
	actualName := DbValueToString(rowData[1], columns[1].Type)
	actualColour := DbValueToString(rowData[2], columns[2].Type)
	if *actualId != expectedId {
		t.Errorf("Row 1 col id expected %d got %d", expectedId, actualId)
	}
	if *actualName != expectedName {
		t.Error("Row 1 col name table foo, incorrect data; expected:", expectedName, "actual:", actualName)
	}
	if *actualColour != expectedColour {
		t.Error("Row 1 col colour table foo, incorrect data; expected:", expectedColour, "actual:", actualColour)
	}
}

func Test_DataTypes(t *testing.T) {
	// todo: test reads correctly from db
}
func Test_TypeConversion(t *testing.T) {
	// todo: test converts correctly to string
}
