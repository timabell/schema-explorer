package render

import (
	"bitbucket.org/timabell/sql-data-viewer/about"
	"bitbucket.org/timabell/sql-data-viewer/params"
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"bitbucket.org/timabell/sql-data-viewer/trail"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
)

type PageTemplateModel struct {
	Title       string
	Db          string
	About       about.AboutType
	Copyright   string
	LicenseText string
	Timestamp   string
}

type tablesViewModel struct {
	LayoutData PageTemplateModel
	Database   schema.Database
	rowLimit   int
	cardView   bool
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

type FieldFilterList []FieldFilter

func (filterList FieldFilterList) AsQueryString() template.URL {
	var parts []string
	for _, part := range filterList {
		// todo: support multiple values correctly
		parts = append(parts, fmt.Sprintf("%s=%s", part.Field, strings.Join(part.Values, ",")))
	}
	return template.URL(strings.Join(parts, "&"))
}

type trailViewModel struct {
	LayoutData PageTemplateModel
	Diagram    diagramViewModel
	Trail      *trail.TrailLog
}
type dataViewModel struct {
	LayoutData PageTemplateModel
	Table      schema.Table
	Query      FieldFilterList
	RowLimit   int
	Rows       []cells
	Diagram    diagramViewModel
	CardView   bool
}

var tablesTemplate *template.Template
var tableTemplate *template.Template
var tableTrailTemplate *template.Template

// Make minus available in templates to be able to convert len to slice index
// https://stackoverflow.com/a/24838050/10245
var funcMap = template.FuncMap{
	"minus": minus,
}

func minus(x, y int) int {
	return x - y
}

func SetupTemplate() {
	templates, err := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/layout.tmpl")).ParseGlob("templates/_*.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tablesTemplate, err = template.Must(templates.Clone()).ParseGlob("templates/tables.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tableTrailTemplate, err = template.Must(templates.Clone()).ParseGlob("templates/table-trail.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tableTemplate, err = template.Must(templates.Clone()).ParseGlob("templates/table.tmpl")
	if err != nil {
		log.Fatal(err)
	}
}

func ShowTableList(resp http.ResponseWriter, database schema.Database, layoutData PageTemplateModel) {
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

func ShowTable(resp http.ResponseWriter, dbReader reader.DbReader, table *schema.Table, params params.TableParams, layoutData PageTemplateModel) error {
	fieldFilter := make(FieldFilterList, 0)
	if len(params.Filter) > 0 {
		fieldKeys := make([]string, 0)
		for field, _ := range params.Filter {
			fieldKeys = append(fieldKeys, field)
		}
		sort.Strings(fieldKeys)
		for _, field := range fieldKeys {
			fieldFilter = append(fieldFilter, FieldFilter{Field: field, Values: params.Filter[field]})
		}
	}

	rowsData, err := reader.GetRows(dbReader, table, params)
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
		RowLimit:   params.RowLimit,
		Rows:       rows,
		Diagram:    diagramViewModel{Tables: diagramTables, TableLinks: tableLinks},
		CardView:   params.CardView,
	}

	viewModel.LayoutData.Title = fmt.Sprintf("%s | %s", table.String(), viewModel.LayoutData.Title)

	err = tableTemplate.ExecuteTemplate(resp, "layout", viewModel)
	if err != nil {
		log.Print("template execution error ", err)
	}

	return nil
}

func ShowTableTrail(resp http.ResponseWriter, database schema.Database, trailInfo *trail.TrailLog, layoutData PageTemplateModel) error {
	log.Printf("%#v", trailInfo)

	var diagramTables []*schema.Table
	for _, x := range trailInfo.Tables {
		tableStub := schema.TableFromString(x)
		table := database.FindTable(&tableStub)
		if table != nil { // this will happen if schema has changed since cookie was set
			diagramTables = append(diagramTables, table)
		}
	}

	var tableLinks []fkViewModel
	for _, tableFks := range database.Fks {
		tableLinks = append(tableLinks, fkViewModel{Source: *tableFks.SourceTable, Destination: *tableFks.DestinationTable})
	}
	// todo: Filter fks

	viewModel := trailViewModel{
		LayoutData: layoutData,
		Diagram:    diagramViewModel{Tables: diagramTables, TableLinks: tableLinks},
		Trail:      trailInfo,
	}

	viewModel.LayoutData.Title = fmt.Sprintf("%s | %s", "trail", viewModel.LayoutData.Title)

	err := tableTrailTemplate.ExecuteTemplate(resp, "layout", viewModel)
	if err != nil {
		log.Print("template execution error", err)
		panic(err)
	}

	return nil
}

func buildRow(rowData reader.RowData, table *schema.Table) cells {
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

// Groups fks by source table, adds table name for each followed by links for each inbound fk for that table
func buildInwardCell(inboundFks []*schema.Fk, rowData []interface{}, cols []*schema.Column) string {
	groupedFks := groupFksByTable(inboundFks)

	// note.... for table, fks := range groupedFks { ... is an unstable sort, don't do it this way! https://stackoverflow.com/a/23332089/10245
	// stable sort:
	// get list of tables in map
	var keys schema.TableList
	for table, _ := range groupedFks {
		keys = append(keys, table)
	}
	// sort list of tables (requires TableList to implement sort interface)
	sort.Sort(keys)
	// iterate through sorted list of keys, using that to find entry in map
	parentHTML := ""
	for _, table := range keys {
		fks := groupedFks[table]
		parentHTML = parentHTML + template.HTMLEscapeString(table.String()) + ":&nbsp;"
		for _, fk := range fks {
			parentHTML = parentHTML + buildInwardLink(fk, rowData)
		}
		parentHTML = parentHTML + " "
	}
	return parentHTML
}

type groupedFkMap map[*schema.Table][]*schema.Fk

func groupFksByTable(inboundFks []*schema.Fk) groupedFkMap {
	var groupedFks = make(map[*schema.Table][]*schema.Fk, 0)
	for _, fk := range inboundFks {
		if _, exists := groupedFks[fk.SourceTable]; !exists {
			groupedFks[fk.SourceTable] = make([]*schema.Fk, 0)
		}
		groupedFks[fk.SourceTable] = append(groupedFks[fk.SourceTable], fk)
	}
	return groupedFks
}

func buildInwardLink(fk *schema.Fk, rowData reader.RowData) string {
	// todo: handle non-string primary key
	linkHTML := fmt.Sprintf(
		"<a href='%s?%s=",
		template.URLQueryEscaper(fk.SourceTable),
		template.URLQueryEscaper(fk.SourceColumns))
	// todo: handle compound keys
	if len(fk.DestinationColumns) > 1 {
		log.Print("unsupported: compound key. " + fk.String())
		return ""
	}
	destinationColumnIndex, _ := fk.DestinationTable.FindColumn(fk.DestinationColumns[0].Name)
	if destinationColumnIndex < 0 {
		log.Print(fk)
		log.Printf("%#v", fk.DestinationTable)
		log.Panic("Destination column not found in referenced table")
	}
	colData := rowData[destinationColumnIndex]
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
		"&_rowLimit=100#data' class='parentFk'>%s</a>&nbsp;",
		template.HTMLEscaper(fk.SourceColumns.String()))
	return linkHTML
}

func buildCell(col *schema.Column, cellData interface{}) string {
	if cellData == nil {
		return "<span class='null'>[null]</span>"
	}
	var valueHTML string
	hasFk := col.Fk != nil
	stringValue := *reader.DbValueToString(cellData, col.Type)
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
		valueHTML = valueHTML + "&_rowLimit=100#data' class='fk'>"
	}
	valueHTML = valueHTML + template.HTMLEscapeString(stringValue)
	if hasFk {
		valueHTML = valueHTML + "</a>"
	}
	return valueHTML
}
