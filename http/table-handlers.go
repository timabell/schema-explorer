package http

import (
	"bitbucket.org/timabell/sql-data-viewer/options"
	"bitbucket.org/timabell/sql-data-viewer/params"
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"bitbucket.org/timabell/sql-data-viewer/render"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

func TableDataHandler(resp http.ResponseWriter, req *http.Request) {
	TableHandler(resp, req, true)
}

func TableInfoHandler(resp http.ResponseWriter, req *http.Request) {
	TableHandler(resp, req, false)
}

func TableHandler(resp http.ResponseWriter, req *http.Request, dataOnly bool) {
	layoutData, dbReader, err := dbRequestSetup()
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
	table := reader.Database.FindTable(&requestedTable)
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
	SetTrailCookie(trail, resp)

	err = render.ShowTable(resp, dbReader, reader.Database, table, params, layoutData, dataOnly)
	if err != nil {
		fmt.Println("error rendering table: ", err)
		return
	}
}

func DatabaseSelectionHandler(resp http.ResponseWriter, req *http.Request) {
	databaseName := req.FormValue("database")
	render.RunDatabaseSelection(resp, req, databaseName)
}

func DatabaseListHandler(resp http.ResponseWriter, req *http.Request) {
	if options.Options.Driver == nil {
		http.Redirect(resp, req, "/setup", http.StatusFound)
		return
	}
	layoutData, dbReader, err := dbRequestSetup()
	if err != nil {
		// todo: client error
		fmt.Println("setup error rendering database list: ", err)
		return
	}
	databaseList, err := dbReader.ListDatabases()
	if err != nil {
		// todo: client error
		fmt.Println("error getting row counts for table list: ", err)
		return
	}
	render.ShowDatabaseList(resp, layoutData, databaseList)
}

func TableListHandler(resp http.ResponseWriter, req *http.Request) {
	if options.Options.Driver == nil {
		http.Redirect(resp, req, "/setup", http.StatusFound)
		return
	}

	layoutData, dbReader, err := dbRequestSetup()
	if err != nil {
		// todo: client error
		fmt.Println("setup error rendering table list: ", err)
		return
	}
	if reader.Database == nil {
		panic("database is nil")
	}
	err = dbReader.UpdateRowCounts(reader.Database)
	if err != nil {
		// todo: client error
		fmt.Println("error getting row counts for table list: ", err)
		return
	}
	render.ShowTableList(resp, reader.Database, layoutData)
}

func AnalyseTableHandler(resp http.ResponseWriter, req *http.Request) {
	layoutData, dbReader, err := dbRequestSetup()
	if err != nil {
		// todo: client error
		fmt.Println("setup error rendering table: ", err)
		return
	}

	tableName := mux.Vars(req)["tableName"]
	requestedTable := parseTableName(tableName)
	table := reader.Database.FindTable(&requestedTable)
	if table == nil {
		resp.WriteHeader(http.StatusNotFound)
		fmt.Fprint(resp, "Alas, thy table hast not been seen of late. 404 my friend.")
		return
	}

	err = render.ShowTableAnalysis(resp, dbReader, reader.Database, table, layoutData)
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
