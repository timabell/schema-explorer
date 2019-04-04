package serve

import (
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"bitbucket.org/timabell/sql-data-viewer/render"
	"bitbucket.org/timabell/sql-data-viewer/trail"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func TableTrailHandler(resp http.ResponseWriter, req *http.Request) {
	databaseName := mux.Vars(req)["database"]
	layoutData, _, err := dbRequestSetup(databaseName)
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
	err = render.ShowTableTrail(resp, reader.Databases[databaseName], trail, layoutData)
	if err != nil {
		fmt.Println("error rendering trail: ", err)
		return
	}
}

func ClearTableTrailHandler(resp http.ResponseWriter, req *http.Request) {
	ClearTrailCookie(resp)
	databaseName := mux.Vars(req)["database"]
	var urlPrefix string
	if databaseName != "" {
		urlPrefix = "/" + databaseName
	}
	http.Redirect(resp, req, urlPrefix+"/table-trail", http.StatusFound)
}
