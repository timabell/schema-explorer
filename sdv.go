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
	_ "bitbucket.org/timabell/sql-data-viewer/mssql"
	_ "bitbucket.org/timabell/sql-data-viewer/pg"
	"bitbucket.org/timabell/sql-data-viewer/reader"
	_ "bitbucket.org/timabell/sql-data-viewer/sqlite"
	"log"
	"os"
)

func main() {
	licensing.Licensing()

	_, err := reader.ArgParser.ParseArgs(os.Args)
	if err != nil {
		reader.ArgParser.WriteHelp(os.Stdout)
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

	log.Printf("%s\n  %s\n  %s\n  Feeback/support/contact: <%s>\n  Driver: %s, connection name: \"%s\"\n",
		about.About.Summary(),
		licensing.CopyrightText(),
		licensing.LicenseText(),
		about.About.Email,
		*reader.Options.Driver,
		*reader.Options.ConnectionDisplayName)

	host.RunServer(reader.Options)
}

func Usage() {
	log.Print("Run with Sql Server: mssql -db \"connectiongstring\" # see https://github.com/simnalamburt/go-mssqldb for connection string options")
	log.Print("Run with postgres: pg -db \"connectiongstring\" # see https://godoc.org/github.com/lib/pq for connectionstring options")
	log.Print("Run with sqlite: sqlite -db \"path\" # see https://github.com/mattn/go-sqlite3 for more info")
}
