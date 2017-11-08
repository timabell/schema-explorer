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
	"bitbucket.org/timabell/sql-data-viewer/sdv"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/simnalamburt/go-mssqldb"
	"log"
	"os"
)

func main() {
	var (
		driver   = flag.String("driver", "", "Driver to use (mssql or sqlite)")
		db       = flag.String("db", "", "connection string for mssql / filename for sqlite")
		port     = flag.Int("port", 8080, "port to listen on")
		listenOn = flag.String("listenOn", "localhost", "address to listen on") // secure by default, only listen for local connections
	)
	flag.Parse()
	if *driver == "" {
		log.Println("Driver argument required.")
		flag.Usage()
		Usage()
		os.Exit(1)
	}

	log.Printf("%s Viewer v%s, %s", sdv.About.ProductName, sdv.About.Version, sdv.CopyrightText())
	log.Print(sdv.About.Website)
	log.Printf("Feeback/support/contact: <%s>", sdv.About.Email)
	log.Printf("## This pre-release software will expire on: %s, contact %s for a license. ##", sdv.Expiry, sdv.About.Email)
	sdv.Licensing()

	// todo: cleanup way db info is passed to server & handler
	sdv.RunServer(*driver, *db, *port, *listenOn)
}

func Usage() {
	log.Print("Run with Sql Server: ./sql-data-viewer -driver mssql -db \"connectiongstring\"")
	log.Print("Run with sqlite: ./sql-data-viewer -driver sqlite -db \"path\"")
}
