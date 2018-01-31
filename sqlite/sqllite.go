package sqlite

// Sqlite doesn't support schema so table.schema is ignored throughout

import (
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"database/sql"
	"fmt"
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

func (model sqliteModel) ReadSchema() (database schema.Database, err error) {
	dbc, err := getConnection(model.path)
	if err != nil {
		return
	}
	defer dbc.Close()

	database = schema.Database{Supports: schema.SupportedFeatures{Schema: false}}

	database.Tables, err = model.getTables()
	if err != nil {
		return
	}

	for tableIndex, table := range database.Tables {
		var cols []*schema.Column
		cols, err = model.getColumns(table)
		if err != nil {
			return
		}
		database.Tables[tableIndex].Columns = append(table.Columns, cols...)
	}

	database.Fks, err = model.allFks()

	for _, fk := range database.Fks{
		source := database.FindTable(fk.SourceTable)
		if source == nil {
			err = fmt.Errorf("failed to find source table for fk %s", fk)
		}
		source.Fks = append(source.Fks, fk)
		log.Printf("%#v", source.Fks)
		destination := database.FindTable(fk.DestinationTable)
		if destination == nil {
			err = fmt.Errorf("failed to find destination table for fk %s", fk)
		}
		destination.InboundFks = append(destination.InboundFks, fk)
		log.Printf("%#v", destination.InboundFks)
	}
	return
}

func (model sqliteModel) getTables() (tables []*schema.Table, err error) {
	dbc, err := getConnection(model.path)
	if err != nil {
		return
	}
	defer dbc.Close()

	rows, err := dbc.Query("SELECT name FROM sqlite_master WHERE type='table' AND name not like 'sqlite_%';")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, &schema.Table{Name: name})
	}
	return tables, nil
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
	tables, err := model.getTables()
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

func (model sqliteModel) allFks() (allFks []*schema.Fk, err error) {
	tables, err := model.getTables()
	if err != nil {
		fmt.Println("error getting table list while building global fk list", err)
		return
	}

	allFks = []*schema.Fk{}

	// todo: share connection with getTables()
	dbc, err := getConnection(model.path)
	if err != nil {
		// todo: show in UI
		return
	}
	defer dbc.Close()

	for _, table := range tables {
		var tableFks []*schema.Fk
		tableFks, err = fks(dbc, table)
		if err != nil {
			// todo: show in UI
			fmt.Println("error getting fks for table "+table.Name, err)
			return
		}
		allFks = append(allFks, tableFks...)
	}
	return
}

func fks(dbc *sql.DB, table *schema.Table) (fks []*schema.Fk, err error) {
	rows, err := dbc.Query("PRAGMA foreign_key_list('" + table.Name + "');")
	if err != nil {
		return
	}
	defer rows.Close()
	fks = []*schema.Fk{}
	for rows.Next() {
		var id, seq int
		var parentTable, from, to, onUpdate, onDelete, match string
		rows.Scan(&id, &seq, &parentTable, &from, &to, &onUpdate, &onDelete, &match)
		sourceColumn := schema.Column{Name: from}
		destinationColumn := schema.Column{Name: to}
		fk := schema.NewFk(table, &sourceColumn, &schema.Table{Name: parentTable}, &destinationColumn)
		fks = append(fks, fk)
	}
	return
}

func (model sqliteModel) GetSqlRows(query schema.RowFilter, table *schema.Table, rowLimit int) (rows *sql.Rows, err error) {
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

	dbc, err := getConnection(model.path)
	if err != nil {
		log.Println(err)
		panic("GetRows to get connection")
		// todo: show in UI
		return
	}
	defer dbc.Close()

	rows, err = dbc.Query(sql)
	if err != nil {
		log.Println(sql)
		log.Println(err)
		panic("GetRows failed to get query")
		// todo: show in UI
		return
	}
	return
}

func (model sqliteModel) getColumns(table *schema.Table) (cols []*schema.Column, err error) {
	dbc, err := getConnection(model.path)
	if err != nil {
		log.Println(err)
		panic("getColumns to get connection")
		// todo: show in UI
		return
	}
	defer dbc.Close()
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
