package render

import (
	"github.com/timabell/schema-explorer/about"
	"github.com/timabell/schema-explorer/driver_interface"
	"github.com/timabell/schema-explorer/drivers"
	"github.com/timabell/schema-explorer/params"
	"github.com/timabell/schema-explorer/reader"
	"github.com/timabell/schema-explorer/resources"
	"github.com/timabell/schema-explorer/schema"
	"github.com/timabell/schema-explorer/trail"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type PageTemplateModel struct {
	Title             string
	ConnectionName    string
	About             about.AboutType
	Copyright         string
	LicenseText       string
	Timestamp         string
	CanSwitchDatabase bool
	DbReady           bool
	DatabaseName      string
}

type driverSelectionViewModel struct {
	LayoutData PageTemplateModel
	Drivers    []*drivers.Driver
}

type driverSetupViewModel struct {
	LayoutData PageTemplateModel
	Driver     *drivers.Driver
	Errors     string
}

type databaseListViewModel struct {
	LayoutData   PageTemplateModel
	DatabaseList []string
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
	LayoutData PageTemplateModel
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
	LayoutData        PageTemplateModel
	Database          *schema.Database
	Table             *schema.Table
	TableParams       *params.TableParams
	Rows              []cells
	TotalRowCount     int
	FilteredRowCount  int
	DisplayedRowCount int
	HasPrevPage       bool
	HasNextPage       bool
	Diagram           diagramViewModel
}
type tableAnalysisDataViewModel struct {
	LayoutData PageTemplateModel
	Database   *schema.Database
	Table      *schema.Table
	Analysis   []schema.ColumnAnalysis
}

var databasesTemplate *template.Template
var tablesTemplate *template.Template
var tableTemplate *template.Template
var tableDataTemplate *template.Template
var tableAnalysisTemplate *template.Template
var tableTrailTemplate *template.Template
var selectDriverTemplate *template.Template
var setupDriverTemplate *template.Template

// global copy for reverse url lookups
// use empty string for databaseName if not selected, irrelevant or not supported
// pairs is route values as per gorilla mux's Get()
type UrlBuilder func(routeName string, database string, pairs []string) *url.URL

var urlBuilder UrlBuilder

func SetRouterFinder(u UrlBuilder) {
	urlBuilder = u
}

// Make minus available in templates to be able to convert len to slice index
// https://stackoverflow.com/a/24838050/10245
var funcMap = template.FuncMap{
	"minus":           minus,
	"DbValueToString": reader.DbValueToString,
	"isNil":           isNil,
}

func minus(x, y int) int {
	return x - y
}

// because https://stackoverflow.com/questions/54578243/how-can-i-prevent-non-nil-values-triggering-a-golang-template-if-nil-block/54579872#54579872
func isNil(value interface{}) bool {
	return value == nil
}

func SetupTemplates() {
	templates, err := template.Must(template.New("").Funcs(funcMap).ParseGlob(resources.TemplateFolder + "/layout.tmpl")).ParseGlob(resources.TemplateFolder + "/_*.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	databasesTemplate, err = template.Must(templates.Clone()).ParseGlob(resources.TemplateFolder + "/databases.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tablesTemplate, err = template.Must(templates.Clone()).ParseGlob(resources.TemplateFolder + "/tables.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tableTrailTemplate, err = template.Must(templates.Clone()).ParseGlob(resources.TemplateFolder + "/table-trail.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tableTemplate, err = template.Must(templates.Clone()).ParseGlob(resources.TemplateFolder + "/table.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tableDataTemplate, err = template.Must(templates.Clone()).ParseGlob(resources.TemplateFolder + "/table-data.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	tableAnalysisTemplate, err = template.Must(templates.Clone()).ParseGlob(resources.TemplateFolder + "/table-analysis.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	selectDriverTemplate, err = template.Must(templates.Clone()).ParseGlob(resources.TemplateFolder + "/select-driver.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	setupDriverTemplate, err = template.Must(templates.Clone()).ParseGlob(resources.TemplateFolder + "/setup-driver.tmpl")
	if err != nil {
		log.Fatal(err)
	}
}

func ShowSelectDriver(resp http.ResponseWriter, layoutData PageTemplateModel) {
	model := driverSelectionViewModel{
		LayoutData: layoutData,
		Drivers:    getDrivers(),
	}
	err := selectDriverTemplate.ExecuteTemplate(resp, "layout", model)
	if err != nil {
		log.Fatal(err)
	}
}

func getDrivers() []*drivers.Driver {
	var driverList []*drivers.Driver
	// stable sort:
	var keys []string
	for name, _ := range drivers.Drivers {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		driverList = append(driverList, drivers.Drivers[name])
	}
	return driverList
}

func ShowSetupDriver(resp http.ResponseWriter, layoutData PageTemplateModel, driver string, errors string) {
	model := driverSetupViewModel{
		LayoutData: layoutData,
		Driver:     drivers.Drivers[driver],
		Errors:     errors,
	}
	err := setupDriverTemplate.ExecuteTemplate(resp, "layout", model)
	if err != nil {
		log.Fatal(err)
	}
}

func ShowDatabaseList(resp http.ResponseWriter, layoutData PageTemplateModel, databaseList []string) {
	model := databaseListViewModel{
		LayoutData:   layoutData,
		DatabaseList: databaseList,
	}
	err := databasesTemplate.ExecuteTemplate(resp, "layout", model)
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
		Diagram:    diagramViewModel{Tables: database.Tables, TableLinks: tableLinks, LayoutData: layoutData},
	}

	err := tablesTemplate.ExecuteTemplate(resp, "layout", model)
	if err != nil {
		log.Fatal(err)
	}
}

func ShowTable(resp http.ResponseWriter, dbReader driver_interface.DbReader, database *schema.Database, table *schema.Table, tableParams *params.TableParams, layoutData PageTemplateModel, dataOnly bool) error {
	unfilteredParams := tableParams.ClearPaging()
	filteredRowCount, err := dbReader.GetRowCount(database.Name, table, &unfilteredParams)
	totalRowCount, err := dbReader.GetRowCount(database.Name, table, &params.TableParams{})
	rowsData, peekFinder, err := reader.GetRows(dbReader, database.Name, table, tableParams)
	if err != nil {
		return err
	}

	rows := []cells{}
	for _, rowData := range rowsData {
		row := buildRow(database.Name, rowData, peekFinder, table)
		rows = append(rows, row)
	}

	diagramTables := []*schema.Table{table}
	var tableLinks []fkViewModel
	if !dataOnly {
		for _, tableFks := range table.Fks {
			diagramTables = append(diagramTables, tableFks.DestinationTable)
			tableLinks = append(tableLinks, fkViewModel{Source: *tableFks.SourceTable, Destination: *tableFks.DestinationTable})
		}
		for _, inboundFks := range table.InboundFks {
			diagramTables = append(diagramTables, inboundFks.SourceTable)
			tableLinks = append(tableLinks, fkViewModel{Source: *inboundFks.SourceTable, Destination: *inboundFks.DestinationTable})
		}
	}

	viewModel := tableDataViewModel{
		LayoutData:        layoutData,
		Database:          database,
		Table:             table,
		TableParams:       tableParams,
		Rows:              rows,
		TotalRowCount:     totalRowCount,
		FilteredRowCount:  filteredRowCount,
		DisplayedRowCount: len(rows),
		HasPrevPage:       tableParams.SkipRows > 0,
		HasNextPage:       tableParams.ToRow() < filteredRowCount,
		Diagram:           diagramViewModel{Tables: diagramTables, TableLinks: tableLinks, LayoutData: layoutData},
	}

	viewModel.LayoutData.Title = fmt.Sprintf("%s | %s", table.String(), viewModel.LayoutData.Title)

	if dataOnly {
		err = tableDataTemplate.ExecuteTemplate(resp, "layout", viewModel)
	} else {
		err = tableTemplate.ExecuteTemplate(resp, "layout", viewModel)
	}
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
		Diagram:    diagramViewModel{Tables: diagramTables, TableLinks: tableLinks, LayoutData: layoutData},
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

func ShowTableAnalysis(resp http.ResponseWriter, dbReader driver_interface.DbReader, database *schema.Database, table *schema.Table, layoutData PageTemplateModel) error {
	analysis, err := dbReader.GetAnalysis(database.Name, table)
	if err != nil {
		return err
	}

	viewModel := tableAnalysisDataViewModel{
		LayoutData: layoutData,
		Database:   database,
		Table:      table,
		Analysis:   analysis,
	}

	viewModel.LayoutData.Title = fmt.Sprintf("%s analysis | %s", table.String(), viewModel.LayoutData.Title)

	err = tableAnalysisTemplate.ExecuteTemplate(resp, "layout", viewModel)
	if err != nil {
		log.Print("template execution error ", err)
	}

	return nil
}

func buildRow(databaseName string, rowData reader.RowData, peekFinder *driver_interface.PeekLookup, table *schema.Table) cells {
	row := cells{}
	for colIndex, col := range table.Columns {
		cellData := rowData[colIndex]
		valueHTML := buildCell(databaseName, col, cellData, rowData, peekFinder)
		row = append(row, template.HTML(valueHTML))
	}
	parentHTML := buildInwardCell(databaseName, table.InboundFks, rowData, peekFinder)
	row = append(row, template.HTML(parentHTML))
	return row
}

// Groups fks by source table, adds table name for each followed by links for each inbound fk for that table
func buildInwardCell(databaseName string, inboundFks []*schema.Fk, rowData []interface{}, peekFinder *driver_interface.PeekLookup) string {
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
		parentHTML = parentHTML + "</span> "
		for _, fk := range fks {
			parentHTML = parentHTML + buildInwardLink(databaseName, fk, rowData, peekFinder) + " "
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

func buildInwardLink(databaseName string, fk *schema.Fk, rowData reader.RowData, peekFinder *driver_interface.PeekLookup) string {
	var queryData []string
	for ix, fkCol := range fk.SourceColumns {
		destinationCol := fk.DestinationColumns[ix]
		fkCellData := rowData[destinationCol.Position]
		escapedName := template.HTMLEscapeString(template.URLQueryEscaper(fkCol.String()))
		escapedValue := template.HTMLEscapeString(template.URLQueryEscaper(reader.DbValueToString(fkCellData, fkCol.Type)))
		queryData = append(queryData, fmt.Sprintf("%s=%s", escapedName, escapedValue))
	}
	var joinedQueryData = strings.Join(queryData, "&")
	suffix := "&_rowLimit=100#data"
	inboundPeekIndex := peekFinder.FindInbound(fk)
	rowCount := rowData[inboundPeekIndex].(int64)
	if rowCount > 0 {
		var pairs = []string{"tableName", fk.SourceTable.String()}
		fkUrl := urlBuilder("route-database-tables", databaseName, pairs)
		return fmt.Sprintf("<a href='%s?%s%s' class='parent-fk-link'>%s - %d rows</a>", fkUrl, joinedQueryData, suffix, fk.SourceColumns, rowCount)
	} else {
		return fmt.Sprintf("%s - %d rows", fk.SourceColumns, rowCount)
	}
}

func buildCell(databaseName string, col *schema.Column, cellData interface{}, rowData reader.RowData, peekFinder *driver_interface.PeekLookup) string {
	if cellData == nil {
		return "<span class='null bare-value'>[null]</span>"
	}
	stringValue := *reader.DbValueToString(cellData, col.Type)
	if col.Fks != nil {
		multiFk := len(col.Fks) > 1
		if multiFk {
			// if multiple fks on this col, put val first
			valueHTML := "<span class='compound-value'>" + template.HTMLEscapeString(stringValue) + "</span> "
			for _, fk := range col.Fks {
				displayText := fmt.Sprintf("%s(%s)", fk.DestinationTable, fk.DestinationColumns)
				valueHTML = valueHTML + buildCompleteFkHref(databaseName, fk, multiFk, rowData, displayText, peekFinder)
			}
			return valueHTML
		} else {
			// otherwise put it in the link
			fk := col.Fks[0]
			displayText := stringValue
			return buildCompleteFkHref(databaseName, fk, multiFk, rowData, displayText, peekFinder)
		}
	} else {
		return "<span class='bare-value'>" + template.HTMLEscapeString(stringValue) + "</span> "
	}
}

func buildCompleteFkHref(databaseName string, fk *schema.Fk, multiFk bool, rowData reader.RowData, displayText string, peekFinder *driver_interface.PeekLookup) string {
	cssClass := buildFkCss(fk, multiFk)
	joinedQueryData := buildQueryData(fk, rowData)

	peekHtml := ""
	for _, peekColumn := range fk.DestinationTable.PeekColumns {
		peekIndex := peekFinder.Find(fk, peekColumn)
		val := rowData[peekIndex]
		var peekString string
		if val == nil {
			// This could be null because the current table's fk col is null in which case we get no value (so nothing to peek)
			// or it could be because the value we are peeking at is null, which isn't very interesting to see so we'll not show it.
			peekString = ""
		} else {
			peekString = template.HTMLEscapeString(*reader.DbValueToString(val, peekColumn.Type))
		}
		peekHtml = peekHtml + fmt.Sprintf("<span class='peek'>%s</span>", peekString)
	}

	return buildFkHref(databaseName, fk.DestinationTable, joinedQueryData, cssClass, displayText, peekHtml)
}

func buildFkCss(fk *schema.Fk, multiFkCol bool) string {
	typeString := "single"
	if multiFkCol {
		typeString = "multi"
	}
	if len(fk.SourceColumns) > 1 {
		return "fk compound " + typeString
	} else {
		return "fk " + typeString
	}
}

func buildFkHref(databaseName string, table *schema.Table, query string, cssClass string, displayText string, peekHtml string) string {
	suffix := "&_rowLimit=100#data"
	var fkUrl *url.URL
	var pairs = []string{"tableName", table.String()}
	fkUrl = urlBuilder("route-database-tables", databaseName, pairs)
	return fmt.Sprintf("<a href='%s?%s%s' class='%s'>%s%s</a> ", fkUrl, query, suffix, cssClass, template.HTMLEscapeString(displayText), peekHtml)
}

func buildQueryData(fk *schema.Fk, rowData reader.RowData) string {
	var queryData []string
	for ix, fkCol := range fk.DestinationColumns {
		sourceCol := fk.SourceColumns[ix]
		fkCellData := rowData[sourceCol.Position]
		fkStringValue := *reader.DbValueToString(fkCellData, fkCol.Type)
		escapedValue := template.URLQueryEscaper(fkStringValue)
		escapedValue = template.HTMLEscapeString(escapedValue)
		queryData = append(queryData, fmt.Sprintf("%s=%s", fkCol, escapedValue))
	}
	var joinedQueryData = strings.Join(queryData, "&")
	return joinedQueryData
}
