/*
Sql Data Viewer, Copyright Tim Abell 2015-17
All rights reserved.

A tool for browsing the data in any rdbms databse
through a series of generated html pages.

Provides navigation between tables via the foreign keys
defined in the database's schema.
*/

package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"sql-data-viewer/sdv"
)

const version = "0.2"

type pageTemplateModel struct {
	Title     string
	Db        string
	Version   string
	Timestamp string
}

type tablesViewModel struct {
	LayoutData pageTemplateModel
	Tables     []sdv.TableName
}

type cells []template.HTML

type dataViewModel struct {
	LayoutData pageTemplateModel
	TableName  sdv.TableName
	Query      string
	RowLimit   int
	Cols       []string
	Rows       []cells
}

var db string
var tmpl *template.Template
var layoutData pageTemplateModel

// var pageTemplate template.Template

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("missing argument: path to sqlite database file")
	}
	db = os.Args[1]

	port := 8080
	if len(os.Args) > 2 {
		portString := os.Args[2]
		var err error
		port, err = strconv.Atoi(portString)
		if err != nil {
			log.Fatal("invalid port ", portString)
		}
	}

	log.Printf("Sql Data Viewer v%s; Copyright 2015-17 Tim Abell <sdv@timwise.co.uk>", version)
	log.Printf("## This pre-release software will expire on: %s, contact sdv@timwise.co.uk for a license. ##", expiry)
	licensing()

	tmpl = template.Must(template.New("template").Parse(headerHTML))
	tmpl = template.Must(tmpl.Parse(footerHTML))
	tmpl = template.Must(tmpl.Parse(tablesHTML))
	tmpl = template.Must(tmpl.Parse(dataHTML))

	log.Printf("Connecting to db: %s\n", db)
	sdv.Serve(handler, port)
}

func handler(resp http.ResponseWriter, req *http.Request) {
	licensing()
	log.Printf("req: %s\n", req.URL)

	dbc, err := sql.Open("sqlite3", db)
	if err != nil {
		log.Println("connection error", err)
		return
	}
	defer dbc.Close()

	layoutData = pageTemplateModel{
		Db:        db,
		Title:     "Sql Data Viewer",
		Version:   version,
		Timestamp: time.Now().String(),
	}

	folders := strings.Split(req.URL.Path, "/")
	switch folders[1] {
	case "tables":
		// todo: check not missing table name
		table := sdv.TableName(folders[2])
		var query = req.URL.Query()
		var rowLimit int
		// todo: more robust separation of query param keys
		const rowLimitKey = "_rowLimit" // this should be reasonably safe from clashes with column names
		rowLimitString := query.Get(rowLimitKey)
		if rowLimitString != "" {
			rowLimit, err = strconv.Atoi(rowLimitString)
			// exclude from column filters
			query.Del(rowLimitKey)
			if err != nil {
				fmt.Println("error converting rows querystring value to int: ", err)
				return
			}
		}
		var rowFilter = sdv.RowFilter(query)
		showTable(resp, dbc, table, rowFilter, rowLimit)
	default:
		showTableList(resp, dbc)
	}
	if err != nil {
		log.Fatal(err) //todo: make non-fatal
	}
}

func showTableList(resp http.ResponseWriter, dbc *sql.DB) {
	tables, err := sdv.GetTables(dbc)
	if err != nil {
		fmt.Println("error getting table list", err)
		return
	}

	model := tablesViewModel{
		LayoutData: layoutData,
		Tables:     tables,
	}

	err = tmpl.ExecuteTemplate(resp, "tables", model)
	if err != nil {
		log.Fatal(err)
	}
}

func showTable(resp http.ResponseWriter, dbc *sql.DB, table sdv.TableName, query sdv.RowFilter, rowLimit int) {
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

	fks := sdv.AllFks(dbc)

	// find all the of the fks that point at this table
	inwardFks := table.FindParents(fks)
	fmt.Println("found: ", inwardFks)

	sql := "select * from " + string(table)

	if len(query) > 0 {
		sql = sql + " where "
		clauses := make([]string, 0, len(query))
		for k, v := range query {
			clauses = append(clauses, k+" = "+v[0])
		}
		sql = sql + strings.Join(clauses, " and ")
	}

	if rowLimit > 0 {
		sql = sql + " limit " + strconv.Itoa(rowLimit)
	}

	log.Println(sql)

	rows, err := dbc.Query(sql)
	if err != nil {
		log.Println("select error", err)
		return
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		log.Println("error getting column names", err)
		// todo: send 500 error to client
		return
	}

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
			return
		}
		for colIndex, col := range cols {
			colData := rowData[colIndex]
			var valueHTML string
			ref, refExists := fks[table][sdv.ColumnName(col)]
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

func licensing() {
	if time.Now().After(expiry) {
		log.Panic("Expired trial, contact sdv@timwise.co.uk to obtain a license")
	}
}

// roughly 3 months from when this is released into the wild
var expiry = time.Date(2017, time.July, 1, 0, 0, 0, 0, time.UTC)

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
	Sql Data Viewer &copy; 2015 <a href='mailto:sdv@timwise.co.uk?subject=Sql Data Viewer'>Tim Abell</a>
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
