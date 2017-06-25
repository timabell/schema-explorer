package mssql

import (
	"database/sql"
	"log"
	//"github.com/denisenkom/go-mssqldb"
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
	showVersion(dbc)
	return
}

func showVersion(dbc *sql.DB) {
	rows, err := dbc.Query("select @@version")
	if err != nil {
		log.Fatal("wat ", err)
		return
	}
	defer rows.Close()
	rows.Next()
	var serverVersion string
	rows.Scan(&serverVersion)
	log.Print("Connected to " + serverVersion)
}

// todo: don't actually need an allfks method for mssql as can filter both incoming and outgoing, rework interface
func (model mssqlModel) AllFks() (allFks schema.GlobalFkList, err error) {
	// todo: share connection with other calls to this package
	dbc, err := getConnection(model.connectionString)
	if err != nil {
		// todo: show in UI
		return
	}
	defer dbc.Close()


	rows, err := dbc.Query(` select
                    --fk.name,
                    parent_sch.name + '.' + parent_tbl.name parent_tbl,
                    parent_col.name parent_col,
                    child_sch.name + '.' + child_tbl.name child_tbl,
                    child_col.name child_col
                from sys.foreign_keys fk
                    -- key members
                    inner join sys.foreign_key_columns fkcol on fkcol.constraint_object_id = fk.object_id
                    -- parent
                    inner join sys.tables parent_tbl on parent_tbl.object_id = fk.parent_object_id
                    inner join sys.schemas parent_sch on parent_sch.schema_id = parent_tbl.schema_id
                    inner join sys.columns parent_col
                        on parent_col.object_id = parent_tbl.object_id
                        and parent_col.column_id = fkcol.parent_column_id
                    -- child
                    inner join sys.tables child_tbl on child_tbl.object_id = fk.referenced_object_id
                    inner join sys.schemas child_sch on child_sch.schema_id = child_tbl.schema_id
                    inner join sys.columns child_col
                        on child_col.object_id = child_tbl.object_id
                        and child_col.column_id = fkcol.referenced_column_id
                order by fk.name`)
	if err != nil {
		log.Fatal("doh", err)
		return
	}
	defer rows.Close()

	allFks = schema.GlobalFkList{}
	for rows.Next() {
		var id, seq int
		var parentTable, parentCol, childTable, childCol string
		rows.Scan(&id, &seq, &parentTable, &parentCol, &childTable, &childCol)
		table := schema.TableName(parentTable)
		col := schema.ColumnName(parentCol)
		//if allFks[table] { // todo: probably need to set up map before using
		allFks[table][col] = schema.Ref{Col: schema.ColumnName(childCol), Table: table}
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

