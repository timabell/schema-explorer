package sdv

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sql-data-viewer/schema"
	"strings"
)

type pageTemplateModel struct {
	Title     string
	Db        string
	Version   string
	Copyright string
	Timestamp string
}

type tablesViewModel struct {
	LayoutData pageTemplateModel
	Tables     []schema.Table
}

type cells []template.HTML

type dataViewModel struct {
	LayoutData pageTemplateModel
	Table      schema.Table
	Query      string
	RowLimit   int
	Cols       []schema.Column
	Rows       []cells
}

var tmpl *template.Template
var layoutData pageTemplateModel

func SetupTemplate() {
	tmpl = template.Must(template.New("template").Parse(headerHTML))
	tmpl = template.Must(tmpl.Parse(footerHTML))
	tmpl = template.Must(tmpl.Parse(tablesHTML))
	tmpl = template.Must(tmpl.Parse(dataHTML))

}

func showTableList(resp http.ResponseWriter, tables []schema.Table) {
	model := tablesViewModel{
		LayoutData: layoutData,
		Tables:     tables,
	}

	err := tmpl.ExecuteTemplate(resp, "tables", model)
	if err != nil {
		log.Fatal(err)
	}
}

func showTable(resp http.ResponseWriter, reader dbReader, table schema.Table, query schema.RowFilter, rowLimit int) error {
	var formattedQuery string
	if len(query) > 0 {
		formattedQuery = fmt.Sprintf("%s", query)
	}

	viewModel := dataViewModel{
		LayoutData: layoutData,
		Table:      table,
		Query:      formattedQuery,
		RowLimit:   rowLimit,
		Cols:       []schema.Column{},
		Rows:       []cells{},
	}

	fks, err := reader.AllFks()
	if err != nil {
		log.Println("error getting fks", err)
		panic("error getting fks")
		// todo: send 500 error to client
		return err
	}

	inwardFks := table.FindParents(fks)

	rows, err := reader.GetRows(query, table, rowLimit)
	if rows == nil {
		panic("GetRows() returned nil")
	}
	defer rows.Close()

	cols, err := reader.GetColumns(table)
	if err != nil{
		panic(err)
	}
	viewModel.Cols = cols

	// http://stackoverflow.com/a/23507765/10245 - getting ad-hoc column data
	rowData := make([]interface{}, len(cols))
	rowDataPointers := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		rowDataPointers[i] = &rowData[i]
	}
	for rows.Next() {

		err := rows.Scan(rowDataPointers...)
		if err != nil {
			log.Println("error reading row data", err)
			return err
		}
		row := buildRow(cols, rowData, fks, table, inwardFks)
		viewModel.Rows = append(viewModel.Rows, row)
	}
	err = tmpl.ExecuteTemplate(resp, "data", viewModel)
	if err != nil {
		log.Print("template execution error", err)
		panic(err)
	}
	return err
}

func buildRow(cols []schema.Column, rowData []interface{}, fks schema.GlobalFkList, table schema.Table, inwardFks schema.GlobalFkList) cells {
	row := cells{}
	for colIndex, col := range cols {
		colData := rowData[colIndex]
		valueHTML := buildCell(fks, table, col, colData)
		row = append(row, template.HTML(valueHTML))
	}
	parentHTML := buildInwardCell(inwardFks, rowData, cols)
	row = append(row, template.HTML(parentHTML))
	return row
}

func buildInwardCell(inwardFks schema.GlobalFkList, rowData []interface{}, cols []schema.Column) string {
	// todo: stable sort order http://stackoverflow.com/questions/23330781/sort-golang-map-values-by-keys
	// todo: pre-calculate fk info so this isn't repeated for every row
	parentHTML := ""
	for parentTable, parentFks := range inwardFks {
		for parentCol, ref := range parentFks {
			parentHTML = parentHTML + buildInwardLink(parentTable, parentCol, rowData, cols, ref)
		}
	}
	return parentHTML
}

func buildInwardLink(parentTable string, parentCol schema.Column, rowData []interface{}, cols []schema.Column, ref schema.Ref) string {
	linkHTML := fmt.Sprintf(
		"<a href='%s?%s=",
		template.URLQueryEscaper(parentTable),
		template.URLQueryEscaper(parentCol))
	// todo: handle non-string primary key
	// todo: handle compound primary key
	colData := rowData[indexOfCol(cols, string(ref.Col.Name))]
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
		"&_rowLimit=100' class='parentFk'>%s.%s</a>&nbsp;",
		template.HTMLEscaper(parentTable),
		template.HTMLEscaper(parentCol))
	return linkHTML
}

func buildCell(fks schema.GlobalFkList, table schema.Table, col schema.Column, colData interface{}) string {
	var valueHTML string
	ref, refExists := fks[table.String()][col]
	if refExists && colData != nil {
		valueHTML = fmt.Sprintf("<a href='%s?%s=", ref.Table, ref.Col)
		switch {
		case col.Type == "integer":
			// todo: url-escape as well
			valueHTML = valueHTML + template.HTMLEscapeString(fmt.Sprintf("%d", colData))
		case strings.Contains(col.Type,"varchar"):
			// todo: sql-quotes here are a hack pending switching to parameterized sql
			valueHTML = valueHTML + "%27" + template.HTMLEscapeString(fmt.Sprintf("%s", colData)) + "%27"
		default:
			valueHTML = valueHTML + template.HTMLEscapeString(fmt.Sprintf("%v", colData))
		}
		valueHTML = valueHTML + "' class='fk'>"
	}
	log.Println(col.Type, colData)
	switch {
	case colData == nil:
		valueHTML = valueHTML + "<span class='null'>[null]</span>"
	case col.Type == "integer":
		valueHTML = valueHTML + template.HTMLEscapeString(fmt.Sprintf("%d", colData))
	case col.Type == "float":
		valueHTML = valueHTML + template.HTMLEscapeString(fmt.Sprintf("%f", colData))
	case strings.Contains(col.Type,"varchar"):
		valueHTML = valueHTML + template.HTMLEscapeString(fmt.Sprintf("%s", colData))
	case strings.Contains(col.Type,"TEXT"):
		// https://stackoverflow.com/a/18615786/10245
		bytes := colData.([]uint8)
		log.Println(bytes)
		valueHTML = valueHTML + template.HTMLEscapeString(fmt.Sprintf("%s", string(bytes)))
	default:
		valueHTML = valueHTML + template.HTMLEscapeString(fmt.Sprintf("%v", colData))
	}
	if refExists && colData != nil {
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

const headerHTML = `
{{define "header"}}
<!DOCTYPE html>
<html lang='en'>
<head>
	<title>{{.Title}}</title>
	<style type='text/css'>
		body { background-color: #f9fff9; margin: 1em; }
		.null { color: #999; }
		#connected { font-style: italic; }
		.config-value { background-color: #eee; }
		footer { color: #666; text-align: right; font-size: smaller; }
		footer a { color: #66c; }
		th.references { font-style: italic }
	</style>
</head>
<body>
<h1>Sql Data Viewer</h1>
<nav><a href='/'>Table list</a></nav>
{{end}}
`
const footerHTML = `
{{define "footer"}}
<footer>
	Generated by Sql Data Viewer v{{.Version}} at {{.Timestamp}}<br/>
	{{.Copyright}}
</footer>
</body>
</html>
{{end}}
`

const tablesHTML = `
{{define "tables"}}
{{template "header" .LayoutData}}
<table border=1>
{{range .Tables}}
	<tr><td><a href='tables/{{.}}?_rowLimit=100'>{{.}}</a></td></tr>
{{end}}
</table>
{{template "footer" .LayoutData}}
{{end}}
`

const dataHTML = `
{{define "data"}}
{{template "header" .LayoutData}}
	<h2>Table {{.Table.Name}}</h2>
	{{ if .Query }}
		<p class='filtered'>Filtered - {{.Query}} &nbsp; &nbsp; <a href="?_rowLimit={{.RowLimit}}">Clear filter</a><p>
	{{end}}
	{{ if .RowLimit }}
		<p class='filtered'>First {{.RowLimit}} rows<p>
	{{end}}
	<table border=1>
		<tr>
		{{ range .Cols }}
			<th title='type: {{.Type}}'>{{.Name}}</th>
		{{end}}
		<th class='references'>referenced by</th>
		</tr>
		{{ range .Rows }}
		<tr>
		{{ range . }}
			<td>{{.}}</td>
		{{end}}
		</tr>
		{{end}}
	</table>
{{template "footer" .LayoutData}}
{{end}}
`
