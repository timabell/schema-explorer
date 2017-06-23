package sdv

import (
"database/sql"
"log"
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
	dbc, err := sql.Open("sqlite3", model.path)
	if err != nil {
		log.Println("connection error", err)
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

