package pg

import (
	"bitbucket.org/timabell/sql-data-viewer/params"
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"strconv"
	"strings"
	"os"
)

type pgModel struct {
	connectionString string
}

type pgOpts struct {
	// todo: break down into host, port etc
	Db *string `long:"pg-db" description:"Postgres connection string. # see https://godoc.org/github.com/lib/pq for connection-string options" env:"schemaexplorer_pg_db"`
}

var opt = &pgOpts{}

func init() {
	// https://github.com/jessevdk/go-flags/blob/master/group_test.go#L33
	reader.RegisterReader("pg", opt, NewPg)
}

func NewPg() reader.DbReader {
	if opt.Db == nil {
		log.Printf("Error: connection string (pg-db) is required")
		reader.ArgParser.WriteHelp(os.Stdout)
		os.Exit(1)
	}
	log.Println("Connecting to pg db")
	return pgModel{
		connectionString: *opt.Db,
	}
}

func (model pgModel) ReadSchema() (database *schema.Database, err error) {
	dbc, err := getConnection(model.connectionString)
	if err != nil {
		return
	}
	defer dbc.Close()

	database = &schema.Database{
		Supports: schema.SupportedFeatures{
			Schema:       true,
			Descriptions: false,
			FkNames:      true,
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

	// fks
	for _, table := range database.Tables {
		err = readConstraints(dbc, table, database)
		if err != nil {
			return
		}
		database.Fks = append(database.Fks, table.Fks...)
	}

	// hook-up inbound fks
	for _, fk := range database.Fks {
		destination := database.FindTable(fk.DestinationTable)
		if destination == nil {
			err = fmt.Errorf("failed to find destination table for fk %s", fk)
		}
		destination.InboundFks = append(destination.InboundFks, fk)
	}
	//log.Print(database.DebugString())
	return
}

func (model pgModel) getRowCount(table *schema.Table) (rowCount int, err error) {
	// todo: parameterise where possible
	// todo: whitelist-sanitize unparameterizable parts
	sql := "select count(*) from \"" + table.Name + "\""

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
	for _, table := range tables {
		rowCount, err := model.getRowCount(table)
		if err != nil {
			log.Printf("Failed to get row count for %d", table)
		}
		table.RowCount = &rowCount
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

func (model pgModel) CheckConnection() (err error) {
	dbc, err := getConnection(model.connectionString)
	if dbc == nil {
		log.Println(err)
		panic("getConnection() returned nil")
	}
	defer dbc.Close()
	tables, err := model.getTables(dbc)
	if err != nil {
		panic(err)
	}
	if len(tables) == 0 {
		panic("No tables found.")
	}
	log.Println("Connected.", len(tables), "tables found")
	return
}

func readConstraints(dbc *sql.DB, sourceTable *schema.Table, database *schema.Database) (err error) {
	// todo: parameterise
	// todo: support multi-column FKs
	// null-proof unnest: https://stackoverflow.com/a/49736694
	sql := fmt.Sprintf(`select con.oid, con.contype, col.attname column_name, ftbl.relname, fcol.attname foreign_column, con.conname
		from
			(
				select pgc.oid, pgc.connamespace, pgc.conrelid, pgc.confrelid, pgc.contype, pgc.conname,
				       unnest(case when pgc.conkey <> '{}' then pgc.conkey else '{null}' end) as conkey,
				       unnest(case when pgc.confkey <> '{}' then pgc.confkey else '{null}' end) as confkey
				from pg_constraint pgc
			) as con
			inner join pg_namespace ns on con.connamespace = ns.oid
			inner join pg_class tbl on tbl.oid = con.conrelid
			inner join pg_attribute col on col.attrelid = tbl.oid and col.attnum = con.conkey
			left outer join pg_class ftbl on ftbl.oid = con.confrelid
			left outer join pg_attribute fcol on fcol.attrelid = ftbl.oid and fcol.attnum = con.confkey
		where ns.nspname = '%s' and tbl.relname = '%s';`,
		sourceTable.Schema, sourceTable.Name)

	rows, err := dbc.Query(sql)
	if err != nil {
		return
	}
	defer rows.Close()
	var fks []*schema.Fk
	for rows.Next() {
		var oid, conType, sourceColumnName, destinationTableName, destinationColumnName, name string
		rows.Scan(&oid, &conType, &sourceColumnName, &destinationTableName, &destinationColumnName, &name)
		_, sourceColumn := sourceTable.FindColumn(sourceColumnName)
		switch conType {
		case "f": // f = foreign key
			destinationTable := database.FindTable(&schema.Table{Schema: database.DefaultSchemaName, Name: destinationTableName})
			if destinationTable == nil {
				log.Print(database.DebugString())
				panic(fmt.Sprintf("couldn't find table %s in database object while hooking up fks", destinationTableName))
			}
			_, destinationColumn := destinationTable.FindColumn(destinationColumnName)
			// see if we are adding columns to an existing fk
			var fk *schema.Fk
			for _, existingFk := range fks {
				if existingFk.Name == name {
					existingFk.SourceColumns = append(existingFk.SourceColumns, sourceColumn)
					existingFk.DestinationColumns = append(existingFk.DestinationColumns, destinationColumn)
					fk = existingFk
					break
				}
			}
			if fk == nil {
				fk = schema.NewFk(name, sourceTable, sourceColumn, destinationTable, destinationColumn)
				fks = append(fks, fk)
			}
			sourceColumn.Fks = append(sourceColumn.Fks, fk)
			//log.Printf("fk: %+v - oid %+v", fk, oid)
		case "p":
			//log.Printf("pk: %s.%s", sourceTable, sourceColumn)
			sourceTable.Pk.Columns = append(sourceTable.Pk.Columns, sourceColumn)
			sourceColumn.IsInPrimaryKey = true
			//default:
			//	log.Printf("?? %s", conType)
		}
	}
	sourceTable.Fks = fks
	return
}

func (model pgModel) GetSqlRows(table *schema.Table, params *params.TableParams) (rows *sql.Rows, err error) {
	// todo: parameterise where possible
	// todo: whitelist-sanitize unparameterizable parts
	sql := "select * from \"" + table.Name + "\""

	var values []interface{}
	query := params.Filter
	if len(query) > 0 {
		sql = sql + " where "
		clauses := make([]string, 0, len(query))
		values = make([]interface{}, 0, len(query))
		var index = 1
		for _, v := range query {
			col := v.Field
			clauses = append(clauses, "\""+col.Name+"\" = $"+strconv.Itoa(index))
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

	if params.RowLimit > 0 {
		sql = sql + " limit " + strconv.Itoa(params.RowLimit)
	}

	dbc, err := getConnection(model.connectionString)
	if err != nil {
		log.Print("GetRows failed to get connection")
		return
	}
	defer dbc.Close()

	log.Println(sql)
	rows, err = dbc.Query(sql, values...)
	if err != nil {
		log.Print("GetRows failed to get query")
		log.Println(sql)
		log.Println(err)
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
