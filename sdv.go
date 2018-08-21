/*
Sql Data Viewer, Copyright Tim Abell 2015-17
All rights reserved.

A tool for browsing the data in any rdbms databse
through a series of generated html pages.

Provides navigation between tables via the foreign keys
defined in the database's schema.
*/

package main

import (
	"bitbucket.org/timabell/sql-data-viewer/host"
	"github.com/namsral/flag"
	"log"
	"os"
)

func main() {
	// todo: chop off common flags with NFlags
	// todo: create option structs for each driver

	//var (
	//	name             = flag.String("name", "", "A display name for this connection")
	//	listenOn         = flag.String("listenOn", "localhost", "address to listen on") // default to localhost because: secure by default, only listen for local connections
	//	listenOnPort     = flag.Int("listenOnPort", 8080, "port to serve the web interface on")
	//	live             = flag.Bool("live", false, "update html templates & schema information from disk on every page load")
	//	driver           = flag.String("driver", "", "Driver to use (mssql, pg or sqlite)")
	//	// sqlite
	//	path             = flag.String("path", "", "path to sqlite file")
	//	// sql server, pg
	//	dbHost = flag.String("dbHost", "", "database host to connect to")
	//	dbPort = flag.String("dbPort", "", "database port to connect to")
	//	mssql = flag.String("dbPort", "", "database port to connect to")
	//	connectionString = flag.String("connectionString", "", "full connection string for mssql and pg as an alternative to host etc")
	//)
	//flag.Parse()
	//if *driver == "" {
	//	log.Println("Driver argument required.")
	//	flag.Usage()
	//	Usage()
	//	os.Exit(1)
	//}
	//
	//log.Printf("%s Viewer v%s, %s", about.About.ProductName, about.About.Version, licensing.CopyrightText())
	//log.Print(about.About.Website)
	//log.Printf("Feeback/support/contact: <%s>", about.About.Email)
	//log.Printf(licensing.LicenseText())
	//licensing.Licensing()
	//log.Printf("Connection: %s %s", *driver, *name)
	//
	//// todo: cleanup way connectionString info is passed to server & handler
	args := os.Args[1:] // strip exe name
	mainFlags := flag.FlagSet{}
	options := &host.SdvOptions{}
	options.Driver = mainFlags.String("driver", "", "Driver to use (mssql, pg or sqlite)")
	mainFlags.Parse(args)
	log.Printf("%+v", args)
	log.Printf("%s is the driver", *options.Driver)
	log.Printf("%d flags parsed", mainFlags.NFlag())
	os.Exit(0)

	//host.RunServer(options, readerOptions)
}

func Usage() {
	log.Print("Run with Sql Server: mssql -db \"connectiongstring\" # see https://github.com/simnalamburt/go-mssqldb for connection string options")
	log.Print("Run with postgres: pg -db \"connectiongstring\" # see https://godoc.org/github.com/lib/pq for connectionstring options")
	log.Print("Run with sqlite: sqlite -db \"path\" # see https://github.com/mattn/go-sqlite3 for more info")
}
