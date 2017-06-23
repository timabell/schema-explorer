package sdv

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"strings"
	"strconv"
)

var db string
var driver string

func RunServer(driverInfo string, dbConn string, port int){
	db = dbConn
	driver = driverInfo

	SetupTemplate()

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

	var model dbInterface
	switch driver {
	case "mssql":
		model = NewMssql(db)
	case "sqlite":
		model = NewSqlite(db)
	}


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
		tables, err := GetTables(dbc)
		if err != nil {
			fmt.Println("error getting table list", err)
			return
		}

		tables := model.
		showTableList(resp, tables)
	}
	if err != nil {
		log.Fatal(err) //todo: make non-fatal
	}
}

