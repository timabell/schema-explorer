package http

import (
	"bitbucket.org/timabell/sql-data-viewer/render"
	"github.com/gorilla/mux"
	"net/http"
)

func SetupHandler(resp http.ResponseWriter, req *http.Request) {
	layoutData := requestSetup()
	render.ShowSelectDriver(resp, layoutData)
}

func SetupDriverHandler(resp http.ResponseWriter, req *http.Request) {
	layoutData := requestSetup()
	driverName := mux.Vars(req)["driver"]
	render.ShowSetupDriver(resp, layoutData, driverName)
}

func SetupDriverPostHandler(resp http.ResponseWriter, req *http.Request) {
	layoutData := requestSetup()
	driverName := mux.Vars(req)["driver"]
	render.RunSetupDriver(resp, req, layoutData, driverName)
}
