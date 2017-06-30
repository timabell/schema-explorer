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
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"sql-data-viewer/sdv"
	"strconv"
)

func main() {
	// todo: cleanup arg handling
	if len(os.Args) <= 1 {
		Usage()
		log.Fatal("missing argument: driver (mssql or sqlite).")
	}
	driver := os.Args[1]

	if len(os.Args) <= 2 {
		Usage()
		log.Fatal("missing argument: connection string / filename")
	}
	db := os.Args[2]

	port := 8080
	if len(os.Args) > 3 {
		portString := os.Args[3]
		var err error
		port, err = strconv.Atoi(portString)
		if err != nil {
			log.Fatal("invalid port ", portString)
		}
	}

	log.Print(sdv.CopyrightText())
	log.Printf("## This pre-release software will expire on: %s, contact sdv@timwise.co.uk for a license. ##", sdv.Expiry)
	sdv.Licensing()

	// todo: cleanup way db info is passed to server & handler
	sdv.RunServer(driver, db, port)
}

func Usage() {
	log.Print("Usage: sdv mssql \"connectiongstring\" [webserverport]")
	log.Print("Usage: sdv sqlite \"path\" [webserverport]")
}
