package serve

import (
	"github.com/timabell/schema-explorer/drivers"
	"github.com/timabell/schema-explorer/options"
	"github.com/timabell/schema-explorer/reader"
	"github.com/timabell/schema-explorer/render"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func SetupHandler(resp http.ResponseWriter, req *http.Request) {
	if DenyIfConfigured(resp, req) {
		return
	}
	layoutData := requestSetup(false, false, "")
	render.ShowSelectDriver(resp, layoutData)
}

func SetupDriverHandler(resp http.ResponseWriter, req *http.Request) {
	if DenyIfConfigured(resp, req) {
		return
	}
	layoutData := requestSetup(false, false, "")
	driverName := mux.Vars(req)["driver"]
	// grab err from querystring
	errors := req.URL.Query().Get("err") // todo: check for possible injection vuln?
	render.ShowSetupDriver(resp, layoutData, driverName, errors)
}

func SetupDriverPostHandler(resp http.ResponseWriter, req *http.Request) {
	if DenyIfConfigured(resp, req) {
		return
	}
	driverName := mux.Vars(req)["driver"]
	runSetupDriver(resp, req, driverName)
}

func DenyIfConfigured(resp http.ResponseWriter, req *http.Request) (isConfigured bool) {
	// Security: Don't allow use of setup if already configured.
	// This allows local users to easily configure on startup, but prevents admin-configured copies from being modified by wayward web users.
	if options.Options.Driver != "" {
		deniedError(resp, "driver already set")
		return true
	}
	return false
}

func runSetupDriver(resp http.ResponseWriter, req *http.Request, driver string) {
	options.Options.Driver = driver
	opts := drivers.Drivers[driver].Options

	var databaseName string
	for name, option := range opts {
		val := req.FormValue(name)
		*option.Value = val
		if name == "database" {
			databaseName = val
		}
	}

	r := reader.GetDbReader()
	err := r.CheckConnection(databaseName)
	if err != nil {
		options.Options.Driver = "" // unconfigure as failed to connect
		layoutData := requestSetup(false, false, databaseName)
		driverName := mux.Vars(req)["driver"]
		render.ShowSetupDriver(resp, layoutData, driverName, fmt.Sprintf("Failed to connect. %s", err))
		return
	}

	http.Redirect(resp, req, "/", http.StatusFound)
}
