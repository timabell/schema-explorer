package pg

import (
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"strconv"
	"strings"
)

type pgModel struct {
	connectionString string
}

func NewPg(connectionString string) pgModel {
	return pgModel{
		connectionString: connectionString,
	}
}

func (model pgModel) ReadSchema() (database schema.Database, err error) {
	dbc, err := getConnection(model.connectionString)
	if err != nil {
		return
	}
	defer dbc.Close()

	database = schema.Database{Supports: schema.SupportedFeatures{Schema: false}}

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

func (model pgModel) getTables(dbc *sql.DB) (tables []*schema.Table, err error) {
	rows, err := dbc.Query("select table_schema, table_name from information_schema.tables where table_type='BASE TABLE' and table_schema not in ('pg_catalog','information_schema')")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name, schemaName string
		rows.Scan(&schemaName, &name)
		tables = append(tables, &schema.Table{Schema: schemaName, Name: name})
	}
	return tables, nil
}

func getConnection(connectionString string) (dbc *sql.DB, err error) {
	dbc, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Println("connection error", err)
	}
	return
}

func (model pgModel) CheckConnection() (err error) {
	dbc, err := getConnection(model.connectionString)
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
		panic("No tables found.")
	}
	log.Println("Connected.", len(tables), "tables found")
	return
}

func getFks(dbc *sql.DB, sourceTable *schema.Table, database schema.Database) (fks []*schema.Fk, err error) {
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
		fk := schema.NewFk(sourceTable, sourceColumn, destinationTable, destinationColumn)
		sourceColumn.Fk = fk
		fks = append(fks, fk)
	}
	return
}

func (model pgModel) GetSqlRows(query schema.RowFilter, table *schema.Table, rowLimit int) (rows *sql.Rows, err error) {
	// todo: parameterise where possible
	// todo: whitelist-sanitize unparameterizable parts
	sql := "select * from " + table.Name

	if len(query) > 0 {
		sql = sql + " where "
		clauses := make([]string, 0, len(query))
		for k, v := range query {
			clauses = append(clauses, k+" = "+v[0])
		}
		sql = sql + strings.Join(clauses, " and ")
	}

	if rowLimit > 0 {
		sql = sql + " limit " + strconv.Itoa(rowLimit)
	}

	dbc, err := getConnection(model.connectionString)
	if err != nil {
		log.Print("GetRows failed to get connection")
		return
	}
	defer dbc.Close()

	rows, err = dbc.Query(sql)
	if err != nil {
		log.Print("GetRows failed to get query")
		log.Println(sql)
		log.Println(err)
	}
	return
}

func (model pgModel) getColumns(dbc *sql.DB, table *schema.Table) (cols []*schema.Column, err error) {
	// todo: parameterise
	rows, err := dbc.Query("PRAGMA table_info('" + table.Name + "');")
	if err != nil {
		return
	}
	defer rows.Close()
	cols = []*schema.Column{}
	for rows.Next() {
		var cid int
		var name, typeName string
		var notNull, pk bool
		var defaultValue interface{}
		rows.Scan(&cid, &name, &typeName, &notNull, &defaultValue, &pk)
		thisCol := schema.Column{Name: name, Type: typeName}
		cols = append(cols, &thisCol)
	}
	return
}
