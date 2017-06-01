package sdv

import (
	"fmt"
	"log"
	"net/http"
	"database/sql"
	"time"
	"strings"
	"strconv"
)

var db string

func RunServer(dbConn string, port int){
	db = dbConn

	SetupTemplate()

	// todo: test connection
	log.Printf("Connecting to db: %s\n", db)

	serve(handler, port)
}

func serve(handler func(http.ResponseWriter, *http.Request), port int) {
	// todo: use multiple handlers properly
	http.HandleFunc("/", handler)
	listenOn := fmt.Sprintf("localhost:%d", port)
	log.Printf("Starting server on http://%s/ - Press Ctrl-C to kill server.\n", listenOn)
	log.Fatal(http.ListenAndServe(listenOn, nil))
	log.Panic("http.ListenAndServe didn't block")
}

func handler(resp http.ResponseWriter, req *http.Request) {
	Licensing()
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
		Version:   Version,
		Copyright: CopyrightText(),
		Timestamp: time.Now().String(),
	}

	folders := strings.Split(req.URL.Path, "/")
	switch folders[1] {
	case "tables":
		// todo: check not missing table name
		table := TableName(folders[2])
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
		var rowFilter = RowFilter(query)
		showTable(resp, dbc, table, rowFilter, rowLimit)
	default:
		showTableList(resp, dbc)
	}
	if err != nil {
		log.Fatal(err) //todo: make non-fatal
	}
}

