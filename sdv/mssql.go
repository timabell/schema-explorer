package sdv

import (
	"database/sql"
	"log"
	//"github.com/denisenkom/go-mssqldb"
)

type mssqlModel struct{
	connectionString string
}

func NewMssql(connectionString string) mssqlModel {
	return mssqlModel{
		connectionString: connectionString,
	}
}

func (model mssqlModel) GetTables() (tables []TableName, err error) {
	dbc, err := sql.Open("mssql", model.connectionString)
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
