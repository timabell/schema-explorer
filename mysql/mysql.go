package mysql

import (
	"bitbucket.org/timabell/sql-data-viewer/options"
	"bitbucket.org/timabell/sql-data-viewer/params"
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strconv"
	"strings"
)

type mysqlModel struct {
	connectionString string
}

type mysqlOpts struct {
	Host             *string `long:"host" description:"MySql host" env:"host"`
	Port             *int    `long:"port" description:"MySql port" env:"port"`
	Database         *string `long:"database" description:"MySql database name" env:"database"`
	User             *string `long:"user" description:"MySql username" env:"user"`
	Password         *string `long:"password" description:"MySql password" env:"password"`
	Parameters       *string `long:"parameters" description:"MySql extra parameters" env:"parameters"`
	ConnectionString *string `long:"connection-string" description:"MySql connection string. Use this instead of host, port etc for advanced driver options. See https://github.com/Go-SQL-Driver/MySQL/#dsn-data-source-name for connection-string options." env:"connection_string"`
}

func (opts mysqlOpts) validate() error {
	if opts.hasAnyDetails() && opts.ConnectionString != nil {
		return errors.New("Specify either a connection string or host etc, not both.")
	}
	return nil
}

func (opts mysqlOpts) hasAnyDetails() bool {
	return opts.Host != nil ||
		opts.Port != nil ||
		opts.Database != nil ||
		opts.User != nil ||
		opts.Password != nil
}

var opts = &mysqlOpts{}

func init() {
	// https://github.com/jessevdk/go-flags/blob/master/group_test.go#L33
	reader.RegisterReader(&reader.Driver{Name: "mysql", Options: opts, CreateReader: newMysql, FullName: "MySql"})
}

func newMysql() reader.DbReader {
	err := opts.validate()
	if err != nil {
		log.Printf("Mysql args error: %s", err)
		options.ArgParser.WriteHelp(os.Stdout)
		os.Exit(1)
	}
	var cs string
	if opts.ConnectionString == nil {
		if opts.User != nil {
			cs = *opts.User
			if opts.Password != nil {
				cs = fmt.Sprintf("%s:%s", cs, *opts.Password)
			}
			cs = fmt.Sprintf("%s@", cs)
		}
		if opts.Host != nil {
			cs = fmt.Sprintf("%s%s", cs, *opts.Host)
			if opts.Port != nil {
				cs = fmt.Sprintf("%s:%d", cs, *opts.Port)
			}
		}
		cs = fmt.Sprintf("%s/", cs)
		if opts.Database != nil {
			cs = fmt.Sprintf("%s%d", cs, *opts.Database)
		}
		if opts.Parameters != nil {
			cs = fmt.Sprintf("%s?%s", cs, *opts.Parameters)
		}
	} else {
		cs = *opts.ConnectionString
	}
	log.Println("Connecting to mysql db")
	return mysqlModel{
		connectionString: cs,
	}
}

func (model mysqlModel) ReadSchema() (database *schema.Database, err error) {
	dbc, err := getConnection(model.connectionString)
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

func (model mysqlModel) ListDatabases() (databaseList []string, err error) {
	sql := "select datname from mysql_database where datistemplate = false;"

	dbc, err := getConnection(model.connectionString)
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

func (model mysqlModel) DatabaseSelected() bool {
	return opts.Database != nil || opts.ConnectionString != nil
}

func (model mysqlModel) UpdateRowCounts(database *schema.Database) (err error) {
	for _, table := range database.Tables {
		rowCount, err := model.getRowCount(table)
		if err != nil {
			log.Printf("Failed to get row count for %s, %s", table, err)
		}
		table.RowCount = &rowCount
	}
	return err
}

func (model mysqlModel) getRowCount(table *schema.Table) (rowCount int, err error) {
	// todo: parameterise where possible
	// todo: whitelist-sanitize unparameterizable parts
	sql := "select count(*) from \"" + table.Schema + "\".\"" + table.Name + "\""

	dbc, err := getConnection(model.connectionString)
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

func (model mysqlModel) getTables(dbc *sql.DB) (tables []*schema.Table, err error) {
	rows, err := dbc.Query("select table_schema, table_name from information_schema.tables;")
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
	dbc, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Println("connection error", err)
	}
	return
}

func (model mysqlModel) CheckConnection() (err error) {
	dbc, err := getConnection(model.connectionString)
	if dbc == nil {
		log.Println(err)
		panic("getConnection() returned nil")
	}
	defer dbc.Close()
	tables, err := model.getTables(dbc)
	if err != nil {
		err = errors.New("getTables() failed - " + err.Error())
		return
	}
	log.Println("Connected.", len(tables), "tables found")
	return
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
				select mysqlc.oid, mysqlc.connamespace, mysqlc.conrelid, mysqlc.confrelid, mysqlc.contype, mysqlc.conname,
				       unnest(case when mysqlc.conkey <> '{}' then mysqlc.conkey else '{null}' end) as conkey,
				       unnest(case when mysqlc.confkey <> '{}' then mysqlc.confkey else '{null}' end) as confkey
				from mysql_constraint mysqlc
			) as con
			inner join mysql_namespace ns on con.connamespace = ns.oid
			inner join mysql_class tbl on tbl.oid = con.conrelid
			inner join mysql_namespace tns on tbl.relnamespace = tns.oid
			inner join mysql_attribute col on col.attrelid = tbl.oid and col.attnum = con.conkey
			left outer join mysql_class ftbl on ftbl.oid = con.confrelid
			left outer join mysql_namespace fns on ftbl.relnamespace = fns.oid
			left outer join mysql_attribute fcol on fcol.attrelid = ftbl.oid and fcol.attnum = con.confkey;`)

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
			select *, unnest(indkey) colnum from mysql_index
		) ix
		left outer join mysql_class oc on oc.oid = ix.indexrelid
		left outer join mysql_class tbl on tbl.oid = ix.indrelid
		left outer join mysql_namespace tns on tbl.relnamespace = tns.oid
		left outer join mysql_attribute col on col.attrelid = ix.indrelid and col.attnum = ix.colnum
		where tns.nspname not like 'mysql_%'
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
		if colName != "" { // more complex indexes don't link back to their columns. See mysql_index.indkey https://www.mysqlql.org/docs/current/static/catalog-mysql-index.html
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

func (model mysqlModel) GetSqlRows(table *schema.Table, params *params.TableParams, peekFinder *reader.PeekLookup) (rows *sql.Rows, err error) {
	dbc, err := getConnection(model.connectionString)
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

func (model mysqlModel) GetRowCount(table *schema.Table, params *params.TableParams) (rowCount int, err error) {
	dbc, err := getConnection(model.connectionString)
	if err != nil {
		log.Print("GetRows failed to get connection")
		return
	}
	defer dbc.Close()

	sql, values := buildQuery(table, params, &reader.PeekLookup{})
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

func (model mysqlModel) GetAnalysis(table *schema.Table) (analysis []schema.ColumnAnalysis, err error) {
	// todo, might be good to stream this all the way to the http response
	dbc, err := getConnection(model.connectionString)
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

func buildQuery(table *schema.Table, params *params.TableParams, peekFinder *reader.PeekLookup) (sql string, values []interface{}) {
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

func (model mysqlModel) getColumns(dbc *sql.DB, table *schema.Table) (cols []*schema.Column, err error) {
	// todo: parameterise
	sql := "select col.attname colname, col.attlen, typ.typname, col.attnotnull from mysql_catalog.mysql_attribute col inner join mysql_catalog.mysql_class tbl on col.attrelid = tbl.oid inner join mysql_catalog.mysql_namespace ns on ns.oid = tbl.relnamespace inner join mysql_catalog.mysql_type typ on typ.oid = col.atttypid where col.attnum > 0 and not col.attisdropped and ns.nspname = '" + table.Schema + "' and tbl.relname = '" + table.Name + "' order by col.attnum;"

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
