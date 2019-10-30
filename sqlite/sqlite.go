// +build !darwin
// +build !skip_sqlite

// This package depends on go-sqlite3 which wraps the C library which I can't
// get to build for mac so it is excluded with the above build tag.

package sqlite

// Sqlite doesn't support schema so table.schema is ignored throughout

import (
	"github.com/timabell/schema-explorer/driver_interface"
	"github.com/timabell/schema-explorer/drivers"
	"github.com/timabell/schema-explorer/params"
	"github.com/timabell/schema-explorer/reader"
	"github.com/timabell/schema-explorer/schema"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strings"
)

var pathVal = ""

const filePathConfigKey = "file"

var driverOpts = drivers.DriverOpts{
	filePathConfigKey: drivers.DriverOpt{Description: "Path to sqlite db file", Value: &pathVal},
}

func init() {
	reader.RegisterReader(&drivers.Driver{Name: "sqlite", Options: driverOpts, CreateReader: newSqlite, FullName: "SQLite"})
}

type sqliteModel struct {
	path      string
	connected bool // todo: technically it's a connection string per db so we could end up in multiple states, ignore for now
}

func newSqlite() driver_interface.DbReader {
	path := driverOpts[filePathConfigKey].Value
	log.Printf("Connecting to sqlite file: '%s'", *path)
	return sqliteModel{path: *path, connected: false}
}

func (model sqliteModel) ReadSchema(databaseName string) (database *schema.Database, err error) {
	dbc, err := getConnection(model.path)
	if err != nil {
		return
	}
	defer dbc.Close()

	database = &schema.Database{
		Supports: schema.SupportedFeatures{
			Schema:               false,
			Descriptions:         false,
			FkNames:              false, // todo: Get sqlite fk names https://stackoverflow.com/a/42365021/10245
			PagingWithoutSorting: true,
		},
	}

	// load table list
	database.Tables, err = model.getTables(dbc)
	if err != nil {
		return
	}

	// add table columns
	for _, table := range database.Tables {
		var cols []*schema.Column
		cols, err = model.getColumns(dbc, table)
		if err != nil {
			return
		}
		table.Columns = append(table.Columns, cols...)
	}

	// fks
	for _, table := range database.Tables {
		var fks []*schema.Fk
		fks, err = getFks(dbc, table, database)
		if err != nil {
			return
		}
		table.Fks = fks
		database.Fks = append(database.Fks, fks...)
	}

	// hook-up inbound fks
	for _, fk := range database.Fks {
		destination := database.FindTable(fk.DestinationTable)
		if destination == nil {
			err = fmt.Errorf("failed to find destination table for fk %s", fk)
		}
		destination.InboundFks = append(destination.InboundFks, fk)
		for _, destCol := range fk.DestinationColumns {
			destCol.InboundFks = append(destCol.InboundFks, fk)
		}
	}

	// indexes
	for _, table := range database.Tables {
		var indexes []*schema.Index
		indexes, err = getIndexes(dbc, table, database)
		if err != nil {
			return
		}
		table.Indexes = indexes
		database.Indexes = append(database.Indexes, indexes...)
	}

	//log.Print(database.DebugString())
	return
}

func (model sqliteModel) CanSwitchDatabase() bool {
	return false
}

func (model sqliteModel) GetConfiguredDatabaseName() string {
	return ""
}

func (model sqliteModel) ListDatabases() (databaseList []string, err error) {
	panic("not available for sqlite")
}

func (model sqliteModel) DatabaseSelected() bool {
	return true // there is only one
}

func (model sqliteModel) UpdateRowCounts(database *schema.Database) (err error) {
	for _, table := range database.Tables {
		rowCount, err := model.getRowCount(table)
		if err != nil {
			// todo: aggregate errors to return
			log.Printf("Failed to get row count for %s, %s", table, err)
			rowCount = -1
		}
		table.RowCount = &rowCount
	}
	return err
}

func (model sqliteModel) getTables(dbc *sql.DB) (tables []*schema.Table, err error) {
	// todo: parameterise
	rows, err := dbc.Query("SELECT name FROM sqlite_master WHERE type='table' AND name not like 'sqlite_%' order by name;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, &schema.Table{Name: name, Pk: &schema.Pk{}})
	}
	return tables, nil
}

func (model sqliteModel) getRowCount(table *schema.Table) (rowCount int, err error) {
	sql := "select count(*) from \"" + table.Name + "\""

	dbc, err := getConnection(model.path)
	if dbc == nil {
		log.Println(err)
		panic("getConnection() returned nil")
	}
	defer dbc.Close()
	rows, err := dbc.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	rows.Next()
	var count int
	rows.Scan(&count)
	return count, nil
}

func getConnection(path string) (dbc *sql.DB, err error) {
	dbc, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Println("connection error", err)
	}
	return
}

func (model sqliteModel) SetDatabase(databaseName string) {
	// n/a
}

func (model sqliteModel) GetDatabaseName() string {
	return "" // n/a
}

func (model sqliteModel) CheckConnection(databaseName string) (err error) {
	if model.path == "" {
		return errors.New("sqlite file path not set")
	}
	dbc, err := getConnection(model.path)
	if dbc == nil {
		log.Println(err)
		panic("getConnection() returned nil")
	}
	defer dbc.Close()
	err = dbc.Ping()
	if err != nil {
		err = errors.New("database ping failed - " + err.Error())
		return
	}
	tables, err := model.getTables(dbc)
	if err != nil {
		err = errors.New("getTables() failed - " + err.Error())
		return
	}
	if len(tables) == 0 {
		// https://stackoverflow.com/q/45777113/10245
		err = errors.New("no tables found. (SQLite will create an empty db if the specified file doesn't exist)")
		return
	}
	model.connected = true
	log.Println("Connected.", len(tables), "tables found")
	return
}

func (model sqliteModel) Connected() bool {
	return model.connected
}

func getFks(dbc *sql.DB, sourceTable *schema.Table, database *schema.Database) (fks []*schema.Fk, err error) {
	// todo: parameterise
	rows, err := dbc.Query("PRAGMA foreign_key_list('" + sourceTable.Name + "');")
	if err != nil {
		return
	}
	defer rows.Close()
	fks = []*schema.Fk{}
	for rows.Next() {
		var id, seq int
		var destinationTableName, sourceColumnName, destinationColumnName, onUpdate, onDelete, match string
		rows.Scan(&id, &seq, &destinationTableName, &sourceColumnName, &destinationColumnName, &onUpdate, &onDelete, &match)
		_, sourceColumn := sourceTable.FindColumn(sourceColumnName)
		destinationTable := database.FindTable(&schema.Table{Name: destinationTableName})
		_, destinationColumn := destinationTable.FindColumn(destinationColumnName)

		// see if we are adding columns to an existing fk
		var fk *schema.Fk
		for _, existingFk := range fks {
			if existingFk.Id == id {
				existingFk.SourceColumns = append(existingFk.SourceColumns, sourceColumn)
				existingFk.DestinationColumns = append(existingFk.DestinationColumns, destinationColumn)
				fk = existingFk
				break
			}
		}
		if fk == nil {
			fk = &schema.Fk{Id: id, SourceTable: sourceTable, SourceColumns: schema.ColumnList{sourceColumn}, DestinationTable: destinationTable, DestinationColumns: schema.ColumnList{destinationColumn}}
			fks = append(fks, fk)
		}

		sourceColumn.Fks = append(sourceColumn.Fks, fk)
	}
	return
}

func getIndexes(dbc *sql.DB, table *schema.Table, database *schema.Database) (indexes []*schema.Index, err error) {
	rows, err := dbc.Query("PRAGMA index_list('" + table.Name + "');")
	if err != nil {
		return
	}
	defer rows.Close()
	indexes = []*schema.Index{}
	for rows.Next() {
		var seq int
		var name, origin string
		var unique, partial bool
		rows.Scan(&seq, &name, &unique, &origin, &partial)
		if strings.HasPrefix(name, "sqlite_autoindex") {
			continue
		}
		thisIndex := schema.Index{
			Name:     name,
			Table:    table,
			IsUnique: unique,
		}
		err = getIndexInfo(dbc, &thisIndex, table)
		if err != nil {
			return
		}
		indexes = append(indexes, &thisIndex)
	}
	database.Indexes = append(database.Indexes, indexes...)
	return
}

func getIndexInfo(dbc *sql.DB, index *schema.Index, table *schema.Table) (err error) {
	rows, err := dbc.Query("PRAGMA index_info('" + index.Name + "');")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var seqno, cid int
		var colName string
		rows.Scan(&seqno, &cid, &colName)
		if colName != "" {
			_, col := table.FindColumn(colName)
			if col == nil {
				err = errors.New(fmt.Sprintf("can't find col '%s' specified in index %s", colName, index.String()))
				return
			}
			col.Indexes = append(col.Indexes, index)
			index.Columns = append(index.Columns, col)
		}
	}
	return
}

func (model sqliteModel) GetSqlRows(databaseName string, table *schema.Table, params *params.TableParams, peekFinder *driver_interface.PeekLookup) (rows *sql.Rows, err error) {
	dbc, err := getConnection(model.path)
	if err != nil {
		log.Print("GetRows failed to get connection")
		return
	}
	defer dbc.Close()

	sql, values := buildQuery(table, params, peekFinder)
	rows, err = dbc.Query(sql, values...)
	if err != nil {
		log.Print("GetRows failed to get query")
		log.Println(sql)
		log.Println(err)
	}
	return
}

func (model sqliteModel) GetRowCount(databaseName string, table *schema.Table, params *params.TableParams) (rowCount int, err error) {
	dbc, err := getConnection(model.path)
	if err != nil {
		log.Print("GetRows failed to get connection")
		return
	}
	defer dbc.Close()

	sql, values := buildQuery(table, params, &driver_interface.PeekLookup{})
	sql = "select count(*) from (" + sql + ")"
	rows, err := dbc.Query(sql, values...)
	if err != nil {
		log.Print("GetRowCount failed to get query")
		log.Println(sql)
		log.Println(err)
		return
	}
	if !rows.Next() {
		err = errors.New("GetRowCount query returned no rows")
		return
	}
	rows.Scan(&rowCount)
	return
}

func (model sqliteModel) GetAnalysis(databaseName string, table *schema.Table) (analysis []schema.ColumnAnalysis, err error) {
	// todo, might be good to stream this all the way to the http response
	dbc, err := getConnection(model.path)
	if err != nil {
		log.Print("GetAnalysis failed to get connection")
		return
	}
	defer dbc.Close()

	analysis = []schema.ColumnAnalysis{}
	for _, col := range table.Columns {
		sql := "select " + col.Name + ", count(*) qty from " + table.Name + " group by " + col.Name + " order by count(*) desc, " + col.Name + " limit 100;"
		rows, err := dbc.Query(sql)
		if err != nil {
			log.Print("GetAnalysis failed to get query")
			log.Println(sql)
			log.Println(err)
			return nil, err
		}
		var valueInfos []schema.ValueInfo
		for rows.Next() {
			var value interface{}
			var quantity int
			rows.Scan(&value, &quantity)
			valueInfos = append(valueInfos, schema.ValueInfo{
				Value:    value,
				Quantity: quantity,
			})
		}
		analysis = append(analysis, schema.ColumnAnalysis{
			Column:      col,
			ValueCounts: valueInfos,
		})
	}
	return
}

func buildQuery(table *schema.Table, params *params.TableParams, peekFinder *driver_interface.PeekLookup) (sql string, values []interface{}) {
	sql = "select t.*"

	// peek cols
	for fkIndex, fk := range peekFinder.Fks {
		for _, peekCol := range fk.DestinationTable.PeekColumns {
			sql = sql + fmt.Sprintf(", fk%d.[%s] fk%d_%s", fkIndex, peekCol, fkIndex, peekCol)
		}
	}

	// inbound fk counts
	for inboundFkIndex, inboundFk := range table.InboundFks {
		onPredicates := []string{}
		for ix, sourceCol := range inboundFk.SourceColumns {
			onPredicates = append(onPredicates, fmt.Sprintf("ifk%d.[%s] = t.[%s]", inboundFkIndex, sourceCol.Name, inboundFk.DestinationColumns[ix].Name))
		}
		onString := strings.Join(onPredicates, " and ")
		sql = sql + fmt.Sprintf(", (select count(*) from [%s] ifk%d where %s) ifk%d_count", inboundFk.SourceTable, inboundFkIndex, onString, inboundFkIndex)
	}

	sql = sql + " from [" + table.Name + "] t"

	// peek tables
	for fkIndex, fk := range peekFinder.Fks {
		sql = sql + fmt.Sprintf(" left outer join [%s] fk%d on ", fk.DestinationTable.String(), fkIndex)
		onPredicates := []string{}
		for ix, sourceCol := range fk.SourceColumns {
			onPredicates = append(onPredicates, fmt.Sprintf("t.[%s] = fk%d.[%s]", sourceCol, fkIndex, fk.DestinationColumns[ix]))
		}
		onString := strings.Join(onPredicates, " and ")
		sql = sql + onString
	}

	query := params.Filter
	if len(query) > 0 {
		sql = sql + " where "
		clauses := make([]string, 0, len(query))
		values = make([]interface{}, 0, len(query))
		for _, v := range query {
			col := v.Field
			clauses = append(clauses, "t.["+col.Name+"] = ?")
			values = append(values, v.Values[0]) // todo: maybe support multiple values
		}
		sql = sql + strings.Join(clauses, " and ")
	}

	if len(params.Sort) > 0 {
		var sortParts []string
		for _, sortCol := range params.Sort {
			sortString := "t.[" + sortCol.Column.Name + "]"
			if sortCol.Descending {
				sortString = sortString + " desc"
			}
			sortParts = append(sortParts, sortString)
		}
		sql = sql + " order by " + strings.Join(sortParts, ", ")
	}

	if params.RowLimit > 0 || params.SkipRows > 0 {
		sql = sql + fmt.Sprintf(" limit %d, %d", params.SkipRows, params.RowLimit)
	}
	return sql, values
}

func (model sqliteModel) getColumns(dbc *sql.DB, table *schema.Table) (cols []*schema.Column, err error) {
	// todo: parameterise
	rows, err := dbc.Query("PRAGMA table_info('" + table.Name + "');")
	if err != nil {
		return
	}
	defer rows.Close()
	cols = []*schema.Column{}
	colIndex := 0
	for rows.Next() {
		var cid, pk int
		var name, typeName string
		var notNull bool
		var defaultValue interface{}
		rows.Scan(&cid, &name, &typeName, &notNull, &defaultValue, &pk)
		thisCol := schema.Column{
			Position:       colIndex,
			Name:           name,
			Type:           typeName,
			IsInPrimaryKey: pk > 0,
			Nullable:       !notNull,
		}
		cols = append(cols, &thisCol)
		if pk > 0 {
			table.Pk.Columns = append(table.Pk.Columns, &thisCol)
		}
		colIndex++
	}
	return
}

func (model sqliteModel) SetTableDescription(database string, table string, description string) (err error) {
	return
}

func (model sqliteModel) SetColumnDescription(database string, table string, column string, description string) (err error) {
	return
}
