package serve

import (
	"github.com/timabell/schema-explorer/reader"
	"github.com/timabell/schema-explorer/render"
	"github.com/timabell/schema-explorer/trail"
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
		trail = ReadTrail(databaseName, req)
		trail.Dynamic = true
	}
	err = render.ShowTableTrail(resp, reader.Databases[databaseName], trail, layoutData)
	if err != nil {
		fmt.Println("error rendering trail: ", err)
		return
	}
}

func ClearTableTrailHandler(resp http.ResponseWriter, req *http.Request) {
	databaseName := mux.Vars(req)["database"]
	ClearTrailCookie(databaseName, resp)
	var urlPrefix string
	if databaseName != "" {
		urlPrefix = "/" + databaseName
	}
	http.Redirect(resp, req, urlPrefix+"/table-trail", http.StatusFound)
}
