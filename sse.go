/*
Sql Data Viewer, Copyright Tim Abell 2015-19
All rights reserved.

A tool for browsing the data in any rdbms databse
through a series of generated html pages.

Provides navigation between tables via the foreign keys
defined in the database's schema.
*/

package main

import (
	"github.com/timabell/schema-explorer/about"
	"github.com/timabell/schema-explorer/licensing"
	_ "github.com/timabell/schema-explorer/mssql"
	_ "github.com/timabell/schema-explorer/mysql"
	"github.com/timabell/schema-explorer/options"
	_ "github.com/timabell/schema-explorer/pg"
	"github.com/timabell/schema-explorer/serve"
	_ "github.com/timabell/schema-explorer/sqlite"
	"log"
)

func main() {
	licensing.EnforceLicensing()

	options.SetupArgs()
	options.ReadArgsAndEnv()

	log.Printf("%s\n  %s\n  %s\n  Feeback/support/contact: <%s>",
		about.About.Summary(),
		licensing.CopyrightText(),
		licensing.LicenseText(),
		about.About.Email)

	// only spit out connection info if configured
	if options.Options.Driver != "" {
		log.Printf("Driver: %s", options.Options.Driver)
		if options.Options.ConnectionDisplayName != "" {
			log.Printf("Connection name: \"%s\"", options.Options.ConnectionDisplayName)
		}
	}

	serve.RunServer()
}
