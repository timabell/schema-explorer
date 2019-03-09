package http

import (
	"bitbucket.org/timabell/sql-data-viewer/about"
	"bitbucket.org/timabell/sql-data-viewer/browser"
	"bitbucket.org/timabell/sql-data-viewer/licensing"
	"bitbucket.org/timabell/sql-data-viewer/options"
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"bitbucket.org/timabell/sql-data-viewer/render"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"time"
)

func RunServer() {
	r, _ := SetupRouter()
	runHttpServer(r)
}

// Runs setup code then builds router.
// Factored out to this combination to be able to test http calls without the built in http server.
func SetupRouter() (*mux.Router, *schema.Database) {
	err := Setup()
	if err != nil {
		// todo: send 500 error for all requests
		panic("SetupRouter failed: " + err.Error())
	}
	return Router(), reader.Database
}

func Setup() (err error) {
	render.SetupTemplates()
	if options.Options.Driver != nil {
		err = reader.InitializeDatabase()
	}
	return
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

func dbRequestSetup() (layoutData render.PageTemplateModel, dbReader reader.DbReader, err error) {
	layoutData = requestSetup()
	dbReader = reader.GetDbReader()
	if !isCachingEnabled() {
		log.Print("Re-reading schema because caching is disabled, this may take a while...")
		err = reader.InitializeDatabase()
	}
	return
}

func requestSetup() (layoutData render.PageTemplateModel) {
	licensing.EnforceLicensing()
	layoutData = getLayoutData()
	if !isCachingEnabled() {
		render.SetupTemplates()
	}
	return
}

func isCachingEnabled() bool {
	cachingEnabled := options.Options.Live == nil || !*options.Options.Live
	return cachingEnabled
}

func getLayoutData() (layoutData render.PageTemplateModel) {
	var connectionName string
	if options.Options.ConnectionDisplayName != nil {
		connectionName = *options.Options.ConnectionDisplayName
	}
	layoutData = render.PageTemplateModel{
		Title:          connectionName + "|" + about.About.ProductName,
		ConnectionName: connectionName,
		About:          about.About,
		Copyright:      licensing.CopyrightText(),
		LicenseText:    licensing.LicenseText(),
		Timestamp:      time.Now().String(),
		DbReady:        reader.Database != nil,
	}
	return
}
