// +build !skip_pg

package pg

import (
	"github.com/timabell/schema-explorer/driver_interface"
	"github.com/timabell/schema-explorer/drivers"
	"github.com/timabell/schema-explorer/params"
	"github.com/timabell/schema-explorer/reader"
	"github.com/timabell/schema-explorer/schema"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strconv"
	"strings"
)

var driverOpts = drivers.DriverOpts{
	"host":              drivers.DriverOpt{Description: "Postgres host", Value: &opts.Host},
	"port":              drivers.DriverOpt{Description: "Postgres port", Value: &opts.Port},
	"database":          drivers.DriverOpt{Description: "Postgres database name", Value: &opts.Database},
	"user":              drivers.DriverOpt{Description: "Postgres username", Value: &opts.User},
	"password":          drivers.DriverOpt{Description: "Postgres password", Value: &opts.Password},
	"ssl-mode":          drivers.DriverOpt{Description: "Postgres ssl mode. Set this to 'disable' if you are connecting to a server that doesn't have ssl enabled.'", Value: &opts.SslMode},
	"connection-string": drivers.DriverOpt{Description: "Postgres connection string. Use this instead of host, port etc for advanced driver options. See https://godoc.org/github.com/lib/pq for connection-string options.", Value: &opts.ConnectionString},
}

type pgModel struct {
	connected bool // todo: technically it's a connection string per db so we could end up in multiple states, ignore for now
}

type pgOpts struct {
	Host             string
	Port             string
	Database         string
	User             string
	Password         string
	SslMode          string
	ConnectionString string
}

func (opts pgOpts) validate() error {
	if opts.hasAnyDetails() && opts.ConnectionString != "" {
		return errors.New("Specify either a connection string or host etc, not both.")
	}
	return nil
}

func (opts pgOpts) hasAnyDetails() bool {
	return opts.Host != "" ||
		opts.Port != "" ||
		opts.Database != "" ||
		opts.User != "" ||
		opts.Password != ""
}

var opts = &pgOpts{}

func init() {
	reader.RegisterReader(&drivers.Driver{Name: "pg", Options: driverOpts, CreateReader: newPg, FullName: "Postgres"})
}

func newPg() driver_interface.DbReader {
	err := opts.validate()
	if err != nil {
		log.Printf("Pg args error: %s", err)
		//options.ArgParser.WriteHelp(os.Stdout)
		os.Exit(1)
	}
	log.Println("Connecting to pg db")
	return pgModel{connected: false}
}

// optionally override db name with param
func buildConnectionString(databaseName string) string {
	if opts.ConnectionString != "" {
		return opts.ConnectionString
	}

	optList := make(map[string]string)
	if opts.Host != "" {
		optList["host"] = opts.Host
	}
	if opts.Port != "" {
		optList["port"] = opts.Port
	}
	if databaseName != "" {
		optList["dbname"] = databaseName
	} else if opts.Database != "" {
		optList["dbname"] = opts.Database
	}
	if opts.User != "" {
		optList["user"] = opts.User
	}
	if opts.Password != "" {
		optList["password"] = opts.Password
	}
	if opts.SslMode != "" {
		optList["sslmode"] = opts.SslMode
	}
	pairs := []string{}
	for key, value := range optList {
		pairs = append(pairs, fmt.Sprintf("%s='%s'", key, strings.Replace(value, "'", "\\'", -1)))
	}
	return strings.Join(pairs, " ")
}

func (model pgModel) ReadSchema(databaseName string) (database *schema.Database, err error) {
	dbc, err := getConnection(buildConnectionString(databaseName))
	if err != nil {
		return
	}
	defer dbc.Close()

	database = &schema.Database{
		Supports: schema.SupportedFeatures{
			Schema:               true,
			Descriptions:         false,
			FkNames:              true,
			PagingWithoutSorting: true,
		},
		DefaultSchemaName: "public",
		Name:              databaseName,
	}

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

	// fks and other constraints
	err = readConstraints(dbc, database)
	if err != nil {
		return
	}

	// indexes
	err = readIndexes(dbc, database)
	if err != nil {
		return
	}

	//log.Print(database.DebugString())
	return
}

func (model pgModel) CanSwitchDatabase() bool {
	return opts.ConnectionString == "" && opts.Database == ""
}

func (model pgModel) GetConfiguredDatabaseName() string {
	return opts.Database
}

func (model pgModel) ListDatabases() (databaseList []string, err error) {
	sql := "select datname from pg_database where datistemplate = false order by datname;"

	dbc, err := getConnection(buildConnectionString(""))
	if dbc == nil {
		log.Println(err)
		panic("getConnection() returned nil")
	}
	defer dbc.Close()
	rows, err := dbc.Query(sql)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		rows.Scan(&name)
		databaseList = append(databaseList, name)
	}
	return
}

func (model pgModel) DatabaseSelected() bool {
	return opts.Database != "" || opts.ConnectionString != ""
}

func (model pgModel) UpdateRowCounts(database *schema.Database) (err error) {
	for _, table := range database.Tables {
		rowCount, err := model.getRowCount(database.Name, table)
		if err != nil {
			// todo: aggregate errors to return
			log.Printf("Failed to get row count for %s, %s", table, err)
			rowCount = -1
		}
		table.RowCount = &rowCount
	}
	return err
}

func (model pgModel) getRowCount(databaseName string, table *schema.Table) (rowCount int, err error) {
	// todo: parameterise where possible
	// todo: whitelist-sanitize unparameterizable parts
	sql := "select count(*) from \"" + table.Schema + "\".\"" + table.Name + "\""

	dbc, err := getConnection(buildConnectionString(databaseName))
	if dbc == nil {
		log.Println(err)
		panic("getConnection() returned nil")
	}
	defer dbc.Close()
	rows, err := dbc.Query(sql)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	rows.Next()
	var count int
	rows.Scan(&count)
	return count, nil
}

func (model pgModel) getTables(dbc *sql.DB) (tables []*schema.Table, err error) {
	rows, err := dbc.Query("select schemaname, tablename from pg_catalog.pg_tables where schemaname not in ('pg_catalog','information_schema') order by schemaname, tablename")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name, schemaName string
		rows.Scan(&schemaName, &name)
		tables = append(tables, &schema.Table{Schema: schemaName, Name: name, Pk: &schema.Pk{}})
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

func (model pgModel) CheckConnection(databaseName string) (err error) {
	dbc, err := getConnection(buildConnectionString(databaseName))
	if dbc == nil {
		log.Println(err)
		panic("getConnection() returned nil")
	}
	defer dbc.Close()
	err = dbc.Ping()
	if err != nil {
		return
	}
	model.connected = true
	log.Println("Postgres connected.")
	return
}

func (model pgModel) Connected() bool {
	return model.connected
}

func readConstraints(dbc *sql.DB, database *schema.Database) (err error) {
	// null-proof unnest: https://stackoverflow.com/a/49736694
	sql := fmt.Sprintf(`
		select
			con.oid, ns.nspname, con.conname, con.contype,
			tns.nspname, tbl.relname, col.attname column_name,
			fns.nspname foreign_namespace_name, ftbl.relname foreign_table_name, fcol.attname foreign_column_name
		from
			(
				select pgc.oid, pgc.connamespace, pgc.conrelid, pgc.confrelid, pgc.contype, pgc.conname,
				       unnest(case when pgc.conkey <> '{}' then pgc.conkey else '{null}' end) as conkey,
				       unnest(case when pgc.confkey <> '{}' then pgc.confkey else '{null}' end) as confkey
				from pg_constraint pgc
			) as con
			inner join pg_namespace ns on con.connamespace = ns.oid
			inner join pg_class tbl on tbl.oid = con.conrelid
			inner join pg_namespace tns on tbl.relnamespace = tns.oid
			inner join pg_attribute col on col.attrelid = tbl.oid and col.attnum = con.conkey
			left outer join pg_class ftbl on ftbl.oid = con.confrelid
			left outer join pg_namespace fns on ftbl.relnamespace = fns.oid
			left outer join pg_attribute fcol on fcol.attrelid = ftbl.oid and fcol.attnum = con.confkey;`)

	rows, err := dbc.Query(sql)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var oid, conType, namespace, name,
			sourceNamespace, sourceTableName, sourceColumnName,
			destinationNamespace, destinationTableName, destinationColumnName string
		rows.Scan(&oid, &namespace, &name, &conType,
			&sourceNamespace, &sourceTableName, &sourceColumnName,
			&destinationNamespace, &destinationTableName, &destinationColumnName)
		tableToFind := &schema.Table{Schema: sourceNamespace, Name: sourceTableName}
		sourceTable := database.FindTable(tableToFind)
		if sourceTable == nil {
			err = errors.New(fmt.Sprintf("Table %s not found, source of constraint %s", tableToFind.String(), name))
			return
		}
		_, sourceColumn := sourceTable.FindColumn(sourceColumnName)
		if sourceColumn == nil {
			err = errors.New(fmt.Sprintf("Column %s not found on table %s, source of constraint %s", sourceColumnName, tableToFind.String(), name))
			return
		}
		switch conType {
		case "f": // f = foreign key
			destinationTable := database.FindTable(&schema.Table{Schema: destinationNamespace, Name: destinationTableName})
			if destinationTable == nil {
				//log.Print(database.DebugString())
				panic(fmt.Sprintf("couldn't find table %s in database object while hooking up fks", destinationTableName))
			}
			_, destinationColumn := destinationTable.FindColumn(destinationColumnName)
			// see if we are adding columns to an existing fk

			var fk *schema.Fk
			for _, existingFk := range database.Fks {
				if existingFk.Name == name {
					existingFk.SourceColumns = append(existingFk.SourceColumns, sourceColumn)
					existingFk.DestinationColumns = append(existingFk.DestinationColumns, destinationColumn)
					fk = existingFk
					break
				}
			}
			if fk == nil { // then this is a never-before-seen fk
				fk = schema.NewFk(name, sourceTable, sourceColumn, destinationTable, destinationColumn)
				database.Fks = append(database.Fks, fk)
				sourceTable.Fks = append(sourceTable.Fks, fk)
				sourceColumn.Fks = append(sourceColumn.Fks, fk)
				destinationTable.InboundFks = append(destinationTable.InboundFks, fk)
				destinationColumn.InboundFks = append(destinationColumn.InboundFks, fk)
			}
			//log.Printf("fk: %+v - oid %+v", fk, oid)
		case "p": // primary key
			//log.Printf("pk: %s.%s", sourceTable, sourceColumn)
			sourceTable.Pk.Columns = append(sourceTable.Pk.Columns, sourceColumn)
			sourceColumn.IsInPrimaryKey = true
		case "c": // todo: check constraint
		case "u": // todo: unique constraint
		case "t": // todo: constraint
		case "x": // todo: exclusion constraint
		default:
			log.Printf("?? %s", conType)
		}
	}
	return
}

func readIndexes(dbc *sql.DB, database *schema.Database) (err error) {
	sql := `
		select
			oc.relname,
			tns.nspname, tbl.relname table_relname,
			col.attname colname,
			ix.indisunique,
			ix.indisclustered
		from (
			select *, unnest(indkey) colnum from pg_index
		) ix
		left outer join pg_class oc on oc.oid = ix.indexrelid
		left outer join pg_class tbl on tbl.oid = ix.indrelid
		left outer join pg_namespace tns on tbl.relnamespace = tns.oid
		left outer join pg_attribute col on col.attrelid = ix.indrelid and col.attnum = ix.colnum
		where tns.nspname not like 'pg_%'
			and not ix.indisprimary;
	`

	//log.Println(sql)
	rows, err := dbc.Query(sql)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var name, tableNamespace, tableName, colName string
		var isUnique, isClustered bool
		rows.Scan(&name, &tableNamespace, &tableName, &colName, &isUnique, &isClustered)
		tableToFind := &schema.Table{Schema: tableNamespace, Name: tableName}
		table := database.FindTable(tableToFind)
		if table == nil {
			err = errors.New(fmt.Sprintf("Table %s not found, owner of index %s", tableToFind.String(), name))
			return
		}
		var index *schema.Index
		for _, existingIndex := range table.Indexes {
			if existingIndex.Name == name {
				index = existingIndex
				break
			}
		}
		if index == nil {
			index = &schema.Index{
				Name:        name,
				Columns:     []*schema.Column{},
				IsUnique:    isUnique,
				Table:       table,
				IsClustered: isClustered,
			}
			database.Indexes = append(database.Indexes, index)
			table.Indexes = append(table.Indexes, index)
		}
		if colName != "" { // more complex indexes don't link back to their columns. See pg_index.indkey https://www.postgresql.org/docs/current/static/catalog-pg-index.html
			_, col := table.FindColumn(colName)
			if col == nil {
				err = errors.New(fmt.Sprintf("Column %s in table %s not found, for index %s", colName, tableToFind.String(), name))
				return
			}
			index.Columns = append(index.Columns, col)
			col.Indexes = append(col.Indexes, index)
		}
		//log.Printf(index.String())
	}
	return
}

func (model pgModel) GetSqlRows(databaseName string, table *schema.Table, params *params.TableParams, peekFinder *driver_interface.PeekLookup) (rows *sql.Rows, err error) {
	dbc, err := getConnection(buildConnectionString(databaseName))
	if err != nil {
		log.Print("GetRows failed to get connection")
		return
	}
	defer dbc.Close()

	sql, values := buildQuery(table, params, peekFinder)
	rows, err = dbc.Query(sql, values...)
	if err != nil {
		log.Print("GetRows failed to get query")
		log.Println(sql)
		log.Println(err)
	}
	return
}

func (model pgModel) GetRowCount(databaseName string, table *schema.Table, params *params.TableParams) (rowCount int, err error) {
	dbc, err := getConnection(buildConnectionString(databaseName))
	if err != nil {
		log.Print("GetRows failed to get connection")
		return
	}
	defer dbc.Close()

	sql, values := buildQuery(table, params, &driver_interface.PeekLookup{})
	sql = "select count(*) from (" + sql + ") as x"
	rows, err := dbc.Query(sql, values...)
	if err != nil {
		log.Print("GetRowCount failed to get query")
		log.Println(sql)
		log.Println(err)
		return
	}
	if !rows.Next() {
		err = errors.New("GetRowCount query returned no rows")
		return
	}
	rows.Scan(&rowCount)
	return
}

func (model pgModel) GetAnalysis(databaseName string, table *schema.Table) (analysis []schema.ColumnAnalysis, err error) {
	// todo, might be good to stream this all the way to the http response
	dbc, err := getConnection(buildConnectionString(databaseName))
	if err != nil {
		log.Print("GetAnalysis failed to get connection")
		return
	}
	defer dbc.Close()

	analysis = []schema.ColumnAnalysis{}
	for _, col := range table.Columns {
		sql := "select \"" + col.Name + "\", count(*) qty from \"" + table.Schema + "\".\"" + table.Name + "\" group by \"" + col.Name + "\" order by count(*) desc, \"" + col.Name + "\" limit 100;"
		rows, err := dbc.Query(sql)
		if err != nil {
			log.Print("GetAnalysis failed to get query")
			log.Println(sql)
			log.Println(err)
			return nil, err
		}
		var valueInfos []schema.ValueInfo
		for rows.Next() {
			var value interface{}
			var quantity int
			rows.Scan(&value, &quantity)
			valueInfos = append(valueInfos, schema.ValueInfo{
				Value:    value,
				Quantity: quantity,
			})
		}
		analysis = append(analysis, schema.ColumnAnalysis{
			Column:      col,
			ValueCounts: valueInfos,
		})
	}
	return
}

func buildQuery(table *schema.Table, params *params.TableParams, peekFinder *driver_interface.PeekLookup) (sql string, values []interface{}) {
	sql = "select t.*"

	// peek cols
	for fkIndex, fk := range peekFinder.Fks {
		for _, peekCol := range fk.DestinationTable.PeekColumns {
			sql = sql + fmt.Sprintf(", fk%d.\"%s\" fk%d_%s", fkIndex, peekCol, fkIndex, peekCol)
		}
	}

	// inbound fk counts
	for inboundFkIndex, inboundFk := range table.InboundFks {
		onPredicates := []string{}
		for ix, sourceCol := range inboundFk.SourceColumns {
			onPredicates = append(onPredicates, fmt.Sprintf("ifk%d.\"%s\" = t.\"%s\"", inboundFkIndex, sourceCol.Name, inboundFk.DestinationColumns[ix].Name))
		}
		onString := strings.Join(onPredicates, " and ")
		sql = sql + fmt.Sprintf(", (select count(*) from \"%s\".\"%s\" ifk%d where %s) ifk%d_count", inboundFk.SourceTable.Schema, inboundFk.SourceTable.Name, inboundFkIndex, onString, inboundFkIndex)
	}

	sql = sql + " from \"" + table.Schema + "\".\"" + table.Name + "\" t"

	// peek tables
	for fkIndex, fk := range peekFinder.Fks {
		sql = sql + fmt.Sprintf(" left outer join \"%s\".\"%s\" fk%d on ", fk.DestinationTable.Schema, fk.DestinationTable.Name, fkIndex)
		onPredicates := []string{}
		for ix, sourceCol := range fk.SourceColumns {
			onPredicates = append(onPredicates, fmt.Sprintf("t.\"%s\" = fk%d.\"%s\"", sourceCol, fkIndex, fk.DestinationColumns[ix]))
		}
		onString := strings.Join(onPredicates, " and ")
		sql = sql + onString
	}

	query := params.Filter
	if len(query) > 0 {
		sql = sql + " where "
		clauses := make([]string, 0, len(query))
		values = make([]interface{}, 0, len(query))
		var index = 1
		for _, v := range query {
			col := v.Field
			clauses = append(clauses, "t.\""+col.Name+"\" = $"+strconv.Itoa(index))
			index = index + 1
			values = append(values, v.Values[0]) // todo: maybe support multiple values
		}
		sql = sql + strings.Join(clauses, " and ")
	}

	if len(params.Sort) > 0 {
		var sortParts []string
		for _, sortCol := range params.Sort {
			sortString := "\"" + sortCol.Column.Name + "\""
			if sortCol.Descending {
				sortString = sortString + " desc"
			}
			sortParts = append(sortParts, sortString)
		}
		sql = sql + " order by " + strings.Join(sortParts, ", ")
	}

	if params.RowLimit > 0 || params.SkipRows > 0 {
		sql = sql + fmt.Sprintf(" limit %d offset %d", params.RowLimit, params.SkipRows)
	}
	return
}

func (model pgModel) getColumns(dbc *sql.DB, table *schema.Table) (cols []*schema.Column, err error) {
	// todo: parameterise
	sql := "select col.attname colname, col.attlen, typ.typname, col.attnotnull from pg_catalog.pg_attribute col inner join pg_catalog.pg_class tbl on col.attrelid = tbl.oid inner join pg_catalog.pg_namespace ns on ns.oid = tbl.relnamespace inner join pg_catalog.pg_type typ on typ.oid = col.atttypid where col.attnum > 0 and not col.attisdropped and ns.nspname = '" + table.Schema + "' and tbl.relname = '" + table.Name + "' order by col.attnum;"

	rows, err := dbc.Query(sql)
	if err != nil {
		log.Print(sql)
		return
	}
	defer rows.Close()
	cols = []*schema.Column{}
	colIndex := 0
	for rows.Next() {
		var len int
		var name, typeName string
		var notNull bool
		rows.Scan(&name, &len, &typeName, &notNull)
		thisCol := schema.Column{Position: colIndex, Name: name, Type: typeName, Nullable: !notNull}
		cols = append(cols, &thisCol)
		colIndex++
	}
	return
}

func (model pgModel) SetTableDescription(database string, table string, description string) (err error) {
	return
}

func (model pgModel) SetColumnDescription(database string, table string, column string, description string) (err error) {
	return
}
