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
	"flag"
	"log"
	"os"
	"strconv"
)

func main() {
	licensing.EnforceLicensing()

	//_, err := options.ArgParser.ParseArgs(os.Args)
	//if err != nil {
	//	// only write out help if not already being written
	//	if e, ok := err.(*flags.Error); !ok || e.Type != flags.ErrHelp {
	//		options.ArgParser.WriteHelp(os.Stdout)
	//	}
	//	os.Stdout.WriteString("\n")
	//	os.Stdout.WriteString("Environment Variables:\n")
	//	os.Stdout.WriteString("  Environment variables can be used instead of of the command line arguments.\n")
	//	os.Stdout.WriteString("  The environment variable names can be found at the end of each option's\n")
	//	os.Stdout.WriteString("  description above in the form [$env_var_name].\n")
	//	os.Stdout.WriteString("  These can then be set with: env_var_name=value\n")
	//	os.Stdout.WriteString("  <3 <3 <3 Because we love https://12factor.net/config <3 <3 <3\n")
	//	os.Stdout.WriteString("\n")
	//	os.Exit(1)
	//}

	if os.Getenv("schemaexplorer_driver") != "" {
		envDriver := os.Getenv("schemaexplorer_driver")
		options.Options.Driver = &envDriver
	}
	if os.Getenv("schemaexplorer_listen_on_port") != "" {
		envPort := os.Getenv("schemaexplorer_listen_on_port")
		portInt64, err := strconv.ParseInt(envPort, 0, 0)
		if err != nil {
			panic(err)
		}
		portInt := int(portInt64)
		options.Options.ListenOnPort = &portInt
	}
	if os.Getenv("schemaexplorer_live") != "" {
		envLive := os.Getenv("schemaexplorer_live")
		boolLive, err := strconv.ParseBool(envLive)
		if err != nil {
			panic(err)
		}
		options.Options.Live = &boolLive
	}
	if os.Getenv("schemaexplorer_display_name") != "" {
		envName := os.Getenv("schemaexplorer_display_name")
		options.Options.ConnectionDisplayName = &envName
	}
	if os.Getenv("schemaexplorer_peek_config_path") != "" {
		envPeek := os.Getenv("schemaexplorer_Peek")
		options.Options.PeekConfigPath = &envPeek
	}

	driver := flag.String("driver", "", "Driver to use") // todo: list loaded drivers
	port := flag.Int("listen-on-port", 0, "Port to listen on. Defaults to random unused high-number.")
	address := flag.String("listen-on-address", "", "Address to listen on. Set to 0.0.0.0 to allow access to schema-explorer from other computers. Listens on localhost by default only allow connections from this machine.")
	live := flag.Bool("live", false, "Update html templates & schema information on from every page load.")
	name := flag.String("display-name", "", "A display name for this connection.")
	peekPath := flag.String("peek-config-path", "", "Path to peek configuration file. Defaults to the file included with schema explorer.")

	flag.Parse()
	if options.Options.Driver == nil && *driver != "" {
		options.Options.Driver = driver
	}
	if options.Options.ListenOnPort == nil {
		options.Options.ListenOnPort = port
	}
	if options.Options.ListenOnAddress == nil && *address != "" {
		options.Options.ListenOnAddress = address
	}
	if options.Options.Live == nil {
		options.Options.Live = live
	}
	if options.Options.ConnectionDisplayName == nil {
		options.Options.ConnectionDisplayName = name
	}
	if options.Options.PeekConfigPath == nil {
		options.Options.PeekConfigPath = peekPath
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

	serve.RunServer()
}
