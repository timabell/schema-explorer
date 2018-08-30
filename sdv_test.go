package main

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
	_ "bitbucket.org/timabell/sql-data-viewer/mssql"
	"bitbucket.org/timabell/sql-data-viewer/params"
	_ "bitbucket.org/timabell/sql-data-viewer/pg"
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	_ "bitbucket.org/timabell/sql-data-viewer/sqlite"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
)

var testDb string
var testDbDriver string

func init() {
	_, err := reader.ArgParser.ParseArgs([]string{})
	if err != nil {
		os.Stderr.WriteString("Note that running sdv under test only supports environment variables because command line args clash with the go-test args.\n\n")
		reader.ArgParser.WriteHelp(os.Stdout)
		os.Exit(1)
	}
	log.Printf("%s is the driver", *reader.Options.Driver)
}

func Test_CheckConnection(t *testing.T) {
	reader := reader.GetDbReader()
	err := reader.CheckConnection()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ReadSchema(t *testing.T) {
	reader := reader.GetDbReader()
	database, err := reader.ReadSchema()
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Checking fk count")
	checkFkCount(database, t)
	t.Log("Checking table pks")
	checkTablePks(database, t)
	t.Log("Checking table compound-pks")
	checkTableCompoundPks(database, t)
	t.Log("Checking table fks")
	checkTableFks(database, t)
	t.Log("Checking row count")
	checkTableRowCount(database, t)
	t.Log("Checking inbound fk count")
	checkInboundTableFkCount(database, t)
	t.Log("Checking column fk count")
	checkColumnFkCount(database, t)
	t.Log("Checking nullable info")
	checkNullable(database, t)
	if database.Supports.Descriptions {
		t.Log("Checking descriptions")
		checkDescriptions(database, t)
	} else {
		t.Log("Descriptions not supported")
	}
}

func checkNullable(database *schema.Database, t *testing.T) {
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "DataTypeTest"}, database, t)
	_, notNullCol := table.FindColumn("field_NotNullInt")
	if notNullCol == nil {
		t.Log(schema.TableDebug(table))
		t.Fatal("Column field_NotNullInt not found")
	}
	if notNullCol.Nullable {
		t.Errorf("%s.%s should not be nullable", table, notNullCol)
	}
	_, nullCol := table.FindColumn("field_NullInt")
	if notNullCol == nil {
		t.Log(schema.TableDebug(table))
		t.Fatal("Column field_NullInt not found")
	}
	if !nullCol.Nullable {
		t.Errorf("%s.%s should be nullable", table, nullCol)
	}

}
func checkColumnFkCount(database *schema.Database, t *testing.T) {
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "pet"}, database, t)
	_, col := table.FindColumn("ownerId")
	if col == nil {
		t.Log(schema.TableDebug(table))
		t.Fatal("Column ownderId not found while checking col fk count")
	}
	if col.Fks == nil {
		t.Log(schema.TableDebug(table))
		t.Logf("%#v", col)
		t.Log(col.Fks)
		t.Errorf("Fks entry missing from column %s.%s", table, col)
	}
}

func checkFkCount(database *schema.Database, t *testing.T) {
	expectedCount := 6
	fkCount := len(database.Fks)
	if fkCount != expectedCount {
		t.Fatalf("Expected %d fks across whole db, found %d", expectedCount, fkCount)
	}
}

func checkInboundTableFkCount(database *schema.Database, t *testing.T) {
	expectedInboundCount := 2
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "person"}, database, t)
	fkCount := len(table.InboundFks)
	if fkCount != expectedInboundCount {
		t.Fatalf("Expected %d inboundFks in table %s, found %d", expectedInboundCount, table, fkCount)
	}
}

func checkTableRowCount(database *schema.Database, t *testing.T) {
	var expectRowCountVal = int(7)
	var expectedRowCount = &expectRowCountVal
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "SortFilterTest"}, database, t)
	if table.RowCount == nil {
		t.Fatalf("Nil row count for table %s", table)
	}
	if *table.RowCount != *expectedRowCount {
		t.Fatalf("Expected row count of %d for table %s, found %d", *expectedRowCount, table, *table.RowCount)
	}
}

func checkTableCompoundPks(database *schema.Database, t *testing.T) {
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "CompoundKeyParent"}, database, t)
	if table.Pk == nil {
		t.Fatalf("Nil Pk in table %s", table)
	}
	pkLen := len(table.Pk.Columns)
	t.Logf("%d Pk columns found in table %s", pkLen, table)
	if pkLen != 2 {
		t.Fatalf("Expected 2 Pk columns in table %s, found %d", table, pkLen)
	}

	t.Logf("%#v", table.Pk)
	t.Logf("%s - %s", table.Pk.Name, table.Pk.Columns.String())
	expectedPkCol1 := "colA"
	pkColumn := table.Pk.Columns[0]
	if pkColumn.Name != expectedPkCol1 {
		t.Fatalf("Expected the first columnn in pk of %s to be %s, found %s", table, expectedPkCol1, pkColumn.Name)
	}
	if !pkColumn.IsInPrimaryKey {
		t.Fatalf("%s.%s not marked as primary key", table, pkColumn.Name)
	}

	expectedPkColPosition := 2
	if pkColumn.Position != expectedPkColPosition {
		t.Fatalf("Expected the first columnn in pk of %s to have position %d, found %d", table, expectedPkColPosition, pkColumn.Position)
	}

	expectedPkCol2 := "colB"
	pkColumn = table.Pk.Columns[1]
	if pkColumn.Name != expectedPkCol2 {
		t.Fatalf("Expected the second columnn in pk of %s to be %s, found %s", table, expectedPkCol2, pkColumn.Name)
	}
	if !pkColumn.IsInPrimaryKey {
		t.Fatalf("%s.%s not marked as primary key", table, pkColumn.Name)
	}

	nonPkColumn := table.Columns[0]
	if nonPkColumn.IsInPrimaryKey {
		t.Fatalf("%s.%s should not be marked as primary key", table, nonPkColumn.Name)
	}
}

func checkTablePks(database *schema.Database, t *testing.T) {
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "pet"}, database, t)
	t.Logf("%#v", schema.TableDebug(table))
	if table.Pk == nil {
		t.Fatalf("Nil Pk in table %s", table)
	}
	pkLen := len(table.Pk.Columns)
	if pkLen != 1 {
		t.Fatalf("Expected 1 Pk column table %s, found %d", table, pkLen)
	}
	pkColumn := table.Pk.Columns[0]
	expectedPkCol := "petId"
	if pkColumn.Name != expectedPkCol {
		t.Fatalf("Expected the only columnn in pk of %s to be %s, found %s", table, expectedPkCol, pkColumn.Name)
	}
	if !pkColumn.IsInPrimaryKey {
		t.Fatalf("%s.%s not marked as primary key", table, pkColumn.Name)
	}
	nonPkColumn := table.Columns[1]
	if nonPkColumn.IsInPrimaryKey {
		t.Fatalf("%s.%s should not be marked as primary key", table, nonPkColumn.Name)
	}
}

func checkTableFks(database *schema.Database, t *testing.T) {
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

func checkDescriptions(database *schema.Database, t *testing.T) {
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

func Test_FilterAndSort(t *testing.T) {
	dbReader := reader.GetDbReader()
	database, err := dbReader.ReadSchema()
	if err != nil {
		t.Fatal(err)
	}

	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "SortFilterTest"}, database, t)

	_, patternCol := table.FindColumn("pattern")
	_, sizeCol := table.FindColumn("size")
	_, colourCol := table.FindColumn("colour")
	filter := params.FieldFilter{Field: patternCol, Values: []string{"plain"}}
	log.Print(filter)
	tableParams := &params.TableParams{
		Filter:   params.FieldFilterList{filter},
		Sort:     []params.SortCol{{Column: colourCol, Descending: false}, {Column: sizeCol, Descending: true}},
		RowLimit: 10,
	}
	rows, err := reader.GetRows(dbReader, table, tableParams)
	if err != nil {
		t.Fatal(err)
	}

	expectedRowCount := 4
	if len(rows) != expectedRowCount {
		t.Errorf("Expected %d filterd rows, got %d", expectedRowCount, len(rows))
	}

	expected := [][]interface{}{
		{int64(5), int64(23), "blue", "plain"},
		{int64(6), int64(22), "blue", "plain"},
		{int64(4), int64(21), "blue", "plain"},
		{int64(3), int64(2), "green", "plain"},
	}
	var actual [][]interface{} = nil
	for _, row := range rows {
		actual = append(actual, []interface{}{row[0], row[1], dbString(row[2]), dbString(row[3])})
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Logf("expected: %+v", expected)
		t.Logf("actual:   %+v", actual)
		for r := 0; r < len(expected); r++ {
			for c := 0; c < len(expected[r]); c++ {
				t.Logf("[%d,%d] %#v==%#v = %t", r, c, expected[r][c], actual[r][c], expected[r][c] == actual[r][c])
				t.Logf("[%d,%d] %#T==%#T = %t", r, c, expected[r][c], actual[r][c], expected[r][c] == actual[r][c])
			}
		}
		t.Fatal("sort-filter fail")
	}
}

func dbString(value interface{}) string {
	return fmt.Sprintf("%s", value)
}

func Test_GetRows(t *testing.T) {
	dbReader := reader.GetDbReader()
	database, err := dbReader.ReadSchema()
	if err != nil {
		t.Fatal(err)
	}

	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "DataTypeTest"}, database, t)

	// read the data from it
	params := &params.TableParams{
		RowLimit: 999,
	}
	rows, err := reader.GetRows(dbReader, table, params)
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
		actualString := reader.DbValueToString(rows[test.row][columnIndex], actualType)
		if *actualString != test.expectedString {
			t.Errorf("Incorrect string '%s' %+v", *actualString, test)
		}
	}
}

// error if not found
func findTable(tableToFind schema.Table, database *schema.Database, t *testing.T) *schema.Table {
	table := database.FindTable(&tableToFind)
	if table == nil {
		t.Fatal(tableToFind.String() + " table missing")
	}
	return table
}