package mssql

import (
	"bitbucket.org/timabell/sql-data-viewer/about"
	"bitbucket.org/timabell/sql-data-viewer/params"
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"log"
	"os"
	"strconv"
	"strings"
)

type mssqlModel struct {
	connectionString string
}

type mssqlOpts struct {
	Host             *string `long:"host" description:"Sql Server host or address" env:"host"`
	Port             *int    `long:"port" description:"Sql Server port" env:"port"`
	Instance         *string `long:"instance" description:"Sql Server instance name" env:"instance"`
	Database         *string `long:"database" description:"Sql Server database name" env:"database"`
	User             *string `long:"user" description:"Sql Server username for sql-auth. Leave out to use integrated auth." env:"user"`
	Password         *string `long:"password" description:"Sql Server password for sql-auth" env:"password"`
	ConnectionString *string `long:"connection-string" description:"Sql Server connection string. Use this instead of host, port etc for advanced driver options. See https://github.com/simnalamburt/go-mssqldb#connection-parameters-and-dsn for connection-string options." env:"connection_string"`
}

var opts = &mssqlOpts{}

func init() {
	// https://github.com/jessevdk/go-flags/blob/master/group_test.go#L33
	reader.RegisterReader("mssql", opts, NewMssql)
}

func (opts mssqlOpts) validate() error {
	if opts.hasAnyDetails() && opts.ConnectionString != nil {
		return errors.New("Specify either a connection string or host etc, not both.")
	}
	return nil
}

func (opts mssqlOpts) hasAnyDetails() bool {
	return opts.Host != nil ||
		opts.Port != nil ||
		opts.Database != nil ||
		opts.User != nil ||
		opts.Password != nil
}

func NewMssql() reader.DbReader {
	err := opts.validate()
	if err != nil {
		log.Printf("Mssql args error: %s", err)
		reader.ArgParser.WriteHelp(os.Stdout)
		os.Exit(1)
	}
	var cs string
	if opts.ConnectionString == nil {
		optList := make(map[string]string)
		if opts.Host != nil {
			if opts.Instance != nil {
				optList["server"] = fmt.Sprintf("%s\\%s", *opts.Host, *opts.Instance)
			} else {
				optList["server"] = *opts.Host
			}
		} else {
			if opts.Instance != nil {
				optList["server"] = fmt.Sprintf("localhost\\%s", *opts.Instance)
			}
		}
		if opts.Port != nil {
			optList["port"] = strconv.Itoa(*opts.Port)
		}
		if opts.Database != nil {
			optList["database"] = *opts.Database
		}
		if opts.User != nil {
			optList["user id"] = *opts.User
		}
		if opts.Password != nil {
			optList["password"] = *opts.Password
		}
		optList["app-name"] = about.About.Summary()
		pairs := []string{}
		for key, value := range optList {
			pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
		}
		cs = strings.Join(pairs, ";")
	} else {
		cs = *opts.ConnectionString
	}
	log.Println("Connecting to mssql db")
	return mssqlModel{
		connectionString: cs,
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
			Schema:               true,
			Descriptions:         true,
			FkNames:              true,
			PagingWithoutSorting: false,
		},
		DefaultSchemaName: "dbo",
	}

	database.Tables, err = getTables(dbc)
	if err != nil {
		return
	}

	// columns
	for tableNumber, table := range database.Tables {
		var cols []*schema.Column
		cols, err = getColumns(dbc, table)
		if err != nil {
			return
		}
		database.Tables[tableNumber].Columns = append(table.Columns, cols...)
	}

	database.Fks, err = allFks(dbc, database)
	if err != nil {
		return
	}

	// attach fks to inbound columns
	for _, fk := range database.Fks {
		for _, col := range fk.DestinationColumns {
			col.InboundFks = append(col.InboundFks, fk)
		}
	}

	getIndexes(dbc, database)

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

func getTables(dbc *sql.DB) (tables []*schema.Table, err error) {

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
	return tables, nil
}

func (model mssqlModel) UpdateRowCounts(database *schema.Database) (err error) {
	for _, table := range database.Tables {
		rowCount, err := model.getRowCount(table)
		if err != nil {
			log.Printf("Failed to get row count for %s", table)
		}
		table.RowCount = &rowCount
	}
	return err
}

func (model mssqlModel) getRowCount(table *schema.Table) (rowCount int, err error) {
	// todo: parameterise where possible
	// todo: whitelist-sanitize unparameterizable parts
	sql := "select count(*) from [" + table.Schema + "].[" + table.Name + "]"

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

func allFks(dbc *sql.DB, database *schema.Database) (allFks []*schema.Fk, err error) {
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
	dbc, err := getConnection(model.connectionString)
	if dbc == nil {
		log.Println(err)
		panic("getConnection() returned nil")
	}
	defer dbc.Close()

	sql, values := buildQuery(table, params)
	rows, err = dbc.Query(sql, values...)
	if params.SkipRows > 0 && len(params.Sort) == 0 {
		// Can't use offset or row_number without a sort order so use a hack.
		// buildQuery has given us rowlimit+skip rows so now we just need to discard the unwanted leading rows
		for i := 0; i < params.SkipRows; i++ {
			if !rows.Next() {
				break // reached end of dataset
			}
		}
	}
	if rows == nil {
		log.Println(sql)
		log.Println(err)
		panic("Query returned nil for rows")
	}
	return
}

func (model mssqlModel) GetRowCount(table *schema.Table, params *params.TableParams) (rowCount int, err error) {
	dbc, err := getConnection(model.connectionString)
	if err != nil {
		log.Print("GetRows failed to get connection")
		return
	}
	defer dbc.Close()

	sql, values := buildQuery(table, params)
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

func (model mssqlModel) GetAnalysis(table *schema.Table) (analysis []schema.ColumnAnalysis, err error) {
	dbc, err := getConnection(model.connectionString)
	if err != nil {
		log.Print("GetAnalysis failed to get connection")
		return
	}
	defer dbc.Close()

	analysis = []schema.ColumnAnalysis{}
	for _, col := range table.Columns {
		sql := "select top 100 [" + col.Name + "], count(*) qty from [" + table.Schema + "].[" + table.Name + "] group by [" + col.Name + "] order by count(*) desc, [" + col.Name + "];"
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

func buildQuery(table *schema.Table, params *params.TableParams) (sql string, values []interface{}) {
	// Limitation: we can't support paging (offset/skip) without a sort order so
	// 		params.SkipRows will be ignored if there is no sorting supplied.
	// As a less performant alternative to keep things consistent we'll fetch the preceding rows and throw them away

	sql = "select"

	// use top when we have a row limit but now sorting (or can't use offset because there's no sort)
	if params.RowLimit > 0 && len(params.Sort) == 0 {
		sql = sql + " top " + strconv.Itoa(params.RowLimit+params.SkipRows)
	}

	sql = sql + " * from [" + table.Schema + "].[" + table.Name + "]"

	query := params.Filter
	if len(query) > 0 {
		sql = sql + " where "
		clauses := make([]string, 0, len(query))
		values = make([]interface{}, 0, len(query))
		for _, v := range query {
			col := v.Field
			clauses = append(clauses, "["+col.Name+"] = ?")
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

		if params.SkipRows > 0 {
			sql = sql + fmt.Sprintf(" offset %d rows", params.SkipRows)
			if params.RowLimit > 0 {
				sql = sql + fmt.Sprintf(" fetch next %d rows only", params.RowLimit)
			}
		}
	}
	return
}

func getColumns(dbc *sql.DB, table *schema.Table) (cols []*schema.Column, err error) {
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

func getIndexes(dbc *sql.DB, database *schema.Database) {
	rows, err := dbc.Query(`
		select
			ix.name index_name,
			s.name schema_name,
			t.name table_name,
			ix.is_primary_key,
			ix.is_disabled,
			ix.is_unique,
			ix.type_desc,
			col.name colname
		from sys.indexes ix
			inner join sys.index_columns ic on ic.object_id = ix.object_id and ic.index_id = ix.index_id
			inner join sys.tables t on t.object_id = ix.object_id
			inner join sys.columns col on col.object_id = ix.object_id and col.column_id = ic.column_id
			inner join sys.schemas s on s.schema_id = t.schema_id
		where s.name <> 'sys';
`)

	if err != nil {
		log.Fatal("error running query to find indexes: ", err)
		return
	}
	if rows == nil {
		log.Fatal("index rows was nil")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name, schemaName, tableName, typeDesc, columnName string
		var isPrimaryKey, isDisabled, isUnique bool
		rows.Scan(&name, &schemaName, &tableName, &isPrimaryKey, &isDisabled, &isUnique, &typeDesc, &columnName)
		isClustered := typeDesc == "CLUSTERED"
		table := database.FindTable(&schema.Table{Schema: schemaName, Name: tableName})
		if table == nil {
			log.Fatalf("Failed to find table %s for index %s", tableName, name)
		}
		_, col := table.FindColumn(columnName)
		if col == nil {
			log.Fatalf("Failed to find col %s in table %s for index %s", columnName, tableName, name)
		}
		if isPrimaryKey {
			col.IsInPrimaryKey = true
			if table.Pk == nil {
				table.Pk = &schema.Pk{Name: name, Columns: schema.ColumnList{col}}
			} else {
				table.Pk.Columns = append(table.Pk.Columns, col)
			}
		} else { // normal index
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
			if columnName != "" {
				_, col := table.FindColumn(columnName)
				if col == nil {
					err = errors.New(fmt.Sprintf("Column %s in table %s not found, for index %s", columnName, table.String(), name))
					return
				}
				index.Columns = append(index.Columns, col)
				col.Indexes = append(col.Indexes, index)
			}
			//log.Printf(index.String())
		}
	}
}
