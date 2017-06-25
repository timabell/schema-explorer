package mssql

import (
	"database/sql"
	"log"
	//"github.com/denisenkom/go-mssqldb"
	"fmt"
	"strings"
	"strconv"
	"sql-data-viewer/schema"
)

type mssqlModel struct{
	connectionString string
}

func NewMssql(connectionString string) mssqlModel {
	return mssqlModel{
		connectionString: connectionString,
	}
}

func (model mssqlModel) GetTables() (tables []schema.TableName, err error) {
	dbc, err := getConnection(model.connectionString)
	if err != nil {
		return
	}
	defer dbc.Close()

	rows, err := dbc.Query("select sch.name + '.' + tbl.name from sys.tables tbl inner join sys.schemas sch on sch.schema_id = tbl.schema_id order by sch.name, tbl.name;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, schema.TableName(name))
	}
	return tables, nil
}

func getConnection(connectionString string) (dbc *sql.DB, err error) {
	dbc, err = sql.Open("mssql", connectionString)
	if err != nil {
		log.Println("connection error", err)
	}
	return
}

func (model mssqlModel) AllFks() (allFks schema.GlobalFkList, err error) {
	tables, err := model.GetTables()
	if err != nil {
		fmt.Println("error getting table list while building global fk list", err)
		return
	}
	allFks = schema.GlobalFkList{}

	// todo: share connection with GetTables()
	dbc, err := getConnection(model.connectionString)
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

func fks(dbc *sql.DB, table schema.TableName) (fks schema.FkList, err error) {
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
		thisRef := schema.Ref{Col: schema.ColumnName(to), Table: schema.TableName(parentTable)}
		fks[schema.ColumnName(from)] = thisRef
	}
	return
}

func (model mssqlModel) GetRows(query schema.RowFilter, table schema.TableName, rowLimit int) (rows *sql.Rows, err error) {
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

	dbc, err := getConnection(model.connectionString)
	if err != nil {
		// todo: show in UI
		return
	}
	defer dbc.Close()

	rows, err = dbc.Query(sql)
	return
}

