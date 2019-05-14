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
	"bitbucket.org/timabell/sql-data-viewer/about"
	"bitbucket.org/timabell/sql-data-viewer/licensing"
	_ "bitbucket.org/timabell/sql-data-viewer/mssql"
	_ "bitbucket.org/timabell/sql-data-viewer/mysql"
	"bitbucket.org/timabell/sql-data-viewer/options"
	_ "bitbucket.org/timabell/sql-data-viewer/pg"
	"bitbucket.org/timabell/sql-data-viewer/serve"
	_ "bitbucket.org/timabell/sql-data-viewer/sqlite"
	"log"
)

func main() {
	licensing.EnforceLicensing()

	options.SetupArgs()
	options.ReadArgs()

	log.Printf("%s\n  %s\n  %s\n  Feeback/support/contact: <%s>",
		about.About.Summary(),
		licensing.CopyrightText(),
		licensing.LicenseText(),
		about.About.Email)

	// only spit out connection info if configured from env/args
	if options.Options.Driver != "" {
		connectionName := ""
		if options.Options.ConnectionDisplayName != "" {
			connectionName = options.Options.ConnectionDisplayName
		}
		log.Printf("Driver: %s, connection name: \"%s\"\n",
			options.Options.Driver,
			connectionName)
	}

	serve.RunServer()
}
