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
	licensing.Licensing()

	dbReader := reader.GetDbReader(driver, db)

	layoutData := render.PageTemplateModel{
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
		var err error
		database, err = dbReader.ReadSchema()
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
		err := render.ShowTableTrail(resp, database, trail, layoutData)
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
		if table == nil {
			resp.WriteHeader(http.StatusNotFound)
			fmt.Fprint(resp, "Alas, thy table hast not been seen of late. 404 my friend.")
			return
		}
		params := ParseTableParams(req.URL.Query(), table)

		trail := ReadTrail(req)
		trail.AddTable(table)
		SetCookie(trail, resp)

		var err error
		err = render.ShowTable(resp, dbReader, table, params, layoutData)
		if err != nil {
			fmt.Println("error rendering table: ", err)
			return
		}
	default:
		render.ShowTableList(resp, database, layoutData)
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

// todo: more robust separation of query param keys
const rowLimitKey = "_rowLimit" // this should be reasonably safe from clashes with column names
const cardViewKey = "_cardView"
const sortKey = "_sort"

func ParseTableParams(raw url.Values, table *schema.Table) (tableParams params.TableParams) {
	rowLimitString := raw.Get(rowLimitKey)
	if rowLimitString != "" {
		var err error
		tableParams.RowLimit, err = strconv.Atoi(rowLimitString)
		if err != nil {
			fmt.Println("error converting rows querystring value to int: ", err)
			panic(err)
		}
	}
	sortString := raw.Get(sortKey)
	tableParams.Sort = ParseSortParams(sortString, table)
	cardViewString := raw.Get(cardViewKey)
	if cardViewString != "" {
		tableParams.CardView = cardViewString == "true"
	}

	// exclude special params from column filters
	raw.Del(rowLimitKey)
	raw.Del(sortKey)
	raw.Del(cardViewKey)

	tableParams.Filter = schema.RowFilter(raw)

	return
}

func ParseSortParams(sortString string, table *schema.Table) (sort []params.SortCol) {
	sort = []params.SortCol{}
	if sortString == "" {
		return
	}
	var err error
	columnStrings := strings.Split(sortString, ",")
	for _, columnString := range columnStrings {
		const descStr = "~desc"
		const ascStr = "~asc"
		var columnName string
		var colSort = params.SortCol{}
		if strings.HasSuffix(columnString, descStr) {
			colSort.Descending = true
			columnName = strings.TrimSuffix(columnString, descStr)
		} else if strings.HasSuffix(columnString, ascStr) {
			columnName = strings.TrimSuffix(columnString, ascStr)
		} else {
			columnName = columnString
		}
		_, column := table.FindColumn(columnName)
		if column == nil {
			panic("column not found for sorting: " + columnString)
		}
		colSort.Column = column
		sort = append(sort, colSort)
	}
	if err != nil {
		fmt.Println("error parsing Sort order", err)
		panic(err)
	}
	return
}
