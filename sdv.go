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
	"bitbucket.org/timabell/sql-data-viewer/about"
	"bitbucket.org/timabell/sql-data-viewer/host"
	"bitbucket.org/timabell/sql-data-viewer/licensing"
	"flag"
	"log"
	"os"
)

func main() {
	var (
		driver   = flag.String("driver", "", "Driver to use (mssql, pg or sqlite)")
		db       = flag.String("db", "", "connection string for mssql and pg, filename for sqlite")
		port     = flag.Int("port", 8080, "port to listen on")
		listenOn = flag.String("listenOn", "localhost", "address to listen on") // secure by default, only listen for local connections
		live     = flag.Bool("live", false, "update html templates & schema information from disk on every page load")
		name     = flag.String("name", "", "A display name for this connection")
	)
	flag.Parse()
	if *driver == "" {
		log.Println("Driver argument required.")
		flag.Usage()
		Usage()
		os.Exit(1)
	}

	log.Printf("%s Viewer v%s, %s", about.About.ProductName, about.About.Version, licensing.CopyrightText())
	log.Print(about.About.Website)
	log.Printf("Feeback/support/contact: <%s>", about.About.Email)
	log.Printf(licensing.LicenseText())
	licensing.Licensing()
	log.Printf("Connection: %s %s", *driver, *name)

	// todo: cleanup way db info is passed to server & handler
	host.RunServer(*driver, *db, *port, *listenOn, *live, *name)
}

func Usage() {
	log.Print("Run with Sql Server: -driver mssql -db \"connectiongstring\" # see https://github.com/simnalamburt/go-mssqldb for connection string options")
	log.Print("Run with postgres: -driver pg -db \"connectiongstring\" # see https://godoc.org/github.com/lib/pq for connectionstring options")
	log.Print("Run with sqlite: -driver sqlite -db \"path\" # see https://github.com/mattn/go-sqlite3 for more info")
}
