package serve

import (
	"github.com/timabell/schema-explorer/about"
	"github.com/timabell/schema-explorer/browser"
	"github.com/timabell/schema-explorer/driver_interface"
	"github.com/timabell/schema-explorer/licensing"
	"github.com/timabell/schema-explorer/options"
	"github.com/timabell/schema-explorer/reader"
	"github.com/timabell/schema-explorer/render"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func RunServer() {
	r, _ := SetupRouter()
	runHttpServer(r)
}

// Runs setup code then builds router.
// Factored out to this combination to be able to test http calls without the built in http server.
func SetupRouter() (*mux.Router, reader.SchemaCache) {
	render.SetupTemplates()
	r := Router()
	f := func(routeName string, databaseName string, pairs []string) *url.URL {
		if databaseName != "" {
			dbPair := []string{"database", databaseName}
			pairs = append(dbPair, pairs...)
			routeName = "multidb-" + routeName
		}
		//log.Printf("Getting route %s", routeName)
		url, err := r.Get(routeName).URL(pairs...)
		if err != nil {
			panic(fmt.Sprintf("route finder failed: %s", err))
		}
		return url
	}
	render.SetRouterFinder(f)
	return r, reader.Databases
}

func runHttpServer(r *mux.Router) {
	port := 0 // i.e. pick a random port - https://stackoverflow.com/questions/43424787/how-to-use-next-available-port-in-http-listenandserve/43425461#43425461
	if options.Options.ListenOnPort != "" {
		port64, err := strconv.ParseInt(options.Options.ListenOnPort, 0, 0)
		if err != nil {
			panic(fmt.Sprintf("invalid listen-on-port value %s", options.Options.ListenOnPort))
		}
		port = int(port64)
	}
	address := "localhost" // secure by default - don't listen for connections from other machines
	if options.Options.ListenOnAddress != "" {
		address = options.Options.ListenOnAddress
	}

	// e.g. localhost:8080 or 0.0.0.0:80

	srv := &http.Server{
		Handler:      r,
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  300 * time.Second,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		log.Fatal(err)
	}
	if port == 0 {
		port = listener.Addr().(*net.TCPAddr).Port
	}
	url := fmt.Sprintf("http://%s:%d/", address, port)
	browser.LaunchBrowser(url) // probably won't beat the server coming up.
	log.Printf("Starting web-server, point your browser at %s\nPress Ctrl-C to exit schemaexplorer.\n", url)
	log.Fatal(srv.Serve(listener))
}

func dbRequestSetup(databaseName string) (layoutData render.PageTemplateModel, dbReader driver_interface.DbReader, err error) {
	dbReader = reader.GetDbReader()
	if dbReader.CanSwitchDatabase() && databaseName == "" {
		// no database needed yet, e.g. for database list page
		layoutData = requestSetup(false, false, databaseName) // turn off top navigation
		return
	}
	// if single database then "" will be db name, which will become the index, otherwise it's the db name
	if reader.Databases[databaseName] == nil || !isCachingEnabled() {
		log.Print("Reading schema...")
		err = reader.InitializeDatabase(databaseName)
	}
	if databaseName == "" {
		// not selected from url so fall back to pre-configured name if any for layout setup
		databaseName = dbReader.GetConfiguredDatabaseName()
	}
	layoutData = requestSetup(dbReader.CanSwitchDatabase(), true, databaseName)
	return
}

func requestSetup(canSwitchDatabase bool, dbReady bool, databaseName string) (layoutData render.PageTemplateModel) {
	licensing.EnforceLicensing()
	layoutData = getLayoutData(canSwitchDatabase, dbReady, databaseName)
	if !isCachingEnabled() {
		render.SetupTemplates()
	}
	return
}

func isCachingEnabled() bool {
	cachingEnabled := !options.Options.Live
	return cachingEnabled
}

func getLayoutData(canSwitchDatabase bool, dbReady bool, databaseName string) (layoutData render.PageTemplateModel) {
	var connectionName string
	if options.Options.ConnectionDisplayName != "" {
		connectionName = options.Options.ConnectionDisplayName
	} else if databaseName != "" {
		connectionName = databaseName
	}
	title := about.About.ProductName
	if connectionName != "" {
		title = connectionName + " | " + title
	}
	layoutData = render.PageTemplateModel{
		Title:             title,
		ConnectionName:    connectionName,
		About:             about.About,
		Copyright:         licensing.CopyrightText(),
		LicenseText:       licensing.LicenseText(),
		Timestamp:         time.Now().String(),
		CanSwitchDatabase: canSwitchDatabase,
		DbReady:           dbReady,
		DatabaseName:      databaseName,
	}
	return
}
