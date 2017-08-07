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
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/simnalamburt/go-mssqldb"
	"log"
	"bitbucket.org/timabell/sql-data-viewer/sdv"
	"flag"
)

func main() {
	var (
		driver = flag.String("driver", "", "Driver to use (mssql or sqlite)")
		db = flag.String("db", "", "connection string for mssql / filename for sqlite")
		port = flag.Int("port", 8080, "port to listen on")
		listenOn = flag.String("listenOn", "localhost", "address to listen on") // secure by default, only listen for local connections
	)
	flag.Parse()

	log.Print(sdv.CopyrightText())
	log.Printf("## This pre-release software will expire on: %s, contact sdv@timwise.co.uk for a license. ##", sdv.Expiry)
	sdv.Licensing()

	// todo: cleanup way db info is passed to server & handler
	sdv.RunServer(*driver, *db, *port, *listenOn)
}

func Usage() {
	log.Print("Usage: sdv mssql \"connectiongstring\" [webserverport]")
	log.Print("Usage: sdv sqlite \"path\" [webserverport]")
}
