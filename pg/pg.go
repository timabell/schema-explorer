package pg

import (
	"bitbucket.org/timabell/sql-data-viewer/params"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"strconv"
	"strings"
)

type pgModel struct {
	connectionString string
}

func NewPg(connectionString string) pgModel {
	return pgModel{
		connectionString: connectionString,
	}
}

func (model pgModel) ReadSchema() (database *schema.Database, err error) {
	dbc, err := getConnection(model.connectionString)
	if err != nil {
		return
	}
	defer dbc.Close()

	database = &schema.Database{
		Supports:          schema.SupportedFeatures{Schema: true, Descriptions: false},
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
		var fks []*schema.Fk
		fks, err = getFks(dbc, table, database)
		if err != nil {
			return
		}
		table.Fks = fks
		database.Fks = append(database.Fks, fks...)
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

func (model pgModel) getTables(dbc *sql.DB) (tables []*schema.Table, err error) {
	rows, err := dbc.Query("select schemaname, tablename from pg_catalog.pg_tables where schemaname not in ('pg_catalog','information_schema')")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name, schemaName string
		rows.Scan(&schemaName, &name)
		tables = append(tables, &schema.Table{Schema: schemaName, Name: name})
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

func getFks(dbc *sql.DB, sourceTable *schema.Table, database *schema.Database) (fks []*schema.Fk, err error) {
	// todo: parameterise
	// todo: support multi-column FKs
	rows, err := dbc.Query("select col.attname column_name, ftbl.relname, fcol.attname foreign_column from pg_constraint con inner join pg_namespace ns on con.connamespace = ns.oid inner join pg_class tbl on tbl.oid = con.conrelid inner join pg_class ftbl on ftbl.oid = con.confrelid inner join pg_attribute col on col.attrelid = tbl.oid and col.attnum = con.conkey[1] inner join pg_attribute fcol on fcol.attrelid = ftbl.oid and fcol.attnum = con.confkey[1] where con.contype = 'f' and ns.nspname = '" + sourceTable.Schema + "' and tbl.relname = '" + sourceTable.Name + "';")
	if err != nil {
		return
	}
	defer rows.Close()
	fks = []*schema.Fk{}
	for rows.Next() {
		var sourceColumnName, destinationTableName, destinationColumnName string
		rows.Scan(&sourceColumnName, &destinationTableName, &destinationColumnName)
		_, sourceColumn := sourceTable.FindColumn(sourceColumnName)
		// todo: read schema of fk table
		destinationTable := database.FindTable(&schema.Table{Schema: database.DefaultSchemaName, Name: destinationTableName})
		if destinationTable == nil {
			log.Print(database.DebugString())
			panic(fmt.Sprintf("couldn't find table %s in database object while hooking up fks", destinationTableName))
		}
		_, destinationColumn := destinationTable.FindColumn(destinationColumnName)
		fk := schema.NewFk("", sourceTable, sourceColumn, destinationTable, destinationColumn)
		sourceColumn.Fk = fk
		fks = append(fks, fk)
	}
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
	for rows.Next() {
		var len int
		var name, typeName string
		var notNull bool
		rows.Scan(&name, &len, &typeName, &notNull)
		thisCol := schema.Column{Name: name, Type: typeName}
		cols = append(cols, &thisCol)
	}
	return
}
