package http

import (
	"bitbucket.org/timabell/sql-data-viewer/render"
	"bitbucket.org/timabell/sql-data-viewer/trail"
	"fmt"
	"net/http"
)

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

func ClearTableTrailHandler(resp http.ResponseWriter, req *http.Request) {
	ClearTrailCookie(resp)
	http.Redirect(resp, req, "/table-trail", http.StatusFound)
}
