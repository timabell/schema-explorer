package sqlite

import (
	"database/sql"
	"log"
	"fmt"
	"strings"
	"strconv"
	"sql-data-viewer/schema"
)

type sqliteModel struct{
	path string
}

func NewSqlite(path string) sqliteModel {
	return sqliteModel{
		path: path,
	}
}

func (model sqliteModel) GetTables() (tables []schema.Table, err error) {
	dbc, err := getConnection(model.path)
	if err != nil {
		return
	}
	defer dbc.Close()

	rows, err := dbc.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, schema.Table(name))
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

func (model sqliteModel) AllFks() (allFks schema.GlobalFkList, err error) {
	tables, err := model.GetTables()
	if err != nil {
		fmt.Println("error getting table list while building global fk list", err)
		return
	}
	allFks = schema.GlobalFkList{}

	// todo: share connection with GetTables()
	dbc, err := getConnection(model.path)
	if err != nil {
		// todo: show in UI
		return
	}
	defer dbc.Close()

	for _, table := range tables {
		allFks[table], err = fks(dbc, table)
		if err != nil {
			// todo: show in UI
			fmt.Println("error getting fks for table " + table, err)
			return
		}
	}
	return
}

func fks(dbc *sql.DB, table schema.Table) (fks schema.FkList, err error) {
	rows, err := dbc.Query("PRAGMA foreign_key_list('" + string(table) + "');")
	if err != nil {
		return
	}
	defer rows.Close()
	fks = schema.FkList{}
	for rows.Next() {
		var id, seq int
		var parentTable, from, to, onUpdate, onDelete, match string
		rows.Scan(&id, &seq, &parentTable, &from, &to, &onUpdate, &onDelete, &match)
		thisRef := schema.Ref{Col: schema.Column(to), Table: schema.Table(parentTable)}
		fks[schema.Column(from)] = thisRef
	}
	return
}

func (model sqliteModel) GetRows(query schema.RowFilter, table schema.Table, rowLimit int) (rows *sql.Rows, err error) {
	sql := "select * from " + string(table)

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

	log.Println(sql)

	dbc, err := getConnection(model.path)
	if err != nil {
		// todo: show in UI
		return
	}
	defer dbc.Close()

	rows, err = dbc.Query(sql)
	return
}

func (model sqliteModel) Columns(table schema.Table) (columns []string, err error) {
	panic("not implemented")
}
