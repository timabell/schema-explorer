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
	expectedName := "DataTypeTest"
	if table.Name != expectedName {
		t.Fatalf("Expected table '%s' found '%s'", expectedName, table.Name)
	}
}

type testCase struct {
	colName        string
	row            int
	expectedType   string
	expectedString string
}

var tests = []testCase{
	testCase{colName: "field_INT", row: 0, expectedType: "INT", expectedString: "20"},
	testCase{colName: "field_INT", row: 1, expectedType: "INT", expectedString: "-33"},
}

// [row][col]
var expectedStrings = [][]string{
	{
		"10",                    //intpk
		"20",                    //field_INT
		"30",                    //field_INTEGER
		"50",                    //field_TINYINT
		"60",                    //field_SMALLINT
		"70",                    //field_MEDIUMINT
		"80",                    //field_BIGINT
		"90",                    //field_UNSIGNED
		"100",                   //field_INT2
		"110",                   //field_INT8
		"field_CHARACTER",       //field_CHARACTER
		"field_VARCHAR",         //field_VARCHAR
		"field_VARYING",         //field_VARYING
		"field_NCHAR",           //field_NCHAR
		"field_NATIVE",          //field_NATIVE
		"field_NVARCHAR",        //field_NVARCHAR
		"field_TEXT",            //field_TEXT
		"field_CLOB",            //field_CLOB
		"field_BLOB",            //field_BLOB
		"field_REAL",            //field_REAL
		"field_DOUBLE",          //field_DOUBLE
		"field_DOUBLEPRECISION", //field_DOUBLEPRECISION
		"field_FLOAT",           //field_FLOAT
		"field_NUMERIC",         //field_NUMERIC
		"field_DECIMAL",         //field_DECIMAL
		"true",                  //field_BOOLEAN
		"field_DATE",            //field_DATE
		"field_DATETIME",        //field_DATETIME
	},
}

func Test_GetRows(t *testing.T) {
	reader := getDbReader(testDbDriver, testDb)
	table := schema.Table{Schema: "dbo", Name: "DataTypeTest"}
	columns, err := reader.GetColumns(table)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := GetRows(reader, nil, table, len(columns), 999)
	if err != nil {
		t.Fatal(err)
	}
	found, countIndex := findCol(columns, "colCount")
	if !found {
		t.Fatal("colCount column missing")
	}
	expectedColCount := int(rows[0][countIndex].(int64))
	if len(columns) != expectedColCount {
		t.Errorf("Expected %#v columns, found %#v", expectedColCount, len(columns))
	}

	for _, test := range tests {
		t.Logf("%+v", test)
		if test.row+1 > len(rows) {
			t.Errorf("Not enough rows. %+v", test)
			continue
		}
		found, columnIndex := findCol(columns, test.colName)
		if !found {
			t.Logf("Skipped test for non-existent column %+v", test)
			continue
		}

		actualType := columns[columnIndex].Type
		if !strings.EqualFold(actualType, test.expectedType) {
			t.Errorf("Incorrect column type %s %+v", actualType, test)
		}
		actualString := DbValueToString(rows[test.row][columnIndex], actualType)
		if *actualString != test.expectedString {
			t.Errorf("Incorrect string '%s' %+v", *actualString, test)
		}
	}
}

func findCol(columns []schema.Column, columnName string) (found bool, index int) {
	for index, col := range columns {
		if col.Name == columnName {
			return true, index
		}
	}
	return false, 0
}

func Test_DataTypes(t *testing.T) {
	// todo: test reads correctly from db
}
func Test_TypeConversion(t *testing.T) {
	// todo: test converts correctly to string
}
