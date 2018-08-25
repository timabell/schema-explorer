package mssql

import (
	"bitbucket.org/timabell/sql-data-viewer/params"
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"database/sql"
	_ "github.com/denisenkom/go-mssqldb"
	"log"
	"strconv"
	"strings"
	"os"
)

type mssqlModel struct {
	connectionString string
}

type mssqlOpts struct {
	// todo: break down into host, port etc
	Db *string `long:"db" description:"Mssql connection string. # see https://godoc.org/github.com/lib/pq for connection-string options"`
}

var opt = &mssqlOpts{}

func init() {
	// https://github.com/jessevdk/go-flags/blob/master/group_test.go#L33
	reader.RegisterReader("mssql", opt, NewMssql)
}

func NewMssql() reader.DbReader {
	if opt.Db == nil {
		log.Printf("Error: connection string (db) is required")
		reader.ArgParser.WriteHelp(os.Stdout)
		os.Exit(1)
	}
	log.Println("Connecting to mssql db")
	return mssqlModel{
		connectionString: *opt.Db,
	}
}

func (model mssqlModel) ReadSchema() (database *schema.Database, err error) {
	dbc, err := getConnection(model.connectionString)
	if err != nil {
		return
	}
	defer dbc.Close()

	database = &schema.Database{
		Supports: schema.SupportedFeatures{
			Schema:       true,
			Descriptions: true,
			FkNames:      true,
		},
		DefaultSchemaName: "dbo",
	}

	database.Tables, err = model.getTables(dbc)
	if err != nil {
		return
	}

	// columns
	for tableIndex, table := range database.Tables {
		var cols []*schema.Column
		cols, err = model.getColumns(dbc, table)
		if err != nil {
			return
		}
		database.Tables[tableIndex].Columns = append(table.Columns, cols...)
	}

	database.Fks, err = model.allFks(dbc, database)
	if err != nil {
		return
	}

	model.getPks(dbc, database)

	addDescriptions(dbc, database)

	//log.Print(database.DebugString())
	return
}

func addDescriptions(dbc *sql.DB, database *schema.Database) error {
	rows, err := dbc.Query(`
		select
			sch.name [schema],
			tbl.name [table],
			col.name [column],
			ep.value [description]
			from sys.extended_properties ep
			inner join sys.objects tbl on tbl.object_id = ep.major_id
			inner join sys.schemas sch on sch.schema_id = tbl.schema_id
			left outer join sys.columns col on col.object_id = ep.major_id and col.column_id = ep.minor_id
			where ep.name = 'MS_Description'
		order by tbl.name, ep.minor_id`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var schemaName, tableName, colName, description *string
		rows.Scan(&schemaName, &tableName, &colName, &description)
		// todo: support non-dbo schema for descriptions
		table := database.FindTable(&schema.Table{Schema: *schemaName, Name: *tableName})
		if table == nil {
			// ignore unknown things. could be for views that we don't currently support
			continue
		}
		if colName == nil {
			table.Description = *description
			continue
		}
		_, col := table.FindColumn(*colName)
		col.Description = *description
	}
	return nil
}

func (model mssqlModel) getTables(dbc *sql.DB) (tables []*schema.Table, err error) {

	rows, err := dbc.Query("select sch.name, tbl.name from sys.tables tbl inner join sys.schemas sch on sch.schema_id = tbl.schema_id order by sch.name, tbl.name;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var schemaName string
		var name string
		rows.Scan(&schemaName, &name)
		tables = append(tables, &schema.Table{Schema: schemaName, Name: name})
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

func (model mssqlModel) getRowCount(table *schema.Table) (rowCount int, err error) {
	// todo: parameterise where possible
	// todo: whitelist-sanitize unparameterizable parts
	sql := "select count(*) from " + table.String()

	dbc, err := getConnection(model.connectionString)
	if dbc == nil {
		log.Println(err)
		panic("getConnection() returned nil")
	}
	defer dbc.Close()
	rows, err := dbc.Query(sql)
	if err != nil {
		log.Println(sql)
		return 0, err
	}
	defer rows.Close()
	rows.Next()
	var count int
	rows.Scan(&count)
	return count, nil
}

func getConnection(connectionString string) (dbc *sql.DB, err error) {
	dbc, err = sql.Open("mssql", connectionString)
	if err != nil {
		log.Println("connection error", err)
	}
	return
}

func (model mssqlModel) CheckConnection() (err error) {
	dbc, err := getConnection(model.connectionString)
	if dbc == nil {
		log.Println(err)
		panic("getConnection() returned nil")
	}
	defer dbc.Close()
	showVersion(dbc)
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
	serverVersion = strings.Replace(serverVersion, "\n", " ", -1)
	serverVersion = strings.Replace(serverVersion, "\t", " ", -1)
	log.Print("Successfully connected to MSSQL. @@version: " + serverVersion)
}

func (model mssqlModel) allFks(dbc *sql.DB, database *schema.Database) (allFks []*schema.Fk, err error) {
	rows, err := dbc.Query(`
		select fk.name,
			parent_sch.name parent_sch_name,
			parent_tbl.name parent_tbl_name,
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

	if err != nil {
		log.Fatal("error running query to find fks: ", err)
		return
	}
	if rows == nil {
		log.Fatal("fk rows was nil")
		return
	}
	defer rows.Close()

	allFks = []*schema.Fk{}
	for rows.Next() {
		var name, sourceSchema, sourceTableName, sourceColumnName, destinationSchema, destinationTableName, destinationColumnName string
		rows.Scan(&name, &sourceSchema, &sourceTableName, &sourceColumnName, &destinationSchema, &destinationTableName, &destinationColumnName)
		sourceTable := database.FindTable(&schema.Table{Schema: sourceSchema, Name: sourceTableName})
		_, sourceColumn := sourceTable.FindColumn(sourceColumnName)
		destinationTable := database.FindTable(&schema.Table{Schema: destinationSchema, Name: destinationTableName})
		_, destinationColumn := destinationTable.FindColumn(destinationColumnName)
		// see if we are adding columns to an existing fk
		var fk *schema.Fk
		for _, existingFk := range allFks {
			if existingFk.Name == name {
				existingFk.SourceColumns = append(existingFk.SourceColumns, sourceColumn)
				existingFk.DestinationColumns = append(existingFk.DestinationColumns, destinationColumn)
				fk = existingFk
				break
			}
		}
		if fk == nil {
			fk = schema.NewFk(name, sourceTable, sourceColumn, destinationTable, destinationColumn)
			allFks = append(allFks, fk)
			sourceTable.Fks = append(sourceTable.Fks, fk)
			destinationTable.InboundFks = append(destinationTable.InboundFks, fk)
		}
		sourceColumn.Fks = append(sourceColumn.Fks, fk)
		//log.Print(fk)
	}
	return
}

func (model mssqlModel) GetSqlRows(table *schema.Table, params *params.TableParams) (rows *sql.Rows, err error) {
	// todo: sql parameters instead of string concatenation
	sql := "select"

	if params.RowLimit > 0 {
		sql = sql + " top " + strconv.Itoa(params.RowLimit)
	}

	sql = sql + " * from " + table.String()

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

	dbc, err := getConnection(model.connectionString)
	if dbc == nil {
		log.Println(err)
		panic("getConnection() returned nil")
	}
	defer dbc.Close()

	rows, err = dbc.Query(sql, values...)
	if rows == nil {
		log.Println(sql)
		log.Println(err)
		panic("Query returned nil for rows")
	}
	return
}

func (model mssqlModel) getColumns(dbc *sql.DB, table *schema.Table) (cols []*schema.Column, err error) {
	// todo: parameterise
	sqlText := `select c.name, type_name(c.system_type_id), is_nullable from sys.columns c
	inner join sys.tables t on t.object_id = c.object_id
	inner join sys.schemas s on s.schema_id = t.schema_id
	where s.name = '` + table.Schema + `' and t.name = '` + table.Name + `'
order by c.column_id`

	rows, err := dbc.Query(sqlText)
	defer rows.Close()
	cols = []*schema.Column{}
	colIndex := 0
	for rows.Next() {
		var name, typeName string
		var nullable bool
		rows.Scan(&name, &typeName, &nullable)
		thisCol := schema.Column{Position: colIndex, Name: name, Type: typeName, Nullable: nullable}
		cols = append(cols, &thisCol)
		colIndex++
	}
	return
}

func (model mssqlModel) getPks(dbc *sql.DB, database *schema.Database) {
	rows, err := dbc.Query(`
		select
			ix.name index_name,
			s.name schema_name,
			t.name table_name,
			col.name colname
		from sys.indexes ix
		inner join sys.index_columns ic on ic.object_id = ix.object_id and ic.index_id = ix.index_id
		inner join sys.tables t on t.object_id = ix.object_id
		inner join sys.columns col on col.object_id = ix.object_id and col.column_id = ic.column_id
		inner join sys.schemas s on s.schema_id = t.schema_id
		where ix.is_primary_key = 1`)

	if err != nil {
		log.Fatal("error running query to find pks: ", err)
		return
	}
	if rows == nil {
		log.Fatal("pk rows was nil")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var pkName, schemaName, tableName, columnName string
		rows.Scan(&pkName, &schemaName, &tableName, &columnName)
		table := database.FindTable(&schema.Table{Schema: schemaName, Name: tableName})
		_, col := table.FindColumn(columnName)
		col.IsInPrimaryKey = true
		if table.Pk == nil {
			table.Pk = &schema.Pk{Name: pkName, Columns: schema.ColumnList{col}}
		} else {
			table.Pk.Columns = append(table.Pk.Columns, col)
		}
	}
}
