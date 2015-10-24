/*
Sql Data Viewer, Copyright Tim Abell 2015
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
)

const version = "0.2"

// filtering of results with column name / value(s) pairs,
// matches type of url.Values so can pass straight through
type rowFilter map[string][]string

// reference to a field in another table, part of a foreign key
type ref struct {
	table string
	col   string
}

type pageTemplateModel struct {
	Title     string
	Db        string
	Version   string
	Timestamp string
}

type tablesViewModel struct {
	LayoutData pageTemplateModel
	Tables     tableList
}

type cells []template.HTML

type dataViewModel struct {
	LayoutData pageTemplateModel
	TableName  string
	Query      string
	RowLimit   int
	Cols       []string
	Rows       []cells
}

type tableList []string

var db string
var tmpl *template.Template
var layoutData pageTemplateModel

// var pageTemplate template.Template

func main() {
	db = os.Args[1]
	log.Printf("Sql Data Viewer v%s; Copyright 2015 Tim Abell <tim@timwise.co.uk>", version)
	licensing()

	tmpl = template.Must(template.New("template").Parse(headerHtml))
	tmpl = template.Must(tmpl.Parse(footerHtml))
	tmpl = template.Must(tmpl.Parse(tablesHtml))
	tmpl = template.Must(tmpl.Parse(dataHtml))

	log.Printf("## This pre-release software will expire on: %s, contact tim@timwise.co.uk for a license. ##", expiryRFC822)
	log.Printf("Connecting to db: %s\n", db)
	// todo: use multiple handlers properly
	http.HandleFunc("/", handler)
	log.Println("Starting server on http://localhost:8080/ - Press Ctrl-C to kill server.")
	log.Fatal(http.ListenAndServe(":8080", nil))
	log.Panic("http.ListenAndServe didn't block")
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
		table := folders[2]
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
		var rowFilter = rowFilter(query)
		showTable(resp, dbc, table, rowFilter, rowLimit)
	default:
		showTableList(resp, dbc)
	}
	if err != nil {
		log.Fatal(err) //todo: make non-fatal
	}
}

func showTableList(resp http.ResponseWriter, dbc *sql.DB) {
	tables, err := getTables(dbc)
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

func showTable(resp http.ResponseWriter, dbc *sql.DB, table string, query rowFilter, rowLimit int) {
	var formattedQuery string
	if len(query) > 0 {
		formattedQuery = fmt.Sprintf("%s", query)
	}

	model := dataViewModel{
		LayoutData: layoutData,
		TableName:  table,
		Query:      formattedQuery,
		RowLimit:   rowLimit,
		Cols:       []string{},
		Rows:       []cells{},
	}

	fks := fks(dbc, table)

	sql := "select * from " + table

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
		model.Cols = append(model.Cols, col)
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
			var valueHtml string
			ref, refExists := fks[col]
			if refExists && colData != nil {
				valueHtml = fmt.Sprintf("<a href='%s?%s=%d' class='fk'>", ref.table, ref.col, colData)
			}
			switch colData.(type) {
			case int64:
				valueHtml = valueHtml + template.HTMLEscapeString(fmt.Sprintf("%d", colData))
			case float64:
				valueHtml = valueHtml + template.HTMLEscapeString(fmt.Sprintf("%f", colData))
			case nil:
				valueHtml = valueHtml + "<span class='null'>[null]</span>"
			default:
				valueHtml = valueHtml + template.HTMLEscapeString(fmt.Sprintf("%s", colData))
			}
			if refExists && colData != nil {
				valueHtml = valueHtml + "</a>"
			}
			row = append(row, template.HTML(valueHtml))
		}
		model.Rows = append(model.Rows, row)
	}

	err = tmpl.ExecuteTemplate(resp, "data", model)
	if err != nil {
		log.Print("template exexution error", err)
	}
}

func fks(dbc *sql.DB, table string) (fks map[string]ref) {
	rows, err := dbc.Query("PRAGMA foreign_key_list('" + table + "');")
	if err != nil {
		log.Println("select error", err)
		return
	}
	defer rows.Close()
	fks = make(map[string]ref)
	for rows.Next() {
		var id, seq int
		var parentTable, from, to, onUpdate, onDelete, match string
		rows.Scan(&id, &seq, &parentTable, &from, &to, &onUpdate, &onDelete, &match)
		thisRef := ref{col: to, table: parentTable}
		fks[from] = thisRef
	}
	return
}

func getTables(dbc *sql.DB) (tables tableList, err error) {
	rows, err := dbc.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, name)
	}
	return tables, nil
}

func licensing() {
	expiry, _ := time.Parse(time.RFC822, expiryRFC822)
	if time.Now().After(expiry) {
		log.Panic("expired trial, contact Tim Abell <tim@timwise.co.uk> to obtain a license")
	}
}

const expiryRFC822 = "16 Jan 16 00:00 UTC" // 3 months from when this was written

const headerHtml = `
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
	</style>
</head>
<body>
<h1>Sql Data Viewer</h1>
<p id='connected'>Connected to <span class='config-value'>{{.Db}}</span></p>
<nav><a href='/'>Table list</a></nav>
{{end}}
`
const footerHtml = `
{{define "footer"}}
<footer>
	Generated by Sql Data Viewer v{{.Version}} at {{.Timestamp}}<br/>
	Sql Data Viewer &copy; 2015 <a href='mailto:tim@timwise.co.uk?subject=Sql Data Viewer'>Tim Abell</a>
</footer>
</body>
</html>
{{end}}
`

const tablesHtml = `
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

const dataHtml = `
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
