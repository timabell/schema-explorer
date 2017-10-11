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
	expectedName := "DataTypeTest"
	if table.Name != expectedName {
		t.Fatalf("Expected table '%s' found '%s'", expectedName, table.Name)
	}
}

// [row][col]
var expectedStrings = [][]string {
	{
"10", //intpk
"20", //field_INT
"30", //field_INTEGER
"50", //field_TINYINT
"60", //field_SMALLINT
"70", //field_MEDIUMINT
"80", //field_BIGINT
"90", //field_UNSIGNED
"100", //field_INT2
"110", //field_INT8
"field_CHARACTER", //field_CHARACTER
"field_VARCHAR", //field_VARCHAR
"field_VARYING", //field_VARYING
"field_NCHAR", //field_NCHAR
"field_NATIVE", //field_NATIVE
"field_NVARCHAR", //field_NVARCHAR
"field_TEXT", //field_TEXT
"field_CLOB", //field_CLOB
"field_BLOB", //field_BLOB
"field_REAL", //field_REAL
"field_DOUBLE", //field_DOUBLE
"field_DOUBLEPRECISION", //field_DOUBLEPRECISION
"field_FLOAT", //field_FLOAT
"field_NUMERIC", //field_NUMERIC
"field_DECIMAL", //field_DECIMAL
"true", //field_BOOLEAN
"field_DATE", //field_DATE
"field_DATETIME", //field_DATETIME
 	},
}

func Test_GetRows(t *testing.T) {
	reader := getDbReader(testDbDriver, testDb)
	table := schema.Table{Schema: "dbo", Name: "DataTypeTest"}
	columns, err := reader.GetColumns(table)
	if err != nil {
		t.Fatal(err)
	}
	rowCount := len(expectedStrings)
	rows, err := GetRows(reader, nil, table, len(columns), rowCount+1)
	if len(rows) != rowCount {
		t.Errorf("Expected %d rows got %d", rowCount, len(rows))
	}
	if err != nil {
		t.Fatal(err)
	}
	for r, row := range expectedStrings {
		for c, expected := range row{
			t.Logf("%d,%d: %s", r, c, expected)
			actual := DbValueToString(rows[r][c], columns[c].Type)
			if *actual != expected {
				t.Errorf("Row %d col %d name table DataTypeTest, incorrect data; expected: '%s' actual: '%s'", r, c, expected, *actual)
			}
		}
	}
}

func Test_DataTypes(t *testing.T) {
	// todo: test reads correctly from db
}
func Test_TypeConversion(t *testing.T) {
	// todo: test converts correctly to string
}
