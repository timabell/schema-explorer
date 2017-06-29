package mssql

import (
	"database/sql"
	"log"
	//"github.com/denisenkom/go-mssqldb"
	"sql-data-viewer/schema"
	"strconv"
	"strings"
)

type mssqlModel struct {
	connectionString string
}

func NewMssql(connectionString string) mssqlModel {
	return mssqlModel{
		connectionString: connectionString,
	}
}

func (model mssqlModel) GetTables() (tables []schema.Table, err error) {
	dbc, err := getConnection(model.connectionString)
	if err != nil {
		return
	}
	defer dbc.Close()

	rows, err := dbc.Query("select sch.name, tbl.name from sys.tables tbl inner join sys.schemas sch on sch.schema_id = tbl.schema_id order by sch.name, tbl.name;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var schemaName string
		var name string
		rows.Scan(&schemaName, &name)
		tables = append(tables, schema.Table{Schema: schemaName, Name: name})
	}
	return tables, nil
}

func getConnection(connectionString string) (dbc *sql.DB, err error) {
	log.Println("Connecting to ", connectionString)
	dbc, err = sql.Open("mssql", connectionString)
	if err != nil {
		log.Println("connection error", err)
	}
	showVersion(dbc)
	log.Println(dbc)
	return
}

func showVersion(dbc *sql.DB) {
	rows, err := dbc.Query("select @@version")
	if err != nil {
		log.Fatal("failed to get server version.", err)
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
		log.Println("get connection failed", err)
		return
	}
	defer dbc.Close()

	// todo: why is it failing on this ? can't handle these columns for some reason.
	// err just says connection reset
	// works in SQuirreL
	//parent_sch.name parent_sch_name,
	//parent_tbl.name parent_tbl_name,
	log.Println("query setup for fks")
	rows, err := dbc.Query(`select fk.name,
	               parent_col.name parent_col_name,
	               child_sch.name child_sch_name,
	               child_tbl.name child_tbl_name,
	               child_col.name child_col_name
	           from sys.foreign_keys fk
	               inner join sys.foreign_key_columns fkcol on fkcol.constraint_object_id = fk.object_id
	               inner join sys.tables parent_tbl on parent_tbl.object_id = fk.parent_object_id
	               inner join sys.schemas parent_sch on parent_sch.schema_id = parent_tbl.schema_id
	               inner join sys.columns parent_col
	                   on parent_col.object_id = parent_tbl.object_id
	                   and parent_col.column_id = fkcol.parent_column_id
	               inner join sys.tables child_tbl on child_tbl.object_id = fk.referenced_object_id
	               inner join sys.schemas child_sch on child_sch.schema_id = child_tbl.schema_id
	               inner join sys.columns child_col
	                   on child_col.object_id = child_tbl.object_id
	                   and child_col.column_id = fkcol.referenced_column_id
	           order by fk.name`)
	log.Println("query setup done for fks")

	if err != nil {
		log.Fatal("error running query to find fks: ", err)
		return
	}
	if rows == nil {
		log.Fatal("fk rows was nil")
		return
	}
	defer rows.Close()

	allFks = schema.GlobalFkList{}
	log.Println("about to process fk rows	")
	for rows.Next() {
		log.Println("processing fk row")
		var id, seq int
		var parentSchema, parentTable, parentCol, childSchema, childTable, childCol string
		rows.Scan(&id, &seq, &parentSchema, &parentTable, &parentCol, &childSchema, &childTable, &childCol)
		table := schema.Table{Schema: parentSchema, Name: parentTable}
		col := schema.Column(parentCol)
		if allFks[table] == nil { // todo: probably need to set up map before using
			allFks[table] = schema.FkList{}
		}
		allFks[table][col] = schema.Ref{Col: schema.Column(childCol), Table: table}
	}
	log.Println("allfks done")
	return
}

func (model mssqlModel) GetRows(query schema.RowFilter, table schema.Table, rowLimit int) (rows *sql.Rows, err error) {
	sqlText := "select"

	if rowLimit > 0 {
		sqlText = sqlText + " top " + strconv.Itoa(rowLimit)
	}

	sqlText = sqlText + " * from " + table.String()

	if len(query) > 0 {
		sqlText = sqlText + " where "
		clauses := make([]string, 0, len(query))
		for k, v := range query {
			clauses = append(clauses, k+" = "+v[0])
		}
		sqlText = sqlText + strings.Join(clauses, " and ")
	}

	dbc, err := getConnection(model.connectionString)
	if dbc == nil {
		log.Println(err)
		panic("getConnection() returned nil")
	}
	defer dbc.Close()

	log.Println(sqlText)
	rows, err = dbc.Query(sqlText)
	if rows == nil {
		log.Println(err)
		panic("Query returned nil for rows")
	}
	return
}

func (model mssqlModel) Columns(table schema.Table) (columns []string, err error) {
	dbc, err := getConnection(model.connectionString)
	if err != nil {
		return nil, err
	}
	defer dbc.Close()
	log.Println("getting cols for table " + table.String())
	rows, err := dbc.Query(`select col.name
                from sys.columns col
                    inner join sys.tables tbl on tbl.object_id = col.object_id
                    inner join sys.schemas sch on sch.schema_id = tbl.schema_id
                where sch.name = ? and tbl.name = ?`, table.Schema, table.Name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var column string
		rows.Scan(&column)
		columns = append(columns, column)
	}
	return
}
