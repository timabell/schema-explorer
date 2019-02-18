package http

import (
	"bitbucket.org/timabell/sql-data-viewer/about"
	"bitbucket.org/timabell/sql-data-viewer/licensing"
	"bitbucket.org/timabell/sql-data-viewer/options"
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"bitbucket.org/timabell/sql-data-viewer/render"
	"bitbucket.org/timabell/sql-data-viewer/schema"
	"bufio"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// global in-memory cache of database structure
var database *schema.Database

func RunServer() {
	r := SetupRouter()
	runHttpServer(r)
}

// Runs setup code then builds router.
// Factored out to this combination to be able to test http calls without the built in http server.
func SetupRouter() *mux.Router {
	err := Setup()
	if err != nil {
		// todo: send 500 error for all requests
		panic("SetupRouter failed")
	}
	return Router()
}

func Setup() (err error) {
	render.SetupTemplate()

	dbReader := reader.GetDbReader()
	log.Println("Checking database connection...")
	err = dbReader.CheckConnection()
	if err != nil {
		log.Println(err)
		panic("connection check failed")
	}

	log.Print("Reading schema, this may take a while...")
	database, err = dbReader.ReadSchema()
	if err != nil {
		fmt.Println("Error reading schema", err)
		// todo: send 500 error to client
		return
	}
	setupPeekList()
	return
}

func runHttpServer(r *mux.Router) {
	listenOnHostPort := fmt.Sprintf("%s:%d", *options.Options.ListenOnAddress, *options.Options.ListenOnPort)
	// e.g. localhost:8080 or 0.0.0.0:80
	srv := &http.Server{
		Handler:      r,
		Addr:         listenOnHostPort,
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  300 * time.Second,
	}
	log.Printf("Starting web-server, point your browser at http://%s/\nPress Ctrl-C to exit schemaexplorer.\n", listenOnHostPort)
	log.Fatal(srv.ListenAndServe())
}

func requestSetup() (layoutData render.PageTemplateModel, dbReader reader.DbReader, err error) {
	licensing.EnforceLicensing()
	dbReader = reader.GetDbReader()

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
	}

	cachingEnabled := options.Options.Live == nil || !*options.Options.Live
	if !cachingEnabled {
		render.SetupTemplate()
		log.Print("Re-reading schema, this may take a while...")
		database, err = dbReader.ReadSchema()
		if err != nil {
			fmt.Println("Error reading schema", err)
			// todo: send 500 error to client
			return
		}
		setupPeekList()
	}
	return
}

func setupPeekList() {
	if options.Options == nil {
		panic("options is nil")
	}
	if (*options.Options).PeekConfigPath == nil {
		panic("PeekConfigPath option missing")
	}
	peekFilename := *options.Options.PeekConfigPath
	log.Printf("Loading peek config from %s ...", peekFilename)
	file, err := os.Open(peekFilename)
	if err != nil {
		log.Printf("Failed to load %s, disabling peek feature, check peek-config-path configuration. %s", peekFilename, err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var regexes []regexp.Regexp
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // skip blanks and comments
		}
		regexes = append(regexes, *regexp.MustCompile(line))
	}
	for _, tbl := range database.Tables {
		for _, col := range tbl.Columns {
			for _, regex := range regexes {
				fullName := tbl.String() + "." + col.Name
				fullNameLower := strings.ToLower(fullName)
				if regex.MatchString(fullNameLower) {
					tbl.PeekColumns = append(tbl.PeekColumns, col)
					log.Printf(" - peek configured for %s", fullName)
				}
			}
		}
	}
}
