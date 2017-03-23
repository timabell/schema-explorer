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
)

const version = "0.2"

// alias to make it clear when we're dealing with table names
type tableName string

// alias to make it clear when we're dealing with column names
type columnName string

// filtering of results with column name / value(s) pairs,
// matches type of url.Values so can pass straight through
type rowFilter map[string][]string

// reference to a field in another table, part of a foreign key
type ref struct {
	table tableName  // target table for the fk
	col   columnName // target col for the fk
}

// list of foreign keys, the column in the current table that the fk is defined on
type fkList map[columnName]ref

// for each table in the database, the list of fks defined on that table
type globalFkList map[tableName]fkList

type pageTemplateModel struct {
	Title     string
	Db        string
	Version   string
	Timestamp string
}

type tablesViewModel struct {
	LayoutData pageTemplateModel
	Tables     []tableName
}

type cells []template.HTML

type dataViewModel struct {
	LayoutData pageTemplateModel
	TableName  tableName
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
	// todo: use multiple handlers properly
	http.HandleFunc("/", handler)
	listenOn := fmt.Sprintf("localhost:%d", port)
	log.Printf("Starting server on http://%s/ - Press Ctrl-C to kill server.\n", listenOn)
	log.Fatal(http.ListenAndServe(listenOn, nil))
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
		table := tableName(folders[2])
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

func showTable(resp http.ResponseWriter, dbc *sql.DB, table tableName, query rowFilter, rowLimit int) {
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

	fks := allFks(dbc)

	// find all the of the fks that point at this table
	inwardFks := findParents(fks, table)
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
			var valueHTML string
			ref, refExists := fks[table][columnName(col)]
			if refExists && colData != nil {
				valueHTML = fmt.Sprintf("<a href='%s?%s=%d' class='fk'>", ref.table, ref.col, colData)
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
					rowData[indexOf(cols, string(ref.col))],
					template.HTMLEscaper(parentTable),
					template.HTMLEscaper(parentCol))
			}
		}
		row = append(row, template.HTML(parentHTML))
		model.Rows = append(model.Rows, row)
	}

	err = tmpl.ExecuteTemplate(resp, "data", model)
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

// filter the fk list down to keys that reference the "child" table
func findParents(fks globalFkList, child tableName) (parents globalFkList) {
	parents = globalFkList{}
	for srcTable, tableFks := range fks {
		newFkList := fkList{}
		for srcCol, ref := range tableFks {
			if ref.table == child {
				// match; copy into new list
				newFkList[srcCol] = ref
				parents[srcTable] = newFkList
			}
		}
	}
	return
}

func allFks(dbc *sql.DB) (allFks globalFkList) {
	tables, err := getTables(dbc)
	if err != nil {
		fmt.Println("error getting table list while building global fk list", err)
		return
	}
	allFks = globalFkList{}
	for _, table := range tables {
		allFks[table] = fks(dbc, table)
	}
	return
}

func fks(dbc *sql.DB, table tableName) (fks fkList) {
	rows, err := dbc.Query("PRAGMA foreign_key_list('" + string(table) + "');")
	if err != nil {
		log.Println("select error", err)
		return
	}
	defer rows.Close()
	fks = fkList{}
	for rows.Next() {
		var id, seq int
		var parentTable, from, to, onUpdate, onDelete, match string
		rows.Scan(&id, &seq, &parentTable, &from, &to, &onUpdate, &onDelete, &match)
		thisRef := ref{col: columnName(to), table: tableName(parentTable)}
		fks[columnName(from)] = thisRef
	}
	return
}

func getTables(dbc *sql.DB) (tables []tableName, err error) {
	rows, err := dbc.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, tableName(name))
	}
	return tables, nil
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
