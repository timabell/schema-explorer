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
	"github.com/timabell/schema-explorer/driver_interface"
	_ "github.com/timabell/schema-explorer/mssql"
	_ "github.com/timabell/schema-explorer/mysql"
	"github.com/timabell/schema-explorer/options"
	"github.com/timabell/schema-explorer/params"
	_ "github.com/timabell/schema-explorer/pg"
	"github.com/timabell/schema-explorer/reader"
	"github.com/timabell/schema-explorer/schema"
	"github.com/timabell/schema-explorer/serve"
	_ "github.com/timabell/schema-explorer/sqlite"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

var testDb string
var testDbDriver string

func init() {
	options.SetupArgs()
	testing.Init() // so that flags for golang's testing package are defined before we Parse. https://stackoverflow.com/a/58192326/10245
	options.ReadArgsAndEnv()
	//if err != nil {
	//	os.Stderr.WriteString("Note that running sse under test only supports environment variables because command line args clash with the go-test args.\n\n")
	//	options.ArgParser.WriteHelp(os.Stdout)
	//	os.Exit(1)
	//}
	log.Printf("%s is the driver", options.Options.Driver)
}

func Test_CheckConnection(t *testing.T) {
	reader := reader.GetDbReader()
	err := reader.CheckConnection("")
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ReadSchema(t *testing.T) {
	reader := reader.GetDbReader()
	databaseName := getDatabaseName()
	database, err := reader.ReadSchema(databaseName)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Checking table fks")
	checkFks(database, t)

	t.Log("Checking table pks")
	checkTablePks(database, t)

	t.Log("Checking table compound-pks")
	checkTableCompoundPks(database, t)

	t.Log("Checking nullable info")
	checkNullable(database, t)

	t.Log("Checking indexes")
	checkIndexes(database, t)

	if database.Supports.Descriptions {
		t.Log("Checking descriptions")
		checkDescriptions(database, t)
	} else {
		t.Log("Descriptions not supported")
	}

	t.Log("Checking row count")
	checkTableRowCount(reader, database, t)

	t.Log("Checking sort/filter")
	checkFilterAndSort(reader, database, t)

	t.Log("Checking paging")
	checkPaging(reader, database, t)

	t.Log("Checking filtered row count")
	checkFilteredRowCount(reader, database, t)

	t.Log("Checking table analysis")
	checkTableAnalysis(reader, database, t)

	t.Log("Checking keyword escaping")
	checkKeywordEscaping(reader, database, t)

	t.Log("Checking peeking")
	checkPeeking(reader, database, t)

	t.Log("Checking inbound peeking")
	checkInboundPeeking(reader, database, t)
}

func checkIndexes(database *schema.Database, t *testing.T) {
	tableName := "index_test"
	indexName := "IX_compound"

	// check at database level
	if database.Indexes == nil {
		t.Fatal("database.Indexes is nil")
	}
	var databaseIndex *schema.Index
	for _, thisIndex := range database.Indexes {
		if thisIndex.Name == indexName {
			databaseIndex = thisIndex
			break
		}
	}
	if databaseIndex == nil {
		t.Fatalf("Index %s on table %s not found in database.Indexes", indexName, tableName)
	}

	// check at table level
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: tableName}, database, t)
	if table.Indexes == nil {
		t.Fatalf("table.Indexes is nil  on table %s", tableName)
	}
	var tableIndex *schema.Index
	for _, thisIndex := range table.Indexes {
		if thisIndex.Name == indexName {
			tableIndex = thisIndex
			break
		}
	}
	if tableIndex == nil {
		t.Fatalf("Index %s not found on table %s", indexName, tableName)
	}

	// check at column level
	colAName := "compound_a"
	_, colA := table.FindColumn(colAName)
	if colA == nil {
		t.Fatalf("Couldn't find column %s on table %s", colAName, tableName)
	}
	if colA.Indexes == nil {
		t.Fatalf("Column %s on table %s has nil indexes", colAName, tableName)
	}
	checkInt(1, len(colA.Indexes), fmt.Sprintf("indexes on %s.%s", tableName, colAName), t)
	colAIndex := colA.Indexes[0]
	colBName := "compound_b"
	_, colB := table.FindColumn(colBName)
	if colB == nil {
		t.Fatalf("Couldn't find column %s on table %s", colBName, tableName)
	}
	if colB.Indexes == nil {
		t.Fatalf("Column %s on table %s has nil indexes", colBName, tableName)
	}
	checkInt(1, len(colB.Indexes), fmt.Sprintf("indexes on %s.%s", tableName, colBName), t)
	colBIndex := colB.Indexes[0]

	// check index pointers are all pointing to the same thing
	if databaseIndex != tableIndex {
		t.Error("database/table index pointers didn't match")
	}
	if colAIndex != tableIndex {
		t.Error("col/table index pointers didn't match")
	}
	if colAIndex != colBIndex {
		t.Error("col index pointers on the two columns in the index didn't match")
	}

	// now that we know they are all the same thing...
	index := tableIndex
	if index.Table != table {
		log.Fatal("Index not pointing to parent table")
	}
	if index.IsUnique {
		t.Fatalf("%s should not be a unique index", indexName)
	}
	checkInt(2, len(index.Columns), fmt.Sprintf("columns in index %s", indexName), t)
	if index.Columns[0] != colA {
		t.Fatalf("col pointer for %s on index %s didn't match", colA, indexName)
	}
	if index.Columns[1] != colB {
		t.Fatalf("col pointer for %s on index %s didn't match", colB, indexName)
	}

	// unique index
	uniqueIndexName := "IX_unique"
	var uniqueIndex *schema.Index
	for _, thisIndex := range table.Indexes {
		if thisIndex.Name == uniqueIndexName {
			uniqueIndex = thisIndex
			break
		}
	}
	if uniqueIndex == nil {
		log.Fatalf("Didn't find unique index %s", uniqueIndexName)
	}
	if index.IsUnique {
		log.Fatalf("Non-unique index %s was incorrectly flagged as unique", index.Name)
	}
	if !uniqueIndex.IsUnique {
		log.Fatalf("Unique index %s wasn't flagged as unique", uniqueIndexName)
	}
}

func checkNullable(database *schema.Database, t *testing.T) {
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "DataTypeTest"}, database, t)

	notNullColName := "field_not_null_int"
	_, notNullCol := table.FindColumn(notNullColName)
	if notNullCol == nil {
		t.Fatalf("Column %s not found", notNullColName)
	} else if notNullCol.Nullable {
		t.Errorf("%s.%s should not be nullable", table, notNullCol)
	}

	nullColName := "field_null_int"
	_, nullCol := table.FindColumn(nullColName)
	if nullCol == nil {
		t.Fatalf("Column %s not found", nullCol)
	} else if !nullCol.Nullable {
		t.Errorf("%s.%s should be nullable", table, nullCol)
	}
}

func checkTableRowCount(reader driver_interface.DbReader, database *schema.Database, t *testing.T) {
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "SortFilterTest"}, database, t)

	// before load should be nil
	if table.RowCount != nil {
		t.Fatalf("Non-nil row count for table %s before UpdateRowCounts() has been run", table)
	}

	// act
	if err := reader.UpdateRowCounts(database); err != nil {
		t.Error("UpdateRowCounts failed", err)
	}

	// after load should have a value
	var expectRowCountVal = int(7)
	var expectedRowCount = &expectRowCountVal
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
	//t.Logf("%d Pk columns found in table %s", pkLen, table)
	if pkLen != 2 {
		t.Fatalf("Expected 2 Pk columns in table %s, found %d", table, pkLen)
	}

	//t.Logf("%#v", table.Pk)
	//t.Logf("%s - %s", table.Pk.Name, table.Pk.Columns.String())
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
	//t.Logf("%#v", schema.TableDebug(table))
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

func checkFks(database *schema.Database, t *testing.T) {
	childTable := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "FkChild"}, database, t)
	parentTable := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "FkParent"}, database, t)
	// check at table level
	checkInt(1, len(childTable.Fks), "Fks in "+childTable.String(), t)
	childTableFk := childTable.Fks[0]
	checkInt(0, len(parentTable.Fks), "Fks in "+parentTable.String(), t)
	checkInt(0, len(childTable.InboundFks), "InboundFks in "+childTable.String(), t)
	parentTableInboundFk := parentTable.InboundFks[0]
	checkInt(1, len(parentTable.InboundFks), "InboundFks in "+parentTable.String(), t)
	// check at database level
	var dbFk *schema.Fk
	for _, fk := range database.Fks {
		if fk.SourceTable.Name == childTable.Name {
			dbFk = fk
		}
	}
	if dbFk == nil {
		t.Error("Didn't find fk from childTable in database.Fks")
	}
	// check at column level
	colName := "parentId"
	colFullName := fmt.Sprintf("%s.%s", childTable.String(), colName)
	_, fkCol := childTable.FindColumn(colName)
	if fkCol == nil {
		t.Errorf("Checking column fks, column %s not found", colFullName)
	}
	checkInt(1, len(fkCol.Fks), "Fks in "+colFullName, t)
	colFk := fkCol.Fks[0]
	// check inbound column fks
	targetColName := "parentPk"
	targetColFullName := fmt.Sprintf("%s.%s", parentTable.String(), targetColName)
	_, targetFkCol := parentTable.FindColumn(targetColName)
	if targetFkCol == nil {
		t.Errorf("Checking inbound column fks, column %s not found", targetColFullName)
	}
	checkInt(1, len(targetFkCol.InboundFks), "InboundFks in "+targetColFullName, t)
	targetColInboundFk := targetFkCol.InboundFks[0]
	// check fk pointers are all pointing to the same thing
	if childTableFk != parentTableInboundFk {
		t.Error("child/parent fks pointers didn't match")
	}
	if childTableFk != dbFk {
		t.Error("table/database fks pointers didn't match")
	}
	if childTableFk != colFk {
		t.Error("col fk pointer didn't match table fk pointer")
	}
	if childTableFk != targetColInboundFk {
		t.Error("col fk pointer didn't match table fk pointer")
	}
	// now that we know everything has pointers to the same fk...
	fk := childTableFk
	// check contents of fk
	checkStr("FkChild", fk.SourceTable.Name, "fk source table", t)
	checkInt(1, len(fk.SourceColumns), "source cols in fk", t)
	checkStr("parentId", fk.SourceColumns[0].Name, "fk source col name", t)
	checkStr("FkParent", fk.DestinationTable.Name, "fk destination table", t)
	checkInt(1, len(fk.DestinationColumns), "destination cols in fk", t)
	checkStr("parentPk", fk.DestinationColumns[0].Name, "fk destination col name", t)
}

// [actual] [subject], expected [expected]
// e.g. 4 foos in bar, expected 3
func checkInt(expected int, actual int, subject string, t *testing.T) {
	if expected != actual {
		t.Errorf("%d %s expected %d", actual, subject, expected)
	}
}

// [actual] [subject], expected [expected]
// e.g. 4 foos in bar, expected 3
func checkInt64(expected int64, actual int64, subject string, t *testing.T) {
	if expected != actual {
		t.Errorf("%d %s expected %d", actual, subject, expected)
	}
}

// [actual] [subject], expected [expected]
// e.g. 4 foos in bar, expected 3
func checkStr(expected string, actual string, subject string, t *testing.T) {
	if expected != actual {
		t.Errorf("Got '%s' for %s, expected '%s'", actual, subject, expected)
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

func checkFilterAndSort(dbReader driver_interface.DbReader, database *schema.Database, t *testing.T) {
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "SortFilterTest"}, database, t)

	_, patternCol := table.FindColumn("pattern")
	_, sizeCol := table.FindColumn("size")
	_, colourCol := table.FindColumn("colour")
	filter := params.FieldFilter{Field: patternCol, Values: []string{"plain"}}
	tableParams := &params.TableParams{
		Filter:   params.FieldFilterList{filter},
		Sort:     []params.SortCol{{Column: colourCol, Descending: false}, {Column: sizeCol, Descending: true}},
		RowLimit: 10,
	}
	rows, _, err := reader.GetRows(dbReader, database.Name, table, tableParams)
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
				//t.Logf("[%d,%d] %#T==%#T = %t", r, c, expected[r][c], actual[r][c], expected[r][c] == actual[r][c])
			}
		}
		t.Fatal("sort-filter fail")
	}
}

func checkPaging(dbReader driver_interface.DbReader, database *schema.Database, t *testing.T) {
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "SortFilterTest"}, database, t)
	_, idCol := table.FindColumn("id")

	tableParams := &params.TableParams{
		RowLimit: 2,
		SkipRows: 3,
	}
	// check without sort (to check mssql hack for lack of offset capability)
	pagingChecker(dbReader, database.Name, table, tableParams, t, idCol)
	tableParams.Sort = []params.SortCol{{Column: idCol}} // have to sort to use paging for sql server
	// check with sort
	pagingChecker(dbReader, database.Name, table, tableParams, t, idCol)
}

func pagingChecker(dbReader driver_interface.DbReader, databaseName string, table *schema.Table, tableParams *params.TableParams, t *testing.T, idCol *schema.Column) {
	rows, _, err := reader.GetRows(dbReader, databaseName, table, tableParams)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != tableParams.RowLimit {
		t.Errorf("Expected %d limited rows, got %d", tableParams.RowLimit, len(rows))
	}
	checkInt(4, int(rows[0][idCol.Position].(int64)), fmt.Sprintf("for skip %d take %d row 1 id", tableParams.SkipRows, tableParams.RowLimit), t)
	checkInt(5, int(rows[1][idCol.Position].(int64)), fmt.Sprintf("for skip %d take %d row 2 id", tableParams.SkipRows, tableParams.RowLimit), t)
}

func dbString(value interface{}) string {
	return fmt.Sprintf("%s", value)
}

func checkFilteredRowCount(dbReader driver_interface.DbReader, database *schema.Database, t *testing.T) {
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "SortFilterTest"}, database, t)
	_, colourCol := table.FindColumn("colour")
	filter := params.FieldFilter{Field: colourCol, Values: []string{"blue"}}
	tableParams := &params.TableParams{
		Filter: params.FieldFilterList{filter},
	}
	rowCount, err := dbReader.GetRowCount(database.Name, table, tableParams)
	if err != nil {
		t.Fatal(err)
	}
	checkInt(3, rowCount, "blue rows", t)
}

func checkTableAnalysis(dbReader driver_interface.DbReader, database *schema.Database, t *testing.T) {
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "analysis_test"}, database, t)
	colName := "colour"
	_, col := table.FindColumn(colName)
	analysis, err := dbReader.GetAnalysis(database.Name, table)
	if err != nil {
		t.Fatal(err)
	}
	checkInt(1, len(analysis), "columns analysed in "+table.String(), t)
	colourAnalysis := analysis[0]
	checkStr(colName, colourAnalysis.Column.Name, "only col in "+table.String(), t)
	checkInt(4, len(colourAnalysis.ValueCounts), "groups in "+colName+" col in "+table.String(), t)

	expected := []schema.ValueInfo{
		{Quantity: 4, Value: nil},
		{Quantity: 3, Value: "red"},
		{Quantity: 2, Value: "blue"},
		{Quantity: 1, Value: "green"},
	}
	for i, v := range expected {
		if v.Quantity != colourAnalysis.ValueCounts[i].Quantity {
			t.Errorf("expected row %d to have quanty %d, found %d", i, v.Quantity, colourAnalysis.ValueCounts[i].Quantity)
		}
		// Have to convert to string because sqlite returns byte array. Use the canonical conversion from the main codebase
		if v.Value == nil {
			if colourAnalysis.ValueCounts[i].Value != nil {
				t.Errorf("expected row %d to have value %s, found %s", i, v.Value, colourAnalysis.ValueCounts[i].Value)
			}
		} else if v.Value != *reader.DbValueToString(colourAnalysis.ValueCounts[i].Value, col.Type) {
			t.Errorf("expected row %d to have value %s, found %s", i, v.Value, colourAnalysis.ValueCounts[i].Value)
		}
	}
}

// Poke all the things that might fall over if a bit of escaping has been missed.
// The names in here are necessarily confusing and misleading because the table has sql keywords for names.
func checkKeywordEscaping(dbReader driver_interface.DbReader, database *schema.Database, t *testing.T) {
	schemaName := database.DefaultSchemaName
	if database.Supports.Schema {
		schemaName = "identity"
	}

	// test 1 - did it get into the database object at all
	table := findTable(schema.Table{Schema: schemaName, Name: "select"}, database, t)

	// test 2 - did the column show up
	colName := "table"
	_, col := table.FindColumn(colName)
	if col == nil {
		t.Fatalf("Column '%s' not found in keyword table '%s'.", colName, table.String())
	}

	// test 3 - did we get a row count?
	dbReader.UpdateRowCounts(database)
	checkInt(1, *table.RowCount, "row count for keyword table", t)

	// test 4 - can we get the data out with a filter?
	// read the data from it
	filter := params.FieldFilter{Field: col, Values: []string{"times"}}
	params := &params.TableParams{
		RowLimit: 999,
		Filter:   params.FieldFilterList{filter},
		Sort:     []params.SortCol{{Column: col, Descending: false}},
	}
	rows, _, err := reader.GetRows(dbReader, database.Name, table, params)
	if err != nil {
		t.Fatal(err)
	}
	checkInt(1, len(rows), "expected one row in keyword table", t)
	val := fmt.Sprintf("%s", rows[0][1])
	checkStr("times", val, "incorrect value in keyword row", t)
}

func checkPeeking(dbReader driver_interface.DbReader, database *schema.Database, t *testing.T) {
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "peek"}, database, t)
	peekFk := table.Fks[0]
	peekTable := peekFk.DestinationTable
	peekColumn := peekTable.Columns[1]
	peekTable.PeekColumns = append(peekTable.PeekColumns, peekColumn)
	filterColumn := findColumn(table, "dumb_filter", t)

	params := &params.TableParams{
		RowLimit: 999,
		Filter:   params.FieldFilterList{{Field: filterColumn, Values: []string{"filtration"}}}, // add a filter to check where clauses join properly
		Sort:     []params.SortCol{{Column: filterColumn, Descending: false}},                   // add a filter to check order by clauses works with peek joins
	}
	data, peek, err := reader.GetRows(dbReader, database.Name, table, params)
	if err != nil {
		t.Fatal(err)
	}

	// check peek lookup data
	checkInt(1, len(peek.Fks), "peekable fks", t)
	peekIndex := peek.Find(peekFk, peekColumn)
	sourceTableColumnCount := 5             // as per sql files "create table"
	baseIndex := sourceTableColumnCount - 1 // convert from one to zero-based
	peekColumnNumber := 1
	checkInt(baseIndex+peekColumnNumber, peekIndex, "peekIndex", t)

	// check returned peek data
	if data == nil {
		t.Fatal("peek failed: getrows returned nil")
	}
	checkInt(sourceTableColumnCount+1, len(data[0]), "columns in result set", t)
	checkInt(4, len(data), "data rows for peeking at", t)
	checkStr("piggy", fmt.Sprintf("%s", data[0][peekIndex]), "peeked data with string", t)
	if data[1][peekIndex] != nil {
		t.Fatal("peeked data with null in peek table wasn't nil")
	}
	if data[2][peekIndex] != nil {
		t.Fatal("peeked data with null in source")
	}
}

func checkInboundPeeking(dbReader driver_interface.DbReader, database *schema.Database, t *testing.T) {
	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "poke"}, database, t)
	idCol := findColumn(table, "id", t)
	checkInt(2, len(table.InboundFks), "inbound fks on table "+table.String(), t)

	params := &params.TableParams{
		RowLimit: 999,
		Sort:     []params.SortCol{{Column: idCol, Descending: false}},
	}
	data, _, err := reader.GetRows(dbReader, database.Name, table, params)
	if err != nil {
		t.Fatal(err)
	}

	// check inbound peek lookup data
	sourceTableColumnCount := 3             // as per sql files "create table"
	baseIndex := sourceTableColumnCount - 1 // convert from one to zero-based
	var peekColIndex int
	var cozColIndex int
	// location of inbound fk is unstable (different for sqlite vs pg) so need to find it
	for ix, fk := range table.InboundFks {
		switch fk.SourceTable.Name {
		case "peek":
			peekColIndex = baseIndex + 1 + ix
		case "coz":
			cozColIndex = baseIndex + 1 + ix
		}

	}
	if peekColIndex == 0 {
		t.Fatal("Failed to find col index for peekCol for inbound fk")
	}
	if cozColIndex == 0 {
		t.Fatal("Failed to find col index for cozCol for inbound fk")
	}

	// check returned peek data
	if data == nil {
		t.Fatal("peek failed: getrows returned nil")
	}

	//insert into poke (id, name) values (11, 'piggy'); --  one inbound ref
	//insert into poke (id, name) values (12, null);    --  two inbound refs
	//insert into poke (id, name) values (13, 'pie');   -- zero inbound refs

	checkInt(sourceTableColumnCount+2, len(data[0]), "columns in inbound peek result set", t)
	checkInt(3, len(data), "data rows for inbound peeking", t)
	checkInt64(1, data[0][peekColIndex].(int64), "inbound refs for peekFk row id 11", t)
	checkInt64(2, data[1][peekColIndex].(int64), "inbound refs for peekFk row id 12", t)
	checkInt64(0, data[2][peekColIndex].(int64), "inbound refs for peekFk row id 13", t)
	checkInt64(2, data[0][cozColIndex].(int64), "inbound refs for cozFk row id 11", t)
	checkInt64(0, data[1][cozColIndex].(int64), "inbound refs for cozFk row id 12", t)
	checkInt64(0, data[2][cozColIndex].(int64), "inbound refs for cozFk row id 13", t)
}

var tests = []testCase{
	// Some of these might be shared across databases, but they will be in order of appearance.
	// The headings are just to make it easier to navigate the list in reality there will be arbitary sharing.
	// i.e. something in sqlite section might also test the pg database if the results are expected to be the same,
	// that test will not be repeated in the pg section.
	// todo: homogenize type reading - varchar(20) -> "varchar" + length info
	// sqlite
	{colName: "field_int", row: 0, expectedType: "INT", expectedString: "20"},
	{colName: "field_int", row: 1, expectedType: "INT", expectedString: "-33"},
	{colName: "field_integer", row: 0, expectedType: "INTEGER", expectedString: "30"},
	{colName: "field_tinyint", row: 0, expectedType: "TINYINT", expectedString: "50"},
	{colName: "field_smallint", row: 0, expectedType: "SMALLINT", expectedString: "60"},
	{colName: "field_mediumint", row: 0, expectedType: "MEDIUMINT", expectedString: "70"},
	{colName: "field_bigint", row: 0, expectedType: "BIGINT", expectedString: "80"},
	{colName: "field_unsigned", row: 0, expectedType: "UNSIGNED BIG INT", expectedString: "90"},
	{colName: "field_int2", row: 0, expectedType: "INT2", expectedString: "100"},
	{colName: "field_int8", row: 0, expectedType: "INT8", expectedString: "110"},
	{colName: "field_numeric", row: 0, expectedType: "numeric", expectedString: "987.12345"},
	{colName: "field_character", row: 0, expectedType: "CHARACTER(20)", expectedString: "a_CHARACTER"},
	{colName: "field_sqlite_varchar", row: 0, expectedType: "VARCHAR(255)", expectedString: "a_VARCHAR"},
	{colName: "field_varying", row: 0, expectedType: "VARYING CHARACTER(255)", expectedString: "a_VARYING"},
	{colName: "field_nchar", row: 0, expectedType: "NCHAR(55)", expectedString: "a_NCHAR"},
	{colName: "field_native", row: 0, expectedType: "NATIVE CHARACTER(70)", expectedString: "a_NATIVE"},
	{colName: "field_nvarchar", row: 0, expectedType: "NVARCHAR(100)", expectedString: "a_NVARCHAR"},
	{colName: "field_text", row: 0, expectedType: "TEXT", expectedString: "a_TEXT"},
	{colName: "field_clob", row: 0, expectedType: "CLOB", expectedString: "a_CLOB"},
	{colName: "field_blob", row: 0, expectedType: "BLOB", expectedString: "[97 95 66 76 79 66]"},
	{colName: "field_real", row: 0, expectedType: "REAL", expectedString: "1.234"},
	{colName: "field_double", row: 0, expectedType: "DOUBLE", expectedString: "1.234"},
	{colName: "field_doubleprecision", row: 0, expectedType: "DOUBLE PRECISION", expectedString: "1.234"},
	{colName: "field_float", row: 0, expectedType: "FLOAT", expectedString: "1.234"},
	{colName: "field_sqlite_decimal", row: 0, expectedType: "DECIMAL(10,5)", expectedString: "1.234"},
	{colName: "field_boolean", row: 0, expectedType: "BOOLEAN", expectedString: "true"},
	{colName: "field_boolean", row: 1, expectedType: "BOOLEAN", expectedString: "false"},
	// todo: all timezone variant things
	{colName: "field_date", row: 0, expectedType: "DATE", expectedString: "1984-04-02 00:00:00 +0000 UTC"},
	{colName: "field_datetime", row: 0, expectedType: "DATETIME", expectedString: "1984-04-02 11:12:00 +0000 UTC"},
	// pg
	{colName: "field_money", row: 0, expectedType: "money", expectedString: "1234.5670"},
	{colName: "field_pg_decimal", row: 0, expectedType: "decimal", expectedString: "666.1234500"},
	{colName: "field_pg_smallint", row: 0, expectedType: "int2", expectedString: "60"},
	{colName: "field_uniqueidentifier", row: 0, expectedType: "uniqueidentifier", expectedString: "b7a16c7a-a718-4ed8-97cb-20ccbadcc339"},
	{colName: "field_json", row: 0, expectedType: "json", expectedString: "[{\"name\": \"frank\"}, {\"name\": \"sinatra\"}]"},
	{colName: "field_jsonb", row: 0, expectedType: "jsonb", expectedString: "[{\"name\": \"frank\"}, {\"name\": \"sinatra\"}]"},
	// mysql
	{colName: "field_mysql_int", row: 0, expectedType: "int", expectedString: "20"},
	{colName: "field_mysql_character", row: 0, expectedType: "char(20)", expectedString: "a_CHARACTER"},
	{colName: "field_mysql_nchar", row: 0, expectedType: "char(55)", expectedString: "a_NCHAR"},
	{colName: "field_mysql_nvarchar", row: 0, expectedType: "varchar(100)", expectedString: "a_NVARCHAR"},
	{colName: "field_mysql_real", row: 0, expectedType: "double", expectedString: "1.234"},
	{colName: "field_mysql_doubleprecision", row: 0, expectedType: "double", expectedString: "1.234"},
	{colName: "field_mysql_boolean", row: 0, expectedType: "tinyint", expectedString: "1"}, // gah! mysql
}

func Test_GetRows(t *testing.T) {
	dbReader := reader.GetDbReader()
	databaseName := getDatabaseName()
	database, err := dbReader.ReadSchema(databaseName)
	if err != nil {
		t.Fatal(err)
	}

	table := findTable(schema.Table{Schema: database.DefaultSchemaName, Name: "DataTypeTest"}, database, t)

	// read the data from it
	params := &params.TableParams{
		RowLimit: 999,
	}
	rows, _, err := reader.GetRows(dbReader, databaseName, table, params)
	if err != nil {
		t.Fatal(err)
	}

	checkedCols := make(map[string]bool, len(table.Columns))
	checkedCols["intpk"] = true              // not part of the test
	checkedCols["field_not_null_int"] = true // tested in another test
	checkedCols["field_null_int"] = true     // tested in another test

	// check the column count is as expected
	colName := "col_count"
	countIndex, column := table.FindColumn(colName)
	if column == nil {
		t.Fatalf("column missing: %s.%s", table, colName)
	}
	expectedColCount := int(rows[0][countIndex].(int64))
	actualColCount := len(table.Columns)
	if actualColCount != expectedColCount {
		t.Errorf("Expected %#v columns, found %#v", expectedColCount, actualColCount)
	}
	checkedCols["col_count"] = true

	for _, test := range tests {
		if test.row+1 > len(rows) {
			t.Fatalf("Not enough rows. %+v", test)
			continue
		}
		checkedCols[test.colName] = true
		columnIndex, column := table.FindColumn(test.colName)
		if column == nil {
			//t.Logf("Skipped test for non-existent column %+v", test)
			continue
		}

		actualType := table.Columns[columnIndex].Type
		if !strings.EqualFold(actualType, test.expectedType) {
			t.Errorf("Incorrect column type for field '%s': '%s', expected '%s'", test.colName, actualType, test.expectedType)
		}
		// todo: check type of retrieved value, turns out you can put anything you like in sqlite cols
		actualString := reader.DbValueToString(rows[test.row][columnIndex], actualType)
		if actualString == nil {
			t.Errorf("Incorrect nil string %+v, actual data type '%s'", test, actualType)
		} else if actualString == nil || *actualString != test.expectedString {
			t.Errorf("Incorrect string '%+v' %+v, actual data type '%s'", *actualString, test, actualType)
		}
	}
	for _, col := range table.Columns {
		if !checkedCols[col.Name] {
			t.Errorf("col %s.%s was not checked", table, col.Name)
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

// error if not found
func findColumn(table *schema.Table, columnName string, t *testing.T) (column *schema.Column) {
	_, column = table.FindColumn(columnName)
	if column == nil {
		t.Fatalf("column missing %s.%s", table, columnName)
	}
	return
}

func Test_Http(t *testing.T) {
	router, databases := serve.SetupRouter()
	var schemaPrefix string
	var dbPrefix string
	r := reader.GetDbReader()
	var database *schema.Database
	databaseName := getDatabaseName()
	if r.CanSwitchDatabase() {
		reader.InitializeDatabase(databaseName)
		CheckForStatus("/", router, 302, t)
		CheckForOk("/databases", router, t)
		dbPrefix = "/" + databaseName
		database = databases[databaseName]
	} else {
		reader.InitializeDatabase(databaseName)
		database = databases[databaseName]

	}
	// run a get first to populate the schema cache so we can access supported feature list
	CheckForOk(fmt.Sprintf("%s/", dbPrefix), router, t)
	if database.Supports.Schema {
		schemaPrefix = database.DefaultSchemaName + "."
	}
	CheckForOk(fmt.Sprintf("%s/tables/%sDataTypeTest", dbPrefix, schemaPrefix), router, t)
	CheckForOk(fmt.Sprintf("%s/tables/%sDataTypeTest/data", dbPrefix, schemaPrefix), router, t)
	CheckForOk(fmt.Sprintf("%s/tables/%sanalysis_test/analyse-data", dbPrefix, schemaPrefix), router, t)
	CheckForOk(fmt.Sprintf("%s/table-trail", dbPrefix), router, t)
	CheckForStatus("/setup", router, 403, t)
	CheckForStatus("/setup/pg", router, 403, t)
	CheckForStatusWithMethod("/setup/pg", "POST", router, 403, t)

	if database.Supports.Descriptions {
		descriptionTests(dbPrefix, schemaPrefix, router, t, databaseName, database)
	}
}

func descriptionTests(dbPrefix string, schemaPrefix string, router *mux.Router, t *testing.T, databaseName string, database *schema.Database) {
	table := schema.Table{Schema: database.DefaultSchemaName, Name: "person"}
	tableDescription := database.FindTable(&table).Description
	// add
	newDescription := "table-description"
	testTableEndpoint(dbPrefix, schemaPrefix, router, newDescription, t, databaseName, table)
	// update
	newDescription = "table-description-modified"
	testTableEndpoint(dbPrefix, schemaPrefix, router, newDescription, t, databaseName, table)
	// delete
	newDescription = ""
	testTableEndpoint(dbPrefix, schemaPrefix, router, newDescription, t, databaseName, table)

	// restore
	testTableEndpoint(dbPrefix, schemaPrefix, router, tableDescription, t, databaseName, table)

	columnName := "favouritePetId"
	// add
	newDescription = "col-description"
	testColumnEndpoint(dbPrefix, schemaPrefix, router, newDescription, t, databaseName, table, columnName)
	// update
	newDescription = "col-description-modified"
	testColumnEndpoint(dbPrefix, schemaPrefix, router, newDescription, t, databaseName, table, columnName)
	// delete
	newDescription = ""
	testColumnEndpoint(dbPrefix, schemaPrefix, router, newDescription, t, databaseName, table, columnName)
}

func testTableEndpoint(dbPrefix string, schemaPrefix string, router *mux.Router, newDescription string, t *testing.T, databaseName string, table schema.Table) {
	tableEndpoint := fmt.Sprintf("%s/tables/%sperson/description", dbPrefix, schemaPrefix)
	testDocEndpoint(tableEndpoint, router, newDescription, t, databaseName, table)

	reader.InitializeDatabase(databaseName)
	updatedDescription := reader.Databases[databaseName].FindTable(&table).Description
	checkStr(newDescription, updatedDescription, "description of "+table.String(), t)
}

func testColumnEndpoint(dbPrefix string, schemaPrefix string, router *mux.Router, newDescription string, t *testing.T, databaseName string, table schema.Table, columnName string) {
	colEndpoint := fmt.Sprintf("%s/tables/%sperson/columns/%s/description", dbPrefix, schemaPrefix, columnName)
	testDocEndpoint(colEndpoint, router, newDescription, t, databaseName, table)

	reader.InitializeDatabase(databaseName)
	_, col := reader.Databases[databaseName].FindTable(&table).FindColumn(columnName)
	updatedDescription := col.Description
	checkStr(newDescription, updatedDescription, "description of "+table.String(), t)
}

func testDocEndpoint(docEndpoint string, router *mux.Router, newDescription string, t *testing.T, databaseName string, table schema.Table) {
	t.Logf("testing %s with description '%s'", docEndpoint, newDescription)
	CheckForStatusWithMethodAndBody(docEndpoint, "POST", router, 200, newDescription, t)
}

func getDatabaseName() string {
	r := reader.GetDbReader()
	if r.CanSwitchDatabase() {
		return "ssetest"
	}
	return ""
}

func CheckForOk(path string, router *mux.Router, t *testing.T) {
	CheckForStatus(path, router, 200, t)
}

func CheckForStatus(path string, router *mux.Router, expectedStatus int, t *testing.T) {
	CheckForStatusWithMethod(path, "GET", router, expectedStatus, t)
}

func CheckForStatusWithMethod(path string, method string, router *mux.Router, expectedStatus int, t *testing.T) {
	CheckForStatusWithMethodAndBody(path, "GET", router, expectedStatus, "", t)

}
func CheckForStatusWithMethodAndBody(path string, method string, router *mux.Router, expectedStatus int, body string, t *testing.T) {
	request, _ := http.NewRequest(method, path, strings.NewReader(body))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != expectedStatus {
		t.Fatalf("%d status for %s, expected %d", response.Code, path, expectedStatus)
	}
}
