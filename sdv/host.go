package sdv

import (
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"net/http/httputil"
)

var db string
var driver string
var liveTemplates bool

func RunServer(driverInfo string, dbConn string, port int, listenOn string, live bool) {
	db = dbConn
	driver = driverInfo
	liveTemplates = live

	SetupTemplate()

	reader := getDbReader(driver, db)
	err := reader.CheckConnection()
	if err != nil {
		log.Println(err)
		panic("connection check failed")
	}

	serve(handler, port, listenOn)
}

func serve(handler func(http.ResponseWriter, *http.Request), port int, listenOn string) {
	http.HandleFunc("/", loggingHandler(handler))
	http.Handle("/static/", http.FileServer(http.Dir("")))
	listenOnHostPort := fmt.Sprintf("%s:%d", listenOn, port) // e.g. localhost:8080 or 0.0.0.0:80
	log.Printf("Starting server on http://%s/ - Press Ctrl-C to kill server.\n", listenOnHostPort)
	log.Fatal(http.ListenAndServe(listenOnHostPort, nil))
	log.Panic("http.ListenAndServe didn't block")
}

func loggingHandler(nextHandler func(w http.ResponseWriter, r *http.Request)) (func(w http.ResponseWriter, r *http.Request)){
	return func(w http.ResponseWriter, r *http.Request){
		dump, err := httputil.DumpRequest(r, false)
		if err!=nil{
			log.Println("couldn't dump request")
			panic(err)
		}
		log.Printf("Request from '%v'\n%s", r.RemoteAddr, dump)
		nextHandler(w,r)
	}
}

func handler(resp http.ResponseWriter, req *http.Request) {
	Licensing()

	if liveTemplates {
		SetupTemplate()
	}

	reader := getDbReader(driver, db)

	layoutData = pageTemplateModel{
		Db:        db,
		Title:     "Sql Data Viewer",
		About:     About,
		Copyright: CopyrightText(),
		LicenseText: LicenseText(),
		Timestamp: time.Now().String(),
	}

	// todo: proper url routing
	folders := strings.Split(req.URL.Path, "/")
	switch folders[1] {
	case "tables":
		// todo: check not missing table name
		table := parseTableName(folders[2])
		var query = req.URL.Query()
		var rowLimit int
		var err error
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
		var rowFilter = schema.RowFilter(query)
		err = showTable(resp, reader, table, rowFilter, rowLimit)
		if err != nil {
			fmt.Println("error converting rows querystring value to int: ", err)
			return
		}
	default:
		tables, err := reader.GetTables()
		if err != nil {
			fmt.Println("error getting table list", err)
			return
		}

		allKeys, err := reader.AllFks()
		if err != nil {
			panic(err)
		}
		showTableList(resp, tables, allKeys)
	}
}

// Split dot-separated name into schema + table name
func parseTableName(tableFullname string) (table schema.Table) {
	if strings.Contains(tableFullname, ".") {
		splitName := strings.SplitN(tableFullname, ".", 2)
		table = schema.Table{Schema: splitName[0], Name: splitName[1]}
	} else {
		table = schema.Table{Schema: "", Name: tableFullname}
	}
	return
}
