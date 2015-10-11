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
	}
	defer dbc.Close()
	fmt.Fprintf(resp, "<p>Connected to %s</p>", db)
}

func main() {
	db = os.Args[1]
	fmt.Printf("Connecting to db: %s\n", db)
	http.HandleFunc("/", handler)
	fmt.Println("Listening on http://localhost:8080/")
	fmt.Println("Press Ctrl-C to kill server")
	http.ListenAndServe(":8080", nil)
}
