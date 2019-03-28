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
			cs = fmt.Sprintf("%s%s", cs, *opts.Database)
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
			Schema:               false,
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
	sql := "select schema_name from information_schema.schemata where schema_name not in ('information_schema', 'mysql');"

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
	sql := "select count(*) from `" + table.Name + "`"

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
	rows, err := dbc.Query("select table_name from information_schema.tables where table_schema = database();")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, &schema.Table{Name: name, Pk: &schema.Pk{}})
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
	dbc.Ping()
	log.Println("Connected.")
	return
}

func readConstraints(dbc *sql.DB, database *schema.Database) (err error) {
	sql := fmt.Sprintf(`
			select
				tc.constraint_type, tc.constraint_name,
				tc.table_name, kc.column_name,
				kc.referenced_table_name, kc.referenced_column_name
			from information_schema.table_constraints tc
			left outer join information_schema.key_column_usage kc
				on kc.constraint_schema = tc.constraint_schema
					and kc.constraint_name = tc.constraint_name
					and kc.table_name = tc.table_name
			where tc.constraint_schema=database()
			order by tc.constraint_type, tc.constraint_name,
				kc.ordinal_position,
				tc.table_name, kc.column_name,
				kc.referenced_table_name, kc.referenced_column_name;`)

	rows, err := dbc.Query(sql)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var conType, name,
			sourceTableName, sourceColumnName,
			destinationTableName, destinationColumnName string
		rows.Scan(&conType, &name,
			&sourceTableName, &sourceColumnName,
			&destinationTableName, &destinationColumnName)
		tableToFind := &schema.Table{Name: sourceTableName}
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
		case "FOREIGN KEY":
			destinationTable := database.FindTable(&schema.Table{Name: destinationTableName})
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
		case "PRIMARY KEY":
			//log.Printf("pk: %s.%s", sourceTable, sourceColumn)
			sourceTable.Pk.Columns = append(sourceTable.Pk.Columns, sourceColumn)
			sourceColumn.IsInPrimaryKey = true
		case "UNIQUE": // todo
		default:
			log.Printf("?? %s", conType)
		}
	}
	return
}

func readIndexes(dbc *sql.DB, database *schema.Database) (err error) {
	sql := `
		select index_name, table_name, column_name, non_unique
		from information_schema.statistics
		where table_schema=database() order by seq_in_index;
	`

	//log.Println(sql)
	rows, err := dbc.Query(sql)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var name, tableName, colName string
		var nonUnique bool
		rows.Scan(&name, &tableName, &colName, &nonUnique)
		tableToFind := &schema.Table{Name: tableName}
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
				Name:     name,
				Columns:  []*schema.Column{},
				IsUnique: !nonUnique,
				Table:    table,
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
	statement, err := dbc.Prepare(sql)
	if err != nil {
		log.Print("GetRows failed to prepare query")
		log.Println(sql)
		log.Println(err)
	}
	rows, err = statement.Query(values...)
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
		sql := "select `" + col.Name + "`, count(*) qty from `" + table.Name + "` group by `" + col.Name + "` order by count(*) desc, `" + col.Name + "` limit 100;"
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
			sql = sql + fmt.Sprintf(", fk%d.`%s` fk%d_%s", fkIndex, peekCol, fkIndex, peekCol)
		}
	}

	// inbound fk counts
	for inboundFkIndex, inboundFk := range table.InboundFks {
		onPredicates := []string{}
		for ix, sourceCol := range inboundFk.SourceColumns {
			onPredicates = append(onPredicates, fmt.Sprintf("ifk%d.`%s` = t.`%s`", inboundFkIndex, sourceCol.Name, inboundFk.DestinationColumns[ix].Name))
		}
		onString := strings.Join(onPredicates, " and ")
		sql = sql + fmt.Sprintf(", (select count(*) from `%s` ifk%d where %s) ifk%d_count", inboundFk.SourceTable.Name, inboundFkIndex, onString, inboundFkIndex)
	}

	sql = sql + " from `" + table.Name + "` t"

	// peek tables
	for fkIndex, fk := range peekFinder.Fks {
		sql = sql + fmt.Sprintf(" left outer join `%s` fk%d on ", fk.DestinationTable.Name, fkIndex)
		onPredicates := []string{}
		for ix, sourceCol := range fk.SourceColumns {
			onPredicates = append(onPredicates, fmt.Sprintf("t.`%s` = fk%d.`%s`", sourceCol, fkIndex, fk.DestinationColumns[ix]))
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
			clauses = append(clauses, "t.`"+col.Name+"` = ?")
			index = index + 1
			values = append(values, v.Values[0]) // todo: maybe support multiple values
		}
		sql = sql + strings.Join(clauses, " and ")
	}

	if len(params.Sort) > 0 {
		var sortParts []string
		for _, sortCol := range params.Sort {
			sortString := "`" + sortCol.Column.Name + "`"
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
	// todo: read all tables' columns in one query hit
	sql := fmt.Sprintf("select column_name, data_type, is_nullable, character_maximum_length from information_schema.columns where table_schema = '%s' and table_name='%s' order by ordinal_position;", *opts.Database, table.Name)

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
		var name, typeName, isNullable string
		rows.Scan(&name, &typeName, &isNullable, &len)
		if strings.Contains(typeName, "char") {
			typeName = fmt.Sprintf("%s(%d)", typeName, len)
		}
		nullable := isNullable == "YES"
		thisCol := schema.Column{Position: colIndex, Name: name, Type: typeName, Nullable: nullable}
		cols = append(cols, &thisCol)
		colIndex++
	}
	return
}
