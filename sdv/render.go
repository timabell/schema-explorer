package sdv

import (
	"html/template"
	"net/http"
	//"database/sql"
	//"fmt"
	"log"
	//"strings"
	//"strconv"
	"fmt"
	"sql-data-viewer/schema"
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
	Tables     []schema.TableName
}

type cells []template.HTML

type dataViewModel struct {
	LayoutData pageTemplateModel
	TableName  schema.TableName
	Query      string
	RowLimit   int
	Cols       []string
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

func showTableList(resp http.ResponseWriter, tables []schema.TableName) {
	model := tablesViewModel{
		LayoutData: layoutData,
		Tables:     tables,
	}

	err := tmpl.ExecuteTemplate(resp, "tables", model)
	if err != nil {
		log.Fatal(err)
	}
}

func showTable(resp http.ResponseWriter, reader dbReader, table schema.TableName, query schema.RowFilter, rowLimit int) error {
	var formattedQuery string
	if len(query) > 0 {
		formattedQuery = fmt.Sprintf("%s", query)
	}

	viewModel := dataViewModel{
		LayoutData: layoutData,
		TableName:  table,
		Query:      formattedQuery,
		RowLimit:   rowLimit,
		Cols:       []string{},
		Rows:       []cells{},
	}

	log.Println("getting fks")
	fks, err := reader.AllFks()
	if err != nil {
		log.Println("error getting fks", err)
		// todo: send 500 error to client
		return err
	}
	log.Println("got fks")

	log.Println("finding parents")
	// find all the of the fks that point at this table
	inwardFks := table.FindParents(fks)
	fmt.Println("found: ", inwardFks)

	log.Println("getting data...")
	rows, err := reader.GetRows(query, table, rowLimit)
	defer rows.Close()

	log.Println("getting columns...")
	// works on sqlite, fails with azure mssql:
	//2017/06/25 12:51:35 http: panic serving 127.0.0.1:42410: runtime error: invalid memory address or nil pointer dereference
	// todo: push col parsing down into custom db reader, provide mssql-proof variant
	cols, err := rows.Columns()
	if err != nil {
		log.Println("error getting column names", err)
		// todo: send 500 error to client
		return err
	}
	log.Println("got columns")

	for _, col := range cols {
		viewModel.Cols = append(viewModel.Cols, col)
	}

	// http://stackoverflow.com/a/23507765/10245 - getting ad-hoc column data
	rowData := make([]interface{}, len(cols))
	rowDataPointers := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		rowDataPointers[i] = &rowData[i]
	}
	for rows.Next() {
		row := cells{}

		err := rows.Scan(rowDataPointers...)
		if err != nil {
			log.Println("error reading row data", err)
			return err
		}
		for colIndex, col := range cols {
			colData := rowData[colIndex]
			var valueHTML string
			ref, refExists := fks[table][schema.ColumnName(col)]
			if refExists && colData != nil {
				valueHTML = fmt.Sprintf("<a href='%s?%s=%d' class='fk'>", ref.Table, ref.Col, colData)
			}
			switch colData.(type) {
			case int64:
				valueHTML = valueHTML + template.HTMLEscapeString(fmt.Sprintf("%d", colData))
			case float64:
				valueHTML = valueHTML + template.HTMLEscapeString(fmt.Sprintf("%f", colData))
			case nil:
				valueHTML = valueHTML + "<span class='null'>[null]</span>"
			default:
				valueHTML = valueHTML + template.HTMLEscapeString(fmt.Sprintf("%s", colData))
			}
			if refExists && colData != nil {
				valueHTML = valueHTML + "</a>"
			}
			row = append(row, template.HTML(valueHTML))
		}
		parentHTML := ""
		// todo: factor out row limit, move to a cookie perhaps
		// todo: stable sort order http://stackoverflow.com/questions/23330781/sort-golang-map-values-by-keys
		// todo: pre-calculate fk info so this isn't repeated for every row
		for parentTable, parentFks := range inwardFks {
			for parentCol, ref := range parentFks {
				parentHTML = parentHTML + fmt.Sprintf(
					"<a href='%s?%s=%d&_rowLimit=100' class='parentFk'>%s.%s</a>&nbsp;",
					template.URLQueryEscaper(parentTable),
					template.URLQueryEscaper(parentCol),
					rowData[indexOf(cols, string(ref.Col))],
					template.HTMLEscaper(parentTable),
					template.HTMLEscaper(parentCol))
			}
		}
		row = append(row, template.HTML(parentHTML))
		viewModel.Rows = append(viewModel.Rows, row)
	}

	err = tmpl.ExecuteTemplate(resp, "data", viewModel)
	if err != nil {
		log.Print("template exexution error", err)
	}
	return err
}

func indexOf(slice []string, value string) (index int) {
	var curValue string
	for index, curValue = range slice {
		if curValue == value {
			return
		}
	}
	log.Panic(value, " not found in slice")
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
<p id='connected'>Connected to <span class='config-value'>{{.Db}}</span></p>
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
	<h2>Table {{.TableName}}</h2>
	{{ if .Query }}
		<p class='filtered'>Filtered - {{.Query}}<p>
	{{end}}
	{{ if .RowLimit }}
		<p class='filtered'>First {{.RowLimit}} rows<p>
	{{end}}
	<table border=1>
		<tr>
		{{ range .Cols }}
			<th>{{.}}</th>
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
