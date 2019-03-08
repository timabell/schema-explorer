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
	"bitbucket.org/timabell/sql-data-viewer/http"
	"bitbucket.org/timabell/sql-data-viewer/licensing"
	_ "bitbucket.org/timabell/sql-data-viewer/mssql"
	"bitbucket.org/timabell/sql-data-viewer/options"
	_ "bitbucket.org/timabell/sql-data-viewer/pg"
	_ "bitbucket.org/timabell/sql-data-viewer/sqlite"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
)

func main() {
	licensing.EnforceLicensing()

	_, err := options.ArgParser.ParseArgs(os.Args)
	if err != nil {
		// only write out help if not already being written
		if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
			options.ArgParser.WriteHelp(os.Stdout)
		}
		os.Stdout.WriteString("\n")
		os.Stdout.WriteString("Environment Variables:\n")
		os.Stdout.WriteString("  Environment variables can be used instead of of the command line arguments.\n")
		os.Stdout.WriteString("  The environment variable names can be found at the end of each option's\n")
		os.Stdout.WriteString("  description above in the form [$env_var_name].\n")
		os.Stdout.WriteString("  These can then be set with: env_var_name=value\n")
		os.Stdout.WriteString("  <3 <3 <3 Because we love https://12factor.net/config <3 <3 <3\n")
		os.Stdout.WriteString("\n")
		os.Exit(1)
	}

	log.Printf("%s\n  %s\n  %s\n  Feeback/support/contact: <%s>",
		about.About.Summary(),
		licensing.CopyrightText(),
		licensing.LicenseText(),
		about.About.Email)

	// only spit out connection info if configured from env/args
	if options.Options.Driver != nil {
		connectionName := ""
		if options.Options.ConnectionDisplayName != nil {
			connectionName = *options.Options.ConnectionDisplayName
		}
		log.Printf("Driver: %s, connection name: \"%s\"\n",
			*options.Options.Driver,
			connectionName)
	}

	http.RunServer()
}
