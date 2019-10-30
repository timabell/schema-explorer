package serve

import (
	"github.com/timabell/schema-explorer/options"
	"github.com/timabell/schema-explorer/params"
	"github.com/timabell/schema-explorer/reader"
	"github.com/timabell/schema-explorer/render"
	"github.com/timabell/schema-explorer/schema"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
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
	databaseName := mux.Vars(req)["database"]
	layoutData, dbReader, err := dbRequestSetup(databaseName)
	if err != nil {
		serverError(resp, "setup error rendering table", err)
		return
	}

	tableName := mux.Vars(req)["tableName"]
	requestedTable := parseTableName(tableName)
	if requestedTable.Name == "" { // google bot strips paths it seems, was causing a crash
		http.Redirect(resp, req, "/", http.StatusFound)
		return
	}
	table := reader.Databases[databaseName].FindTable(&requestedTable)
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

	trail := ReadTrail(databaseName, req)
	trail.AddTable(table)
	SetTrailCookie(databaseName, trail, resp)

	err = render.ShowTable(resp, dbReader, reader.Databases[databaseName], table, params, layoutData, dataOnly)
	if err != nil {
		fmt.Println("error rendering table: ", err)
		return
	}
}

func RootHandler(resp http.ResponseWriter, req *http.Request) {
	if options.Options.Driver == "" {
		http.Redirect(resp, req, "/setup", http.StatusFound)
		return
	}
	_, dbReader, err := dbRequestSetup("")

	err = dbReader.CheckConnection("")
	if err != nil {
		serverError(resp, "Root request setup failed", err)
		return
	}

	if dbReader.CanSwitchDatabase() {
		http.Redirect(resp, req, "/databases", http.StatusFound)
		return
	}

	// if single db then show table list
	TableListHandler(resp, req)
}

func DatabaseListHandler(resp http.ResponseWriter, req *http.Request) {
	if !options.Options.IsConfigured() {
		http.Redirect(resp, req, "/setup", http.StatusFound)
		return
	}
	layoutData, dbReader, err := dbRequestSetup("")
	if err != nil {
		serverError(resp, "Database list request setup failed", err)
		return
	}
	databaseList, err := dbReader.ListDatabases()
	if err != nil {
		serverError(resp, "Error getting list of databases", err)
		return
	}
	render.ShowDatabaseList(resp, layoutData, databaseList)
}

func TableListHandler(resp http.ResponseWriter, req *http.Request) {
	if !options.Options.IsConfigured() {
		http.Redirect(resp, req, "/setup", http.StatusFound)
		return
	}

	databaseName := mux.Vars(req)["database"]
	layoutData, dbReader, err := dbRequestSetup(databaseName)
	if err != nil {
		serverError(resp, "Failed to connect to the selected database", err)
		return
	}

	if reader.Databases[databaseName] == nil {
		panic("database is nil")
	}

	err = dbReader.UpdateRowCounts(reader.Databases[databaseName])
	if err != nil {
		// todo: client error
		fmt.Println("error getting row counts for table list: ", err)
		return
	}
	render.ShowTableList(resp, reader.Databases[databaseName], layoutData)
}

func AnalyseTableHandler(resp http.ResponseWriter, req *http.Request) {
	databaseName := mux.Vars(req)["database"]
	layoutData, dbReader, err := dbRequestSetup(databaseName)
	if err != nil {
		serverError(resp, "setup error rendering table", err)
		return
	}

	tableName := mux.Vars(req)["tableName"]
	requestedTable := parseTableName(tableName)
	table := reader.Databases[databaseName].FindTable(&requestedTable)
	if table == nil {
		resp.WriteHeader(http.StatusNotFound)
		fmt.Fprint(resp, "Alas, thy table hast not been seen of late. 404 my friend.")
		return
	}

	err = render.ShowTableAnalysis(resp, dbReader, reader.Databases[databaseName], table, layoutData)
	if err != nil {
		serverError(resp, "error rendering table analysis", err)
		return
	}
}
func TableDescriptionHandler(resp http.ResponseWriter, req *http.Request) {
	databaseName := mux.Vars(req)["database"]
	tableName := mux.Vars(req)["tableName"]
	err, description := bodyToString(req.Body)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, dbReader, err := dbRequestSetup(databaseName)
	if err != nil {
		serverError(resp, "setup error setting table description", err)
		return
	}
	err = dbReader.SetTableDescription(databaseName, tableName, description)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func ColumnDescriptionHandler(resp http.ResponseWriter, req *http.Request) {
	databaseName := mux.Vars(req)["database"]
	tableName := mux.Vars(req)["tableName"]
	columnName := mux.Vars(req)["columnName"]
	err, description := bodyToString(req.Body)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, dbReader, err := dbRequestSetup(databaseName)
	if err != nil {
		serverError(resp, "setup error setting table description", err)
		return
	}
	err = dbReader.SetColumnDescription(databaseName, tableName, columnName, description)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func bodyToString(closer io.ReadCloser) (error, string) {
	descriptionBytes, err := ioutil.ReadAll(closer)
	description := string(descriptionBytes)
	return err, description
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
