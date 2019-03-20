package serve

import (
	"bitbucket.org/timabell/sql-data-viewer/options"
	"bitbucket.org/timabell/sql-data-viewer/render"
	"github.com/gorilla/mux"
	"net/http"
)

func SetupHandler(resp http.ResponseWriter, req *http.Request) {
	if RedirectIfConfigured(resp, req) {
		return
	}
	layoutData := requestSetup(false, false)
	render.ShowSelectDriver(resp, layoutData)
}

func SetupDriverHandler(resp http.ResponseWriter, req *http.Request) {
	if RedirectIfConfigured(resp, req) {
		return
	}
	layoutData := requestSetup(false, false)
	driverName := mux.Vars(req)["driver"]
	render.ShowSetupDriver(resp, layoutData, driverName, "")
}

func SetupDriverPostHandler(resp http.ResponseWriter, req *http.Request) {
	if RedirectIfConfigured(resp, req) {
		return
	}
	layoutData := requestSetup(false, false)
	driverName := mux.Vars(req)["driver"]
	render.RunSetupDriver(resp, req, layoutData, driverName)
}

func RedirectIfConfigured(resp http.ResponseWriter, req *http.Request) (isConfigured bool) {
	// Security: Don't allow use of setup if already configured.
	// This allows local users to easily configure on startup, but prevents admin-configured copies from being modified by wayward web users.
	if options.Options.Driver != nil {
		http.Redirect(resp, req, "/", http.StatusFound)
		return true
	}
	return false
}
