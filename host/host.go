package host

import (
	"bitbucket.org/timabell/sql-data-viewer/about"
	"bitbucket.org/timabell/sql-data-viewer/licensing"
	"bitbucket.org/timabell/sql-data-viewer/params"
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"bitbucket.org/timabell/sql-data-viewer/render"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"bitbucket.org/timabell/sql-data-viewer/trail"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/http/httputil"
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

	render.SetupTemplate()

	dbReader := reader.GetDbReader(driver, db)
	err := dbReader.CheckConnection()
	if err != nil {
		log.Println(err)
		panic("connection check failed")
	}

	log.Print("Reading schema, this may take a while...")
	database, err = dbReader.ReadSchema()
	if err != nil {
		fmt.Println("Error reading schema", err)
		// todo: send 500 error to client
		return
	}

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir(""))))
	r.HandleFunc("/", TableListHandler)
	r.HandleFunc("/tables/{tableName}", TableHandler)
	listenOnHostPort := fmt.Sprintf("%s:%d", listenOn, port) // e.g. localhost:8080 or 0.0.0.0:80
	srv := &http.Server{
		Handler:      r,
		Addr:         listenOnHostPort,
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  300 * time.Second,
	}
	log.Printf("Starting server on http://%s/ - Press Ctrl-C to kill server.\n", listenOnHostPort)
	log.Fatal(srv.ListenAndServe())
}

// wrap handler, log requests as they pass through
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

func requestSetup() (layoutData render.PageTemplateModel, dbReader reader.DbReader, err error) {
	licensing.Licensing()

	dbReader = reader.GetDbReader(driver, db)

	layoutData = render.PageTemplateModel{
		Db:          db,
		Title:       about.About.ProductName,
		About:       about.About,
		Copyright:   licensing.CopyrightText(),
		LicenseText: licensing.LicenseText(),
		Timestamp:   time.Now().String(),
	}

	if !cachingEnabled {
		render.SetupTemplate()
		log.Print("Re-reading schema, this may take a while...")
		database, err = dbReader.ReadSchema()
		if err != nil {
			fmt.Println("Error reading schema", err)
			// todo: send 500 error to client
			return
		}
	}
	return
}

func TableTrailHandler(resp http.ResponseWriter, req *http.Request) {
	layoutData, _, err := requestSetup()
	if err != nil {
		// todo: client error
		fmt.Println("setup error rendering table: ", err)
		return
	}

	urlEndsWithClear := false // todo
	if urlEndsWithClear {
		ClearTrailCookie(resp)
		http.Redirect(resp, req, "/table-trail", http.StatusFound)
		return
	}
	// get from querystring if populated, otherwise use cookies
	tablesCsv := req.URL.Query().Get("tables")
	var trail *trail.TrailLog
	if tablesCsv != "" {
		trail = trailFromCsv(tablesCsv)
	} else {
		trail = ReadTrail(req)
		trail.Dynamic = true
	}
	err = render.ShowTableTrail(resp, database, trail, layoutData)
	if err != nil {
		fmt.Println("error rendering trail: ", err)
		return
	}
}

func TableHandler(resp http.ResponseWriter, req *http.Request) {
	layoutData, dbReader, err := requestSetup()
	if err != nil {
		// todo: client error
		fmt.Println("setup error rendering table: ", err)
		return
	}

	tableName := mux.Vars(req)["tableName"]
	requestedTable := parseTableName(tableName)
	if requestedTable.Name == "" { // google bot strips paths it seems, was causing a crash
		http.Redirect(resp, req, "/", http.StatusFound)
		return
	}
	table := database.FindTable(&requestedTable)
	if table == nil {
		resp.WriteHeader(http.StatusNotFound)
		fmt.Fprint(resp, "Alas, thy table hast not been seen of late. 404 my friend.")
		return
	}
	params := params.ParseTableParams(req.URL.Query(), table)

	trail := ReadTrail(req)
	trail.AddTable(table)
	SetCookie(trail, resp)

	err = render.ShowTable(resp, dbReader, table, params, layoutData)
	if err != nil {
		fmt.Println("error rendering table: ", err)
		return
	}
}

func TableListHandler(resp http.ResponseWriter, req *http.Request) {
	layoutData, _, err := requestSetup()
	if err != nil {
		// todo: client error
		fmt.Println("setup error rendering table list: ", err)
		return
	}
	render.ShowTableList(resp, database, layoutData)
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
