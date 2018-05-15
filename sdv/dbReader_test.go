package sdv

/*
Tests multiple rdbms implementations by way of test.sh shell script that repeatedly runs the same
tests for each supported rdbms.
Relies on matching sql files having been run to set up each test database.

The tests are testing pulling data from a real database (integration testing) because
the layer between the code and the database is the most fragile.
The tests do not cover the UI layer beyond translation of data from the database into
strings for display.

In order to test different databases where they support an overlapping but not identical
set of data types the following strategy is used:

Every supported db system gets a table with a column for each data type that is supported by
that rdbms, named to match, then the test code tests the conversion to string for each
available data type. This allows data types that are common to be tested with a single test
but still support data types that are unique to a particular rdbms.

The expected number of cols is included in an extra column so we can double-check that we
aren't silently missing any of the supported data types.
*/

import (
	"flag"
	"testing"

	"bitbucket.org/timabell/sql-data-viewer/schema"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/simnalamburt/go-mssqldb"
	"log"
	"strings"
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

func Test_ReadSchema(t *testing.T) {
	reader := getDbReader(testDbDriver, testDb)
	database, err := reader.ReadSchema()
	if err != nil {
		t.Fatal(err)
	}

	checkFkCount(database, t)
	checkTableFks(database, t)
	checkInboundTableFkCount(database, t)
	checkColumnFkCount(database, t)
	if database.Supports.Descriptions {
		checkDescriptions(database, t)
	}
}
func checkColumnFkCount(database schema.Database, t *testing.T) {
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "pet"}, database, t)
	_, col := table.FindColumn("ownerId")
	if col == nil {
		t.Log(schema.TableDebug(table))
		t.Fatal("Column ownderId not found while checking col fk count")
	}
	if col.Fk == nil {
		t.Log(schema.TableDebug(table))
		t.Logf("%#v", col)
		t.Log(col.Fk)
		t.Errorf("Fk entry missing from column %s.%s", table, col)
	}
}

func checkFkCount(database schema.Database, t *testing.T) {
	expectedCount := 4
	fkCount := len(database.Fks)
	if fkCount != expectedCount {
		t.Fatalf("Expected %d fks across whole db, found %d", expectedCount, fkCount)
	}
}

func checkInboundTableFkCount(database schema.Database, t *testing.T) {
	expectedInboundCount := 2
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "person"}, database, t)
	fkCount := len(table.InboundFks)
	if fkCount != expectedInboundCount {
		t.Fatalf("Expected %d inboundFks in table %s, found %d", expectedInboundCount, table, fkCount)
	}
}

func checkTableFks(database schema.Database, t *testing.T) {
	expectedFkCount := 2
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "pet"}, database, t)
	fkCount := len(table.Fks)
	if fkCount != expectedFkCount {
		t.Fatalf("Expected %d fks in table %s, found %d", expectedFkCount, table, fkCount)
	}
}

type descriptionCase struct {
	schema      string
	table       string
	column      string
	description string
}

func checkDescriptions(database schema.Database, t *testing.T) {
	var descriptions = []descriptionCase{
		{schema: database.DefaultSchemaName, table: "person", column: "", description: "somebody to love"},
		{schema: database.DefaultSchemaName, table: "person", column: "personName", description: "say my name!"},
		{schema: "kitchen", table: "sink", column: "", description: "call a plumber!!!"},
		{schema: "kitchen", table: "sink", column: "sinkId", description: "gotta number your sinks man!"},
	}

	for _, testCase := range descriptions {
		log.Println(testCase)
		table := findTable(schema.Table{Schema: testCase.schema, Name: testCase.table}, database, t)
		if testCase.column == "" {
			if table.Description != testCase.description {
				t.Errorf("Expected description for table '%s' of '%s', got '%s'", table, testCase.description, table.Description)
			}
		} else {
			_, col := table.FindColumn(testCase.column)
			if col.Description != testCase.description {
				t.Errorf("Expected description for column '%s' table '%s' of '%s', got '%s'", col, table, testCase.description, col.Description)
			}
		}
	}
}

type testCase struct {
	colName        string
	row            int
	expectedType   string
	expectedString string
}

var tests = []testCase{
	{colName: "field_INT", row: 0, expectedType: "int", expectedString: "20"},
	{colName: "field_INT", row: 1, expectedType: "int", expectedString: "-33"},
	{colName: "field_money", row: 0, expectedType: "money", expectedString: "1234.5670"},
	{colName: "field_numeric", row: 0, expectedType: "numeric", expectedString: "987.1234500"},
	{colName: "field_decimal", row: 0, expectedType: "decimal", expectedString: "666.1234500"},
	{colName: "field_uniqueidentifier", row: 0, expectedType: "uniqueidentifier", expectedString: "b7a16c7a-a718-4ed8-97cb-20ccbadcc339"},
}

func Test_GetRows(t *testing.T) {
	reader := getDbReader(testDbDriver, testDb)
	database, err := reader.ReadSchema()
	if err != nil {
		t.Fatal(err)
	}

	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "DataTypeTest"}, database, t)

	// read the data from it
	rows, err := GetRows(reader, nil, table, 999)
	if err != nil {
		t.Fatal(err)
	}

	// check the column count is as expected
	countIndex, column := table.FindColumn("colCount")
	if column == nil {
		t.Fatal("colCount column missing from " + table.String())
	}
	expectedColCount := int(rows[0][countIndex].(int64))
	actualColCount := len(table.Columns)
	if actualColCount != expectedColCount {
		t.Errorf("Expected %#v columns, found %#v", expectedColCount, actualColCount)
	}

	for _, test := range tests {
		if test.row+1 > len(rows) {
			t.Errorf("Not enough rows. %+v", test)
			continue
		}
		columnIndex, column := table.FindColumn(test.colName)
		if column == nil {
			t.Logf("Skipped test for non-existent column %+v", test)
			continue
		}

		actualType := table.Columns[columnIndex].Type
		if !strings.EqualFold(actualType, test.expectedType) {
			t.Errorf("Incorrect column type %s %+v", actualType, test)
		}
		actualString := DbValueToString(rows[test.row][columnIndex], actualType)
		if *actualString != test.expectedString {
			t.Errorf("Incorrect string '%s' %+v", *actualString, test)
		}
	}
}

// error if not found
func findTable(tableToFind schema.Table, database schema.Database, t *testing.T) *schema.Table {
	table := database.FindTable(&tableToFind)
	if table == nil {
		t.Fatal(tableToFind.String() + " table missing")
	}
	return table
}
