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
	Title          string
	ConnectionName string
	Db             string
	About          about.AboutType
	Copyright      string
	LicenseText    string
	Timestamp      string
}

type tableListViewModel struct {
	LayoutData PageTemplateModel
	Database   *schema.Database
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

type trailViewModel struct {
	LayoutData PageTemplateModel
	Diagram    diagramViewModel
	Trail      *trail.TrailLog
}
type tableDataViewModel struct {
	LayoutData  PageTemplateModel
	Database    *schema.Database
	Table       *schema.Table
	TableParams *params.TableParams
	Rows        []cells
	Diagram     diagramViewModel
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

func ShowTableList(resp http.ResponseWriter, database *schema.Database, layoutData PageTemplateModel) {
	var tableLinks []fkViewModel
	for _, fk := range database.Fks {
		tableLinks = append(tableLinks, fkViewModel{Source: *fk.SourceTable, Destination: *fk.DestinationTable})
	}

	model := tableListViewModel{
		LayoutData: layoutData,
		Database:   database,
		Diagram:    diagramViewModel{Tables: database.Tables, TableLinks: tableLinks},
	}

	err := tablesTemplate.ExecuteTemplate(resp, "layout", model)
	if err != nil {
		log.Fatal(err)
	}
}

func ShowTable(resp http.ResponseWriter, dbReader reader.DbReader, database *schema.Database, table *schema.Table, tableParams *params.TableParams, layoutData PageTemplateModel) error {
	rowsData, err := reader.GetRows(dbReader, table, tableParams)
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

	viewModel := tableDataViewModel{
		LayoutData:  layoutData,
		Database:    database,
		Table:       table,
		TableParams: tableParams,
		Rows:        rows,
		Diagram:     diagramViewModel{Tables: diagramTables, TableLinks: tableLinks},
	}

	viewModel.LayoutData.Title = fmt.Sprintf("%s | %s", table.String(), viewModel.LayoutData.Title)

	err = tableTemplate.ExecuteTemplate(resp, "layout", viewModel)
	if err != nil {
		log.Print("template execution error ", err)
	}

	return nil
}

func ShowTableTrail(resp http.ResponseWriter, database *schema.Database, trailInfo *trail.TrailLog, layoutData PageTemplateModel) error {
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
		valueHTML := buildCell(col, cellData, rowData)
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
	parentHTML := "<span class='parent-fks'>"
	for _, table := range keys {
		fks := groupedFks[table]
		parentHTML = parentHTML + "<span class='parent-fk-table'>"
		parentHTML = parentHTML + template.HTMLEscapeString(table.String()) + ":"
		parentHTML = parentHTML + "</span>"
		for _, fk := range fks {
			parentHTML = parentHTML + buildInwardLink(fk, rowData) + " "
		}
		parentHTML = parentHTML + "<br/>"
	}
	parentHTML = parentHTML + "</span>"
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
	var queryData []string
	for ix, fkCol := range fk.SourceColumns {
		destinationCol := fk.DestinationColumns[ix]
		fkCellData := rowData[destinationCol.Index]
		escapedName := template.HTMLEscapeString(template.URLQueryEscaper(fkCol.String()))
		escapedValue := template.HTMLEscapeString(template.URLQueryEscaper(reader.DbValueToString(fkCellData, fkCol.Type)))
		queryData = append(queryData, fmt.Sprintf("%s=%s", escapedName, escapedValue))
	}
	var joinedQueryData = strings.Join(queryData, "&")
	suffix := "&_rowLimit=100#data"
	linkHTML := fmt.Sprintf("<a href='%s?%s%s' class='parent-fk-link'>%s</a>", fk.SourceTable, joinedQueryData, suffix, fk.SourceColumns)
	return linkHTML
}

func buildCell(col *schema.Column, cellData interface{}, rowData reader.RowData) string {
	if cellData == nil {
		return "<span class='null bare-value'>[null]</span>"
	}
	var valueHTML string
	hasFk := col.Fks != nil
	stringValue := *reader.DbValueToString(cellData, col.Type)
	if hasFk {
		// todo: possible performance optimisation to save lots of lookups within a loop for the majority case of single column fks
		//if len(col.Fks.SourceColumns) ==1{
		//	valueHTML = fmt.Sprintf("<a href='%s?%s=", col.Fks.DestinationTable, col.Fks.DestinationColumns[0].Name)
		//  valueHTML = fmt.Sprintf("%s=", col.Fks.DestinationTable, col.Fks.DestinationColumns[0].Name)
		//  valueHTML = valueHTML + template.HTMLEscapeString(stringValue)
		//}else{
		suffix := "&_rowLimit=100#data"
		if len(col.Fks) > 1 {
			valueHTML = "<span class='compound-value'>" + template.HTMLEscapeString(stringValue) + "</span> "
			for _, fk := range col.Fks {
				var cssClass string
				if len(fk.SourceColumns) > 1 {
					cssClass = "fk compound multi"
				} else {
					cssClass = "fk multi"
				}
				joinedQueryData := buildQueryData(fk, rowData)
				valueHTML = valueHTML + fmt.Sprintf("<a href='%s?%s%s' class='%s'>%s(%s)</a> ", fk.DestinationTable, joinedQueryData, suffix, cssClass, fk.DestinationTable, fk.DestinationColumns)
			}
		} else {
			fk := col.Fks[0]
			var cssClass string
			if len(fk.SourceColumns) > 1 {
				cssClass = "fk compound single"
			} else {
				cssClass = "fk single"
			}
			joinedQueryData := buildQueryData(fk, rowData)
			valueHTML = valueHTML + fmt.Sprintf("<a href='%s?%s%s' class='%s'>%s</a> ", fk.DestinationTable, joinedQueryData, suffix, cssClass, template.HTMLEscapeString(stringValue))
		}
	} else {
		valueHTML = "<span class='bare-value'>" + template.HTMLEscapeString(stringValue) + "</span> "
	}
	return valueHTML
}

func buildQueryData(fk *schema.Fk, rowData reader.RowData) string {
	var queryData []string
	for ix, fkCol := range fk.DestinationColumns {
		sourceCol := fk.SourceColumns[ix]
		fkCellData := rowData[sourceCol.Index]
		fkStringValue := *reader.DbValueToString(fkCellData, fkCol.Type)
		escapedValue := template.URLQueryEscaper(fkStringValue)
		escapedValue = template.HTMLEscapeString(escapedValue)
		queryData = append(queryData, fmt.Sprintf("%s=%s", fkCol, escapedValue))
	}
	var joinedQueryData = strings.Join(queryData, "&")
	return joinedQueryData
}
