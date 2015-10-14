package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
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
	fmt.Fprint(resp, "<html><head><style type='text/css'>.null { color: #999; }</style></head><body><h1>bonjour!</h1>\n<p>Hello soapie</p>")
	dbc, err := sql.Open("sqlite3", db)
	if err != nil {
		fmt.Println("connection error", err)
		return
	}
	defer dbc.Close()
	fmt.Fprintf(resp, "<p>Connected to %s</p>", db)
	tables, err := getTables(dbc)
	if err != nil {
		fmt.Println("error getting table list", err)
		return
	}
	for _, table := range tables {
		showTable(resp, dbc, table)
	}
	fmt.Fprint(resp, "</body></html>")
}

func showTable(resp http.ResponseWriter, dbc *sql.DB, table string) {
	rows, err := dbc.Query("select * from " + table)
	if err != nil {
		fmt.Println("select error", err)
		return
	}
	defer rows.Close()
	fmt.Fprintf(resp, "<h2>Table %s</h2><table border=1>", table)
	cols, err := rows.Columns()
	if err != nil {
		fmt.Println("error getting column names", err)
		return
	}
	fmt.Fprintf(resp, "<tr>")
	for _, col := range cols {
		fmt.Fprintf(resp, "<th>%s</th>", col)
	}
	fmt.Fprintf(resp, "</tr>")
	// http://stackoverflow.com/a/23507765/10245 - getting ad-hoc column data
	rowData := make([]interface{}, len(cols))
	rowDataPointers := make([]interface{}, len(cols))
	for i := 0; i < len(cols); i++ {
		rowDataPointers[i] = &rowData[i]
	}
	for rows.Next() {
		err := rows.Scan(rowDataPointers...)
		if err != nil {
			fmt.Println("error reading row data", err)
			return
		}
		fmt.Fprintf(resp, "<tr>")
		for colIndex := range cols {
			colData := rowData[colIndex]
			fmt.Fprint(resp, "<td>")
			switch colData.(type) {
			case int64:
				fmt.Fprintf(resp, "%d", colData)
			case nil:
				fmt.Fprint(resp, "<span class='null'>[null]</span>", colData)
			default:
				fmt.Fprintf(resp, "%s", colData)
			}
			fmt.Fprint(resp, "</td>")
		}
		fmt.Fprintf(resp, "</tr>")
	}
	fmt.Fprintf(resp, "</table>")
	fks(resp, dbc, table)
}

func fks(resp http.ResponseWriter, dbc *sql.DB, table string) {
	rows, err := dbc.Query("PRAGMA foreign_key_list('" + table + "');")
	if err != nil {
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
