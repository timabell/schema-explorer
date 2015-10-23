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
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// reference to a field in another table, part of a foreign key
type ref struct {
	table string
	col   string
}

var db string

func main() {
	db = os.Args[1]
	log.Println("Sql Data Viewer; Copyright 2015 Tim Abell <tim@timwise.co.uk>")
	licensing()
	log.Printf("## This pre-release software will expire on: %s, contact tim@timwise.co.uk for a license. ##", expiryRFC822)
	log.Printf("Connecting to db: %s\n", db)
	http.HandleFunc("/", handler)
	log.Println("Listening on http://localhost:8080/")
	log.Println("Press Ctrl-C to kill server")
	http.ListenAndServe(":8080", nil)
}

func handler(resp http.ResponseWriter, req *http.Request) {
	licensing()
	log.Printf("req: %s\n", req.URL)
	fmt.Fprintln(resp, htmlHeader)
	dbc, err := sql.Open("sqlite3", db)
	if err != nil {
		log.Println("connection error", err)
		return
	}
	defer dbc.Close()
	fmt.Fprintf(resp, "<p id='connected'>Connected to <span class='config-value'>%s</span></p>\n", db)
	folders := strings.Split(req.URL.Path, "/")
	switch folders[1] {
	case "tables":
		// todo: check not missing table name
		table := folders[2]
		query := req.URL.Query()
		showTable(resp, dbc, table, query)
	default:
		showTableList(resp, dbc)
	}
	fmt.Fprintln(resp, htmlFooter)
}

func showTableList(resp http.ResponseWriter, dbc *sql.DB) {
	tables, err := getTables(dbc)
	if err != nil {
		fmt.Println("error getting table list", err)
		return
	}
	fmt.Fprintln(resp, `<table border=1>`)
	for _, table := range tables {
		fmt.Fprintf(resp, "<tr><td><a href='tables/%s'>%s</a></td></tr>\n", table, table) // todo: html encode table name
	}
	fmt.Fprintln(resp, "</table>")
}

func showTable(resp http.ResponseWriter, dbc *sql.DB, table string, query map[string][]string) {
	fks := fks(dbc, table)
	fmt.Fprintf(resp, "<h2>Table %s</h2>\n", table)
	if len(query) > 0 {
		fmt.Fprintf(resp, "<p class='filtered'>Filtered - %s<p>", query)
	}
	sql := "select * from " + table
	if len(query) > 0 {
		sql = sql + " where "
		clauses := make([]string, 0, len(query))
		for k, v := range query {
			clauses = append(clauses, k+" = "+v[0])
		}
		sql = sql + strings.Join(clauses, " and ")
	}
	log.Println(sql)
	rows, err := dbc.Query(sql)
	if err != nil {
		log.Println("select error", err)
		return
	}
	defer rows.Close()
	fmt.Fprintln(resp, `<table border=1>`)
	cols, err := rows.Columns()
	if err != nil {
		log.Println("error getting column names", err)
		return
	}
	fmt.Fprintln(resp, "<tr>")
	for _, col := range cols {
		fmt.Fprintf(resp, "<th>%s</th>\n", col)
	}
	fmt.Fprintln(resp, "</tr>")
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
			return
		}
		fmt.Fprintln(resp, "\t<tr>")
		for colIndex, col := range cols {
			colData := rowData[colIndex]
			fmt.Fprint(resp, "\t\t<td>")
			ref, refExists := fks[col]
			if refExists && colData != nil {
				fmt.Fprintf(resp, "<a href='%s?%s=%d' class='fk'>", ref.table, ref.col, colData)
			}
			switch colData.(type) {
			case int64:
				fmt.Fprintf(resp, "%d", colData)
			case nil:
				fmt.Fprint(resp, "<span class='null'>[null]</span>")
			default:
				fmt.Fprintf(resp, "%s", colData)
			}
			if refExists && colData != nil {
				fmt.Fprintf(resp, "</a>")
			}
			fmt.Fprintln(resp, "</td>")
		}
		fmt.Fprintln(resp, "\t</tr>")
	}
	fmt.Fprintln(resp, "</table>")
}

func fks(dbc *sql.DB, table string) (fks map[string]ref) {
	rows, err := dbc.Query("PRAGMA foreign_key_list('" + table + "');")
	if err != nil {
		log.Println("select error", err)
		return
	}
	defer rows.Close()
	//fmt.Fprintln(resp, `<h3>foreign keys</h3> <ul>`)
	fks = make(map[string]ref)
	for rows.Next() {
		var id, seq int
		var parentTable, from, to, onUpdate, onDelete, match string
		rows.Scan(&id, &seq, &parentTable, &from, &to, &onUpdate, &onDelete, &match)
		//thisFk := Fk{FromCol: from, ToTable parentTable, col: to}
		thisRef := ref{col: to, table: parentTable}
		fks[from] = thisRef
		//fmt.Fprintf(resp, "\t<li>key: %s references %s.%s</li>\n", from, parentTable, to)
	}
	//fmt.Fprintln(resp, "</ul>")
	return
}

func getTables(dbc *sql.DB) (tables []string, err error) {
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

const htmlHeader = `<!DOCTYPE html>
<html lang='en'>
<head>
	<title>Sql Data Viewer</title>
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
<nav><a href='/'>Table list</a></nav>
`

const htmlFooter = `<footer>
	Generated by Sql Data Viewer<br/>
	Sql Data Viewer &copy; 2015 <a href='mailto:tim@timwise.co.uk?subject=Sql Data Viewer'>Tim Abell</a>
</footer>
</body>
</html>`
