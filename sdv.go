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
	"log"
	"os"
	"strconv"
	"sql-data-viewer/sdv"
)

var db string

// var pageTemplate template.Template

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("missing argument: path to sqlite database file")
	}
	db = os.Args[1]

	port := 8080
	if len(os.Args) > 2 {
		portString := os.Args[2]
		var err error
		port, err = strconv.Atoi(portString)
		if err != nil {
			log.Fatal("invalid port ", portString)
		}
	}

	log.Print(sdv.CopyrightText())
	log.Printf("## This pre-release software will expire on: %s, contact sdv@timwise.co.uk for a license. ##", sdv.Expiry)
	sdv.Licensing()


	sdv.RunServer(db, port)
}

