package options

import (
	"bitbucket.org/timabell/sql-data-viewer/reader"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func init() {
}

type SseOptions struct {
	Driver                *string `short:"d" long:"driver" description:"Driver to use" choice:"mssql" choice:"mysql" choice:"pg" choice:"sqlite" env:"schemaexplorer_driver"`
	Live                  *bool   `short:"l" long:"live" description:"update html templates & schema information from disk on every page load" env:"schemaexplorer_live"`
	ConnectionDisplayName *string `short:"n" long:"display-name" description:"A display name for this connection" env:"schemaexplorer_display_name"`
	ListenOnAddress       *string `short:"a" long:"listen-on-address" description:"address to listen on" default:"localhost" env:"schemaexplorer_listen_on_address"` // localhost so that it's secure by default, only listen for local connections
	ListenOnPort          *int    `short:"p" long:"listen-on-port" description:"port to listen on" env:"schemaexplorer_listen_on_port"`
	PeekConfigPath        *string `long:"peek-config-path" description:"path to peek configuration file" env:"schemaexplorer_peek_config_path"`
}

var Options = &SseOptions{}

//var ArgParser = flags.NewParser(Options, flags.Default)

func SetupArgs() {
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

	var driverStrings []string
	for _, driver := range reader.Drivers {
		driverStrings = append(driverStrings, driver.Name)
	}
	supportedDrivers := strings.Join(driverStrings, ", ")
	driver = flag.String("driver", "", "Driver to use. Available drivers: "+supportedDrivers)
	port = flag.Int("listen-on-port", 0, "Port to listen on. Defaults to random unused high-number.")
	address = flag.String("listen-on-address", "", "Address to listen on. Set to 0.0.0.0 to allow access to schema-explorer from other computers. Listens on localhost by default only allow connections from this machine.")
	live = flag.Bool("live", false, "Update html templates & schema information on from every page load.")
	name = flag.String("display-name", "", "A display name for this connection.")
	peekPath = flag.String("peek-config-path", "", "Path to peek configuration file. Defaults to the file included with schema explorer.")
}

var driver *string
var port *int
var address *string
var live *bool
var name *string
var peekPath *string

func ReadArgs() {
	if os.Getenv("schemaexplorer_driver") != "" {
		envDriver := os.Getenv("schemaexplorer_driver")
		Options.Driver = &envDriver
	}
	if os.Getenv("schemaexplorer_listen_on_port") != "" {
		envPort := os.Getenv("schemaexplorer_listen_on_port")
		portInt64, err := strconv.ParseInt(envPort, 0, 0)
		if err != nil {
			panic(err)
		}
		portInt := int(portInt64)
		Options.ListenOnPort = &portInt
	}
	if os.Getenv("schemaexplorer_live") != "" {
		envLive := os.Getenv("schemaexplorer_live")
		boolLive, err := strconv.ParseBool(envLive)
		if err != nil {
			panic(err)
		}
		Options.Live = &boolLive
	}
	if os.Getenv("schemaexplorer_display_name") != "" {
		envName := os.Getenv("schemaexplorer_display_name")
		Options.ConnectionDisplayName = &envName
	}
	if os.Getenv("schemaexplorer_peek_config_path") != "" {
		envPeek := os.Getenv("schemaexplorer_Peek")
		Options.PeekConfigPath = &envPeek
	}

	for _, driver := range reader.Drivers {
		for key, driverOpt := range driver.NewOptions {
			flag.StringVar(driverOpt.Value, fmt.Sprintf("%s-%s", driver.Name, key), "", driverOpt.Description)
		}
	}

	flag.Parse()

	for _, driver := range reader.Drivers {
		for key, driverOpt := range driver.NewOptions {
			if *driverOpt.Value != "" {
				continue // command line flags take precedence over environment
			}
			envKey := fmt.Sprintf("schemaexplorer_%s_%s", driver.Name, strings.Replace(key, "-", "_", 0))
			if os.Getenv(envKey) != "" {
				envValue := os.Getenv(envKey)
				*driverOpt.Value = envValue
			}
		}
	}

	if Options.Driver == nil && *driver != "" {
		Options.Driver = driver
	}
	if Options.ListenOnPort == nil {
		Options.ListenOnPort = port
	}
	if Options.ListenOnAddress == nil && *address != "" {
		Options.ListenOnAddress = address
	}
	if Options.Live == nil {
		Options.Live = live
	}
	if Options.ConnectionDisplayName == nil {
		Options.ConnectionDisplayName = name
	}
	if Options.PeekConfigPath == nil {
		Options.PeekConfigPath = peekPath
	}
}
