package sqlite

// Sqlite doesn't support schema so table.schema is ignored throughout

import (
	"bitbucket.org/timabell/sql-data-viewer/params"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strconv"
	"strings"
)

type sqliteModel struct {
	path string
}

func NewSqlite(path string) sqliteModel {
	return sqliteModel{
		path: path,
	}
}

func (model sqliteModel) ReadSchema() (database *schema.Database, err error) {
	dbc, err := getConnection(model.path)
	if err != nil {
		return
	}
	defer dbc.Close()

	database = &schema.Database{
		Supports: schema.SupportedFeatures{
			Schema:       true,
			Descriptions: false,
			FkNames:      false, // todo: Get sqlite fk names https://stackoverflow.com/a/42365021/10245
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
	}
	//log.Print(database.DebugString())
	return
}

func (model sqliteModel) getTables(dbc *sql.DB) (tables []*schema.Table, err error) {
	// todo: parameterise
	rows, err := dbc.Query("SELECT name FROM sqlite_master WHERE type='table' AND name not like 'sqlite_%';")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, &schema.Table{Name: name, Pk: &schema.Pk{}})
	}
	for _, table := range tables {
		rowCount, err := model.getRowCount(table)
		if err != nil {
			log.Printf("Failed to get row count for %d", table)
		}
		table.RowCount = &rowCount
	}
	return tables, nil
}

func (model sqliteModel) getRowCount(table *schema.Table) (rowCount int, err error) {
	// todo: parameterise where possible
	// todo: whitelist-sanitize unparameterizable parts
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

func (model sqliteModel) CheckConnection() (err error) {
	dbc, err := getConnection(model.path)
	if dbc == nil {
		log.Println(err)
		panic("getConnection() returned nil")
	}
	defer dbc.Close()
	tables, err := model.getTables(dbc)
	if err != nil {
		panic(err)
	}
	if len(tables) == 0 {
		// https://stackoverflow.com/q/45777113/10245
		panic("No tables found. (Sqlite will create an empty db if the specified file doesn't exist).")
	}
	log.Println("Connected.", len(tables), "tables found")
	return
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

func (model sqliteModel) GetSqlRows(table *schema.Table, params *params.TableParams) (rows *sql.Rows, err error) {
	// todo: parameterise where possible
	// todo: whitelist-sanitize unparameterizable parts
	sql := "select * from " + table.Name

	var values []interface{}
	query := params.Filter
	if len(query) > 0 {
		sql = sql + " where "
		clauses := make([]string, 0, len(query))
		values = make([]interface{}, 0, len(query))
		for _, v := range query {
			col := v.Field
			clauses = append(clauses, col.Name+" = ?")
			values = append(values, v.Values[0]) // todo: maybe support multiple values
		}
		sql = sql + strings.Join(clauses, " and ")
	}

	if len(params.Sort) > 0 {
		var sortParts []string
		for _, sortCol := range params.Sort {
			sortString := sortCol.Column.Name
			if sortCol.Descending {
				sortString = sortString + " desc"
			}
			sortParts = append(sortParts, sortString)
		}
		sql = sql + " order by " + strings.Join(sortParts, ", ")
	}

	rowLimit := params.RowLimit
	if rowLimit > 0 {
		sql = sql + " limit " + strconv.Itoa(rowLimit)
	}

	dbc, err := getConnection(model.path)
	if err != nil {
		log.Print("GetRows failed to get connection")
		return
	}
	defer dbc.Close()

	rows, err = dbc.Query(sql, values...)
	if err != nil {
		log.Print("GetRows failed to get query")
		log.Println(sql)
		log.Println(err)
	}
	return
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
		thisCol := schema.Column{Index: colIndex, Name: name, Type: typeName, IsInPrimaryKey: pk > 0}
		cols = append(cols, &thisCol)
		if pk > 0 {
			table.Pk.Columns = append(table.Pk.Columns, &thisCol)
		}
		colIndex++
	}
	return
}
