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
	Title     string
	Db        string
	About     aboutType
	Copyright string
	LicenseText string
	Timestamp string
}

type tablesViewModel struct {
	LayoutData pageTemplateModel
	Tables     []schema.Table
	Diagram    diagramViewModel
}

type diagramViewModel struct {
	Tables []string
	//Tables     []schema.Table // todo: use proper type
	TableLinks []fkViewModel
}

type fkViewModel struct {
	Source      string
	Destination string
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
	Cols       []schema.Column
	Rows       []cells
	Diagram    diagramViewModel
}

var templates *template.Template
var layoutData pageTemplateModel

func SetupTemplate() {
	templates = template.Must(template.ParseGlob("templates/*.tmpl"))
}

func showTableList(resp http.ResponseWriter, tables []schema.Table, fks schema.GlobalFkList) {
	var tableLinks []fkViewModel
	for table, keys := range fks {
		// todo: per field refs, the below is currently aggregated to table level
		for _, ref := range keys {
			tableLinks = append(tableLinks, fkViewModel{Source: table, Destination: ref.Table.String()})
		}
	}
	tableList := []string{}
	for _, table := range tables {
		tableList = append(tableList, table.String())
	}
	model := tablesViewModel{
		LayoutData: layoutData,
		Tables:     tables,
		Diagram:    diagramViewModel{Tables: tableList, TableLinks: tableLinks},
	}

	err := templates.ExecuteTemplate(resp, "tables", model)
	if err != nil {
		log.Fatal(err)
	}
}

func showTable(resp http.ResponseWriter, reader dbReader, table schema.Table, query schema.RowFilter, rowLimit int) error {
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

	fks, err := reader.AllFks()
	if err != nil {
		log.Println("error getting fks", err)
		panic("error getting fks")
		// todo: send 500 error to client
		return err
	}

	inwardFks := table.FindParents(fks)

	cols, err := reader.GetColumns(table)
	if err != nil {
		panic(err)
	}

	rowsData, err := GetRows(reader, query, table, len(cols), rowLimit)
	if err != nil {
		return err
	}

	rows := []cells{}
	for _, rowData := range rowsData {
		row := buildRow(cols, rowData, fks, table, inwardFks)
		rows = append(rows, row)
	}

	//diagramTables:=[]schema.Table{table}
	diagramTables := []string{table.String()}
	for srcTable, tableFks := range fks {
		if srcTable == table.String() {
			for _, ref := range tableFks {
				diagramTables = append(diagramTables, ref.Table.String())
			}
		}
	}
	tableLinks := []fkViewModel{}
	for srcTable, tableFks := range fks {
		if srcTable == table.String() {
			for _, ref := range tableFks {
				tableLinks = append(tableLinks, fkViewModel{Source: srcTable, Destination: ref.Table.String()})
			}
		}
	}
	for srcTable, inwardFk := range inwardFks {
		diagramTables = append(diagramTables, srcTable)
		for _, ref := range inwardFk {
			tableLinks = append(tableLinks, fkViewModel{Source: srcTable, Destination: ref.Table.String()})
		}
	}

	viewModel := dataViewModel{
		LayoutData: layoutData,
		Table:      table,
		Query:      fieldFilter,
		RowLimit:   rowLimit,
		Cols:       cols,
		Rows:       rows,
		Diagram:    diagramViewModel{Tables: diagramTables, TableLinks: tableLinks},
	}

	err = templates.ExecuteTemplate(resp, "table", viewModel)
	if err != nil {
		log.Print("template execution error", err)
		panic(err)
	}

	return nil
}

func buildRow(cols []schema.Column, rowData RowData, fks schema.GlobalFkList, table schema.Table, inwardFks schema.GlobalFkList) cells {
	row := cells{}
	for colIndex, col := range cols {
		cellData := rowData[colIndex]
		valueHTML := buildCell(fks, table, col, cellData)
		row = append(row, template.HTML(valueHTML))
	}
	parentHTML := buildInwardCell(inwardFks, rowData, cols)
	row = append(row, template.HTML(parentHTML))
	return row
}

func buildInwardCell(inwardFks schema.GlobalFkList, rowData []interface{}, cols []schema.Column) string {
	// todo: pre-calculate fk info so this isn't repeated for every row
	// stable sort order http://stackoverflow.com/questions/23330781/sort-golang-map-values-by-keys
	tables := make([]string, 0)
	for key, _ := range inwardFks {
		tables = append(tables, key)
	}
	sort.Strings(tables)
	parentHTML := ""
	for _, table := range tables {
		parentHTML = parentHTML + template.HTMLEscapeString(table) + ":&nbsp;"
		parentCols := make([]string, 0)
		for colKey, _ := range inwardFks[table] {
			parentCols = append(parentCols, colKey)
		}
		sort.Strings(parentCols)
		for _, parentCol := range parentCols {
			parentHTML = parentHTML + buildInwardLink(table, parentCol, rowData, cols, inwardFks[table][parentCol])
		}
		parentHTML = parentHTML + " "
	}
	return parentHTML
}

func buildInwardLink(parentTable string, parentCol string, rowData RowData, cols []schema.Column, ref schema.Ref) string {
	linkHTML := fmt.Sprintf(
		"<a href='%s?%s=",
		template.URLQueryEscaper(parentTable),
		template.URLQueryEscaper(parentCol))
	// todo: handle non-string primary key
	// todo: handle compound primary key
	colData := rowData[indexOfCol(cols, string(ref.Col))]
	switch colData.(type) {
	case int64:
		// todo: url-escape as well
		linkHTML = linkHTML + template.HTMLEscapeString(fmt.Sprintf("%d", colData))
	case string:
		// todo: sql-quotes here are a hack pending switching to parameterized sql
		linkHTML = linkHTML + "%27" + template.HTMLEscapeString(fmt.Sprintf("%s", colData)) + "%27"
	default:
		linkHTML = linkHTML + template.HTMLEscapeString(fmt.Sprintf("%v", colData))
	}
	linkHTML = linkHTML + fmt.Sprintf(
		// todo: factor out row limit, move to a cookie perhaps
		"&_rowLimit=100' class='parentFk'>%s</a>&nbsp;",
		template.HTMLEscaper(parentCol))
	return linkHTML
}

func buildCell(fks schema.GlobalFkList, table schema.Table, col schema.Column, cellData interface{}) string {
	if cellData == nil {
		return "<span class='null'>[null]</span>"
	}
	var valueHTML string
	ref, hasFk := fks[table.String()][col.Name]
	stringValue := *DbValueToString(cellData, col.Type)
	if hasFk {
		valueHTML = fmt.Sprintf("<a href='%s?%s=", ref.Table, ref.Col)
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
