package sdv

import (
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
)

type pageTemplateModel struct {
	Title       string
	Db          string
	About       aboutType
	Copyright   string
	LicenseText string
	Timestamp   string
}

type tablesViewModel struct {
	LayoutData pageTemplateModel
	Database   schema.Database
	Diagram    diagramViewModel
}

type diagramViewModel struct {
	Tables     []*schema.Table
	TableLinks []fkViewModel
}

type fkViewModel struct {
	Source      schema.Table
	Destination schema.Table
}

type cells []template.HTML

type FieldFilter struct {
	Field  string
	Values []string
}

type dataViewModel struct {
	LayoutData pageTemplateModel
	Table      schema.Table
	Query      []FieldFilter
	RowLimit   int
	Rows       []cells
	Diagram    diagramViewModel
}

var tablesTemplate *template.Template
var tableTemplate *template.Template
var layoutData pageTemplateModel

func SetupTemplate() {
	templates, err := template.Must(template.ParseGlob("templates/layout.tmpl")).ParseGlob("templates/_*.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tablesTemplate, err = template.Must(templates.Clone()).ParseGlob("templates/tables.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tableTemplate, err = template.Must(templates.Clone()).ParseGlob("templates/table.tmpl")
	if err != nil {
		log.Fatal(err)
	}
}

func showTableList(resp http.ResponseWriter, database schema.Database) {
	var tableLinks []fkViewModel
	for _, fk := range database.Fks {
		tableLinks = append(tableLinks, fkViewModel{Source: *fk.SourceTable, Destination: *fk.DestinationTable})
	}

	model := tablesViewModel{
		LayoutData: layoutData,
		Database:   database,
		Diagram:    diagramViewModel{Tables: database.Tables, TableLinks: tableLinks},
	}

	err := tablesTemplate.ExecuteTemplate(resp, "layout", model)
	if err != nil {
		log.Fatal(err)
	}
}

func showTable(resp http.ResponseWriter, reader dbReader, table *schema.Table, query schema.RowFilter, rowLimit int) error {
	fieldFilter := make([]FieldFilter, 0)
	if len(query) > 0 {
		fieldKeys := make([]string, 0)
		for field, _ := range query {
			fieldKeys = append(fieldKeys, field)
		}
		sort.Strings(fieldKeys)
		for _, field := range fieldKeys {
			fieldFilter = append(fieldFilter, FieldFilter{Field: field, Values: query[field]})
		}
	}

	rowsData, err := GetRows(reader, query, table, rowLimit)
	if err != nil {
		return err
	}

	rows := []cells{}
	for _, rowData := range rowsData {
		row := buildRow(rowData, table)
		rows = append(rows, row)
	}

	diagramTables := []*schema.Table{table}
	var tableLinks []fkViewModel
	for _, tableFks := range table.Fks {
		diagramTables = append(diagramTables, tableFks.DestinationTable)
		tableLinks = append(tableLinks, fkViewModel{Source: *tableFks.SourceTable, Destination: *tableFks.DestinationTable})
	}
	for _, inboundFks := range table.InboundFks {
		diagramTables = append(diagramTables, inboundFks.SourceTable)
		tableLinks = append(tableLinks, fkViewModel{Source: *inboundFks.SourceTable, Destination: *inboundFks.DestinationTable})
	}

	viewModel := dataViewModel{
		LayoutData: layoutData,
		Table:      *table,
		Query:      fieldFilter,
		RowLimit:   rowLimit,
		Rows:       rows,
		Diagram:    diagramViewModel{Tables: diagramTables, TableLinks: tableLinks},
	}

	err = tableTemplate.ExecuteTemplate(resp, "layout", viewModel)
	if err != nil {
		log.Print("template execution error", err)
		panic(err)
	}

	return nil
}

func buildRow(rowData RowData, table *schema.Table) cells {
	row := cells{}
	for colIndex, col := range table.Columns {
		cellData := rowData[colIndex]
		valueHTML := buildCell(col, cellData)
		row = append(row, template.HTML(valueHTML))
	}
	parentHTML := buildInwardCell(table.InboundFks, rowData, table.Columns)
	row = append(row, template.HTML(parentHTML))
	return row
}

func buildInwardCell(inboundFks []*schema.Fk, rowData []interface{}, cols []*schema.Column) string {
	// todo: post-refactor fixup
	// todo: performance - pre-calculate fk info so this isn't repeated for every row
	// stable sort order http://stackoverflow.com/questions/23330781/sort-golang-map-values-by-keys
	//tables := make([]schema.Table, 0)
	//for _, fk := range inboundFks {
	//	tables = append(tables, fk.SourceTable)
	//}
	//sort.Strings(tables)
	parentHTML := ""
	//for _, fk := range inboundFks {
	//	parentHTML = parentHTML + template.HTMLEscapeString(fk.SourceTable.String()) + ":&nbsp;"
	//	parentCols := make([]string, 0)
	//	for colKey, _ := range inwardFks[table] {
	//		parentCols = append(parentCols, colKey)
	//	}
	//	sort.Strings(parentCols)
	//	for _, parentCol := range parentCols {
	//		parentHTML = parentHTML + buildInwardLink(table, parentCol, rowData, cols, inwardFks[table][parentCol])
	//	}
	//	parentHTML = parentHTML + " "
	//}
	return parentHTML
}

//func buildInwardLink(parentTable string, parentCol string, rowData RowData, cols []schema.Column, ref schema.Ref) string {
//	linkHTML := fmt.Sprintf(
//		"<a href='%s?%s=",
//		template.URLQueryEscaper(parentTable),
//		template.URLQueryEscaper(parentCol))
//	// todo: handle non-string primary key
//	// todo: handle compound primary key
//	colData := rowData[indexOfCol(cols, string(ref.Col))]
//	switch colData.(type) {
//	case int64:
//		// todo: url-escape as well
//		linkHTML = linkHTML + template.HTMLEscapeString(fmt.Sprintf("%d", colData))
//	case string:
//		// todo: sql-quotes here are a hack pending switching to parameterized sql
//		linkHTML = linkHTML + "%27" + template.HTMLEscapeString(fmt.Sprintf("%s", colData)) + "%27"
//	default:
//		linkHTML = linkHTML + template.HTMLEscapeString(fmt.Sprintf("%v", colData))
//	}
//	linkHTML = linkHTML + fmt.Sprintf(
//		// todo: factor out row limit, move to a cookie perhaps
//		"&_rowLimit=100' class='parentFk'>%s</a>&nbsp;",
//		template.HTMLEscaper(parentCol))
//	return linkHTML
//}

func buildCell(col *schema.Column, cellData interface{}) string {
	if cellData == nil {
		return "<span class='null'>[null]</span>"
	}
	var valueHTML string
	hasFk := col.Fk != nil
	stringValue := *DbValueToString(cellData, col.Type)
	if hasFk {
		// todo: compound-fk support
		valueHTML = fmt.Sprintf("<a href='%s?%s=", col.Fk.DestinationTable, col.Fk.DestinationColumns[0].Name)
		// todo: url-escape as well as htmlencode
		switch {
		case strings.Contains(col.Type, "varchar"):
			// todo: sql-quotes here are a hack pending switching to parameterized sql
			valueHTML = valueHTML + "%27" + template.HTMLEscapeString(stringValue) + "%27"
		default:
			valueHTML = valueHTML + template.HTMLEscapeString(stringValue)
		}
		valueHTML = valueHTML + "' class='fk'>"
	}
	valueHTML = valueHTML + template.HTMLEscapeString(stringValue)
	if hasFk {
		valueHTML = valueHTML + "</a>"
	}
	return valueHTML
}

func indexOfCol(cols []schema.Column, name string) (index int) {
	var curValue schema.Column
	for index, curValue = range cols {
		if curValue.Name == name {
			return
		}
	}
	log.Panic(name, " not found in column list")
	return
}
