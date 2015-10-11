package main

import (
	"fmt"
	"net/http"
	"os"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var db string

func main() {
	db = os.Args[1]
	fmt.Printf("Connecting to db: %s\n", db)
	http.HandleFunc("/", handler)
	fmt.Println("Listening on http://localhost:8080/")
	fmt.Println("Press Ctrl-C to kill server")
	http.ListenAndServe(":8080", nil)
}

func handler(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(resp, "<h1>bonjour!</h1>\n<p>Hello soapie</p>")
	dbc, err :=sql.Open("sqlite3", db)
	if (err != nil) {
		fmt.Println("connection error", err)
		return
	}
	defer dbc.Close()
	fmt.Fprintf(resp, "<p>Connected to %s</p>", db)
	tables, err := getTables(dbc)
	if (err != nil) {
		fmt.Println("error getting table list", err)
		return
	}
	for _, table := range tables {
		showTable(resp, dbc, table)
	}
}

func showTable(resp http.ResponseWriter, dbc *sql.DB, table string) {
	rows, err := dbc.Query("select * from " + table)
	if (err != nil) {
		fmt.Println("select error", err)
		return
	}
	defer rows.Close()
	fmt.Fprintf(resp, "<h2>Table %s</h2><table border=1>", table)
	cols, err := rows.Columns()
	if (err != nil) {
		fmt.Println("error getting column names", err)
		return
	}
	fmt.Fprintf(resp, "<tr>")
	for _, col := range cols {
		fmt.Fprintf(resp, "<th>%s</th>", col)
	}
	fmt.Fprintf(resp, "</tr>")
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		fmt.Fprintf(resp, "<tr><td>%d</td><td>%s</td></tr>", id, name)
	}
	fmt.Fprintf(resp, "</table>")
	fks(resp, dbc, table)
}

func fks(resp http.ResponseWriter, dbc *sql.DB, table string) {
	rows, err := dbc.Query("PRAGMA foreign_key_list('" + table + "');")
	if (err != nil) {
		fmt.Println("select error", err)
		return
	}
	defer rows.Close()
	fmt.Fprintf(resp, "<ul>")
	for rows.Next() {
		var id, seq int
		var parentTable, from, to, on_update, on_delete string
		rows.Scan(&id, &seq, &parentTable, &from, &to, &on_update, &on_delete)
		fmt.Fprintf(resp, "<li>key: %s references %s.%s</li>", from, parentTable, to)
	}
	fmt.Fprintf(resp, "</ul>")
}

func getTables(dbc *sql.DB) (tables []string, err error){
	rows, err := dbc.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if (err != nil) {
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
