package sdv

import (
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"
	"net/url"
)

var db string
var driver string
var cachingEnabled bool
var database schema.Database

func RunServer(driverInfo string, dbConn string, port int, listenOn string, live bool) {
	db = dbConn
	driver = driverInfo
	cachingEnabled = !live

	SetupTemplate()

	reader := getDbReader(driver, db)
	err := reader.CheckConnection()
	if err != nil {
		log.Println(err)
		panic("connection check failed")
	}

	log.Print("Reading schema, this may take a while...")
	database, err = reader.ReadSchema()
	if err != nil {
		fmt.Println("Error reading schema", err)
		// todo: send 500 error to client
		return
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

func loggingHandler(nextHandler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, false)
		if err != nil {
			log.Println("couldn't dump request")
			panic(err)
		}
		log.Printf("Request from '%v'\n%s", r.RemoteAddr, dump)
		nextHandler(w, r)
	}
}

func handler(resp http.ResponseWriter, req *http.Request) {
	Licensing()

	reader := getDbReader(driver, db)

	layoutData = pageTemplateModel{
		Db:          db,
		Title:       About.ProductName,
		About:       About,
		Copyright:   CopyrightText(),
		LicenseText: LicenseText(),
		Timestamp:   time.Now().String(),
	}

	if !cachingEnabled {
		SetupTemplate()
		log.Print("Re-reading schema, this may take a while...")
		var err error
		database, err = reader.ReadSchema()
		if err != nil {
			fmt.Println("Error reading schema", err)
			// todo: send 500 error to client
			return
		}
	}

	// todo: proper url routing
	folders := strings.Split(req.URL.Path, "/")
	switch folders[1] {
	case "tables":
		requestedTable := parseTableName(folders[2])
		if requestedTable.Name == "" { // google bot strips paths it seems, was causing a crash
			http.Redirect(resp, req, "/", http.StatusFound)
			return
		}
		table := database.FindTable(&requestedTable)
		params := ParseParams(req.URL.Query())

		// trail handling
		trailCookie, _ := req.Cookie("table-trail")
		var exists = false
		var trail []string
		if trailCookie == nil {
			trailCookie=&http.Cookie{Name: "table-trail", Value: ""}
			trail = append(trail, table.String())
		}else{
			trail = strings.Split(trailCookie.Value, ",")
			for _, x := range trail {
				if x == table.String() {
					exists = true
				}
			}
			if !exists {
				trail = append(trail, table.String())
			}
		}
		trailCookie.Value = strings.Join(trail, ",")
		http.SetCookie(resp, trailCookie)

		var err error
		err = showTable(resp, reader, table, params)
		if err != nil {
			fmt.Println("error converting rows querystring value to int: ", err)
			return
		}
	default:
		showTableList(resp, database)
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

type tableParams struct {
	rowLimit int
	cardView bool
	filter   schema.RowFilter
}

// todo: more robust separation of query param keys
const rowLimitKey = "_rowLimit" // this should be reasonably safe from clashes with column names
const cardViewKey = "_cardView"

func  ParseParams(raw url.Values)(params tableParams){
	params = tableParams{
		filter: schema.RowFilter(raw),
	}
	rowLimitString := raw.Get(rowLimitKey)
	if rowLimitString != "" {
		var err error
		params.rowLimit, err = strconv.Atoi(rowLimitString)
		// exclude from column filters
		raw.Del(rowLimitKey)
		if err != nil {
			fmt.Println("error converting rows querystring value to int: ", err)
			return
		}
	}
	cardViewString := raw.Get(cardViewKey)
	if cardViewString != "" {
		params.cardView = cardViewString == "true"
		raw.Del(cardViewKey)
	}
	params.filter = schema.RowFilter(raw)
	return
}
