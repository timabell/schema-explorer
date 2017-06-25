package sdv

import (
	"database/sql"
	"log"
	"fmt"
	"strings"
	"strconv"
)

type sqliteModel struct{
	path string
}

func NewSqlite(path string) sqliteModel {
	return sqliteModel{
		path: path,
	}
}

func (model sqliteModel) GetTables() (tables []TableName, err error) {
	dbc, err := getConnection(model.path)
	if err != nil {
		// todo: show in UI
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
		tables = append(tables, TableName(name))
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

func (model sqliteModel) AllFks() (allFks GlobalFkList) {
	tables, err := model.GetTables()
	if err != nil {
		fmt.Println("error getting table list while building global fk list", err)
		return
	}
	allFks = GlobalFkList{}

	// todo: share connection with GetTables()
	dbc, err := getConnection(model.path)
	if err != nil {
		// todo: show in UI
		return
	}
	defer dbc.Close()

	for _, table := range tables {
		allFks[table] = fks(dbc, table)
	}
	return
}

func fks(dbc *sql.DB, table TableName) (fks FkList) {
	rows, err := dbc.Query("PRAGMA foreign_key_list('" + string(table) + "');")
	if err != nil {
		log.Println("select error", err)
		return
	}
	defer rows.Close()
	fks = FkList{}
	for rows.Next() {
		var id, seq int
		var parentTable, from, to, onUpdate, onDelete, match string
		rows.Scan(&id, &seq, &parentTable, &from, &to, &onUpdate, &onDelete, &match)
		thisRef := Ref{Col: ColumnName(to), Table: TableName(parentTable)}
		fks[ColumnName(from)] = thisRef
	}
	return
}

func (model sqliteModel) GetRows(query RowFilter, table TableName, rowLimit int) (rows *sql.Rows, err error) {
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
	if err != nil {
		// todo: show in UI
		log.Println("select error", err)
	}
	return
}
