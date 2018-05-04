package sdv

import (
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
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
	case "table-trail": // todo: this should require http post
		if len(folders) > 2 && folders[2] == "clear" {
			clearTrailCookie(resp)
			http.Redirect(resp, req, "/table-trail", http.StatusFound)
			return
		}
		trail := readTrail(req)
		err := showTableTrail(resp, database, trail)
		if err != nil {
			fmt.Println("error rendering trail: ", err)
			return
		}
	case "tables":
		requestedTable := parseTableName(folders[2])
		if requestedTable.Name == "" { // google bot strips paths it seems, was causing a crash
			http.Redirect(resp, req, "/", http.StatusFound)
			return
		}
		table := database.FindTable(&requestedTable)
		params := ParseParams(req.URL.Query())

		trail := readTrail(req)
		trail.addTable(table)
		trail.setCookie(resp)

		var err error
		err = showTable(resp, reader, table, params)
		if err != nil {
			fmt.Println("error rendering table: ", err)
			return
		}
	default:
		showTableList(resp, database)
	}
}

type trail struct {
	tables []string
}

const trailCookieName = "table-trail"

func readTrail(req *http.Request) *trail {
	trailCookie, _ := req.Cookie(trailCookieName)
	log.Printf("%#v", trailCookie)
	if trailCookie != nil {
		return &trail{strings.Split(trailCookie.Value, ",")}
	}
	return &trail{}
}

func (trailInfo *trail) addTable(table *schema.Table) {
	var exists = false
	for _, x := range trailInfo.tables {
		if x == table.String() {
			exists = true
		}
	}
	if !exists {
		trailInfo.tables = append(trailInfo.tables, table.String())
	}
}
func (trailInfo *trail) setCookie(resp http.ResponseWriter) {
	value := strings.Join(trailInfo.tables, ",")
	trailCookie := &http.Cookie{Name: trailCookieName, Value: value, Path: "/"}
	http.SetCookie(resp, trailCookie)
}
func clearTrailCookie(resp http.ResponseWriter) {
	trailCookie := &http.Cookie{Name: trailCookieName, Value: "", Path: "/", Expires: time.Now().Add(-10000)}
	http.SetCookie(resp, trailCookie)
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

func ParseParams(raw url.Values) (params tableParams) {
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
