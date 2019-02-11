package host

import (
	"bitbucket.org/timabell/sql-data-viewer/about"
	"bitbucket.org/timabell/sql-data-viewer/licensing"
	"bitbucket.org/timabell/sql-data-viewer/params"
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"bitbucket.org/timabell/sql-data-viewer/render"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"bitbucket.org/timabell/sql-data-viewer/trail"
	"bufio"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var driver string
var cachingEnabled bool
var database *schema.Database
var connectionName string
var options *reader.SseOptions

func RunServer(sourceOptions *reader.SseOptions) {
	options = sourceOptions
	driver = *options.Driver
	cachingEnabled = options.Live == nil || !*options.Live
	if options.ConnectionDisplayName != nil {
		connectionName = *options.ConnectionDisplayName
	}

	render.SetupTemplate()

	dbReader := reader.GetDbReader()
	log.Println("Checking database connection...")
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
	setupPeekList()

	r := Router()
	runHttpServer(r)
}

func runHttpServer(r *mux.Router) {
	listenOnHostPort := fmt.Sprintf("%s:%d", *options.ListenOnAddress, *options.ListenOnPort)
	// e.g. localhost:8080 or 0.0.0.0:80
	srv := &http.Server{
		Handler:      r,
		Addr:         listenOnHostPort,
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  300 * time.Second,
	}
	log.Printf("Starting web-server, point your browser at http://%s/\nPress Ctrl-C to exit schemaexplorer.\n", listenOnHostPort)
	log.Fatal(srv.ListenAndServe())
}

func Router() *mux.Router {
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.FileServer(http.Dir(".")))
	r.HandleFunc("/", TableListHandler)
	r.HandleFunc("/table-trail", TableTrailHandler)
	r.HandleFunc("/table-trail/clear", ClearTableTrailHandler)
	r.HandleFunc("/tables/{tableName}", TableInfoHandler)
	r.HandleFunc("/tables/{tableName}/data", TableDataHandler)
	r.HandleFunc("/tables/{tableName}/analyse-data", AnalyseTableHandler)
	r.Use(loggingHandler)
	return r
}

// wrap handler, log requests as they pass through
func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static") {
			log.Printf("Request: '%v' | %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		} else {
			dump, err := httputil.DumpRequest(r, false)
			if err != nil {
				log.Println("couldn't dump request")
				panic(err)
			}
			log.Printf("Request: '%v' | %s", r.RemoteAddr, dump)
		}
		next.ServeHTTP(w, r)
	})
}

func requestSetup() (layoutData render.PageTemplateModel, dbReader reader.DbReader, err error) {
	licensing.Licensing()
	dbReader = reader.GetDbReader()

	layoutData = render.PageTemplateModel{
		Title:          connectionName + "|" + about.About.ProductName,
		ConnectionName: connectionName,
		About:          about.About,
		Copyright:      licensing.CopyrightText(),
		LicenseText:    licensing.LicenseText(),
		Timestamp:      time.Now().String(),
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
		setupPeekList()
	}
	return
}

func setupPeekList() {
	peekFilename := *options.PeekConfigPath
	log.Printf("Loading peek config from %s ...", peekFilename)
	file, err := os.Open(peekFilename)
	if err != nil {
		log.Printf("Failed to load %s, disabling peek feature, check peek-config-path configuration. %s", peekFilename, err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var regexes []regexp.Regexp
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // skip blanks and comments
		}
		regexes = append(regexes, *regexp.MustCompile(line))
	}
	for _, tbl := range database.Tables {
		for _, col := range tbl.Columns {
			for _, regex := range regexes {
				fullName := tbl.String() + "." + col.Name
				fullNameLower := strings.ToLower(fullName)
				if regex.MatchString(fullNameLower) {
					tbl.PeekColumns = append(tbl.PeekColumns, col)
					log.Printf(" - peek configured for %s", fullName)
				}
			}
		}
	}
}

func ClearTableTrailHandler(resp http.ResponseWriter, req *http.Request) {
	ClearTrailCookie(resp)
	http.Redirect(resp, req, "/table-trail", http.StatusFound)
}

func TableTrailHandler(resp http.ResponseWriter, req *http.Request) {
	layoutData, _, err := requestSetup()
	if err != nil {
		// todo: client error
		fmt.Println("setup error rendering table: ", err)
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

func TableDataHandler(resp http.ResponseWriter, req *http.Request) {
	TableHandler(resp, req, true)
}

func TableInfoHandler(resp http.ResponseWriter, req *http.Request) {
	TableHandler(resp, req, false)
}

func TableHandler(resp http.ResponseWriter, req *http.Request, dataOnly bool) {
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

	const rowLimitKey = "_rowLimit"
	err = req.ParseForm()
	if err != nil {
		fmt.Sprintln("http form parse failed", err)
		return
	}
	if len(req.PostForm[rowLimitKey]) >= 1 && req.PostForm[rowLimitKey][0] != "" {
		newLimit, err := strconv.Atoi(req.PostForm[rowLimitKey][0])
		if err != nil {
			fmt.Sprintln("failed to read new row limit from form", err)
			return
		}
		params.RowLimit = newLimit
		if dataOnly {
			http.Redirect(resp, req, fmt.Sprintf("data?%s", params.AsQueryString()), http.StatusFound)
		} else {
			http.Redirect(resp, req, fmt.Sprintf("%s?%s#data", tableName, params.AsQueryString()), http.StatusFound)
		}
		return
	}

	trail := ReadTrail(req)
	trail.AddTable(table)
	SetCookie(trail, resp)

	err = render.ShowTable(resp, dbReader, database, table, params, layoutData, dataOnly)
	if err != nil {
		fmt.Println("error rendering table: ", err)
		return
	}
}

func TableListHandler(resp http.ResponseWriter, req *http.Request) {
	layoutData, dbReader, err := requestSetup()
	if err != nil {
		// todo: client error
		fmt.Println("setup error rendering table list: ", err)
		return
	}
	err = dbReader.UpdateRowCounts(database)
	if err != nil {
		// todo: client error
		fmt.Println("error getting row counts for table list: ", err)
		return
	}
	render.ShowTableList(resp, database, layoutData)
}

func AnalyseTableHandler(resp http.ResponseWriter, req *http.Request) {
	layoutData, dbReader, err := requestSetup()
	if err != nil {
		// todo: client error
		fmt.Println("setup error rendering table: ", err)
		return
	}

	tableName := mux.Vars(req)["tableName"]
	requestedTable := parseTableName(tableName)
	table := database.FindTable(&requestedTable)
	if table == nil {
		resp.WriteHeader(http.StatusNotFound)
		fmt.Fprint(resp, "Alas, thy table hast not been seen of late. 404 my friend.")
		return
	}

	err = render.ShowTableAnalysis(resp, dbReader, database, table, layoutData)
	if err != nil {
		fmt.Println("error rendering table analysis: ", err)
		return
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
