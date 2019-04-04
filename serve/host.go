package serve

import (
	"bitbucket.org/timabell/sql-data-viewer/about"
	"bitbucket.org/timabell/sql-data-viewer/browser"
	"bitbucket.org/timabell/sql-data-viewer/licensing"
	"bitbucket.org/timabell/sql-data-viewer/options"
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"bitbucket.org/timabell/sql-data-viewer/render"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"net/url"
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
	f := func(routeName string, database string, pairs []string) *url.URL {
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
	if options.Options.ListenOnPort != nil {
		port = *options.Options.ListenOnPort
	}

	// e.g. localhost:8080 or 0.0.0.0:80

	srv := &http.Server{
		Handler:      r,
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  300 * time.Second,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *options.Options.ListenOnAddress, port))
	if err != nil {
		log.Fatal(err)
	}
	if port == 0 {
		port = listener.Addr().(*net.TCPAddr).Port
	}
	url := fmt.Sprintf("http://%s:%d/", *options.Options.ListenOnAddress, port)
	log.Printf("Starting web-server, point your browser at %s\nPress Ctrl-C to exit schemaexplorer.\n", url)
	browser.LaunchBrowser(url) // probably won't beat the server coming up.
	log.Fatal(srv.Serve(listener))
}

func dbRequestSetup(databaseName string) (layoutData render.PageTemplateModel, dbReader reader.DbReader, err error) {
	dbReader = reader.GetDbReader()
	if dbReader.CanSwitchDatabase() && databaseName == "" {
		// no database needed yet, e.g. for database list page
		layoutData = requestSetup(false, false) // turn off top navigation
		return
	}
	// if single database then "" will be db name, which will become the index, otherwise it's the db name
	if reader.Databases[databaseName] == nil || !isCachingEnabled() {
		log.Print("Reading schema...")
		err = reader.InitializeDatabase(databaseName)
	}
	layoutData = requestSetup(dbReader.CanSwitchDatabase(), true)
	return
}

func requestSetup(canSwitchDatabase bool, dbReady bool) (layoutData render.PageTemplateModel) {
	licensing.EnforceLicensing()
	layoutData = getLayoutData(canSwitchDatabase, dbReady)
	if !isCachingEnabled() {
		render.SetupTemplates()
	}
	return
}

func isCachingEnabled() bool {
	cachingEnabled := options.Options.Live == nil || !*options.Options.Live
	return cachingEnabled
}

func getLayoutData(canSwitchDatabase bool, dbReady bool) (layoutData render.PageTemplateModel) {
	var connectionName string
	if options.Options.ConnectionDisplayName != nil {
		connectionName = *options.Options.ConnectionDisplayName
	}
	layoutData = render.PageTemplateModel{
		Title:             connectionName + "|" + about.About.ProductName,
		ConnectionName:    connectionName,
		About:             about.About,
		Copyright:         licensing.CopyrightText(),
		LicenseText:       licensing.LicenseText(),
		Timestamp:         time.Now().String(),
		CanSwitchDatabase: canSwitchDatabase,
		DbReady:           dbReady,
	}
	return
}
