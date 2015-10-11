package main

import (
	"fmt"
	"net/http"
	"os"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var db string

func handler(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(resp, "<h1>bonjour!</h1>\n<p>Hello soapie</p>")
	dbc, err :=sql.Open("sqlite3", db)
	if (err != nil) {
		fmt.Println("connection error", err)
		return
	}
	fmt.Fprintf(resp, "<p>Connected to %s</p>", db)
	rows, err := dbc.Query("select * from foo;")
	if (err != nil) {
		fmt.Println("select error", err)
		return
	}
	defer rows.Close()
	fmt.Fprintf(resp, "<table border=1>")
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		fmt.Fprintf(resp, "<tr><td>%d</td><td>%s</td></tr>", id, name)
	}
	fmt.Fprintf(resp, "</table>")
	defer dbc.Close()
}

func main() {
	db = os.Args[1]
	fmt.Printf("Connecting to db: %s\n", db)
	http.HandleFunc("/", handler)
	fmt.Println("Listening on http://localhost:8080/")
	fmt.Println("Press Ctrl-C to kill server")
	http.ListenAndServe(":8080", nil)
}
