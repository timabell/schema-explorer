package options

import (
	"github.com/timabell/schema-explorer/drivers"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func init() {
}

type SseOptions struct {
	Driver                string
	Live                  bool
	ConnectionDisplayName string
	ListenOnAddress       string
	ListenOnPort          string
	PeekConfigPath        string
}

var Options = &SseOptions{}

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
	for _, driver := range drivers.Drivers {
		driverStrings = append(driverStrings, driver.Name)
	}
	supportedDrivers := strings.Join(driverStrings, ", ")
	flag.StringVar(&Options.Driver, "driver", "", "Driver to use. Available drivers: "+supportedDrivers)
	flag.StringVar(&Options.ListenOnPort, "listen-on-port", "", "Port to listen on. Defaults to random unused high-number.")
	flag.StringVar(&Options.ListenOnAddress, "listen-on-address", "", "Address to listen on. Set to 0.0.0.0 to allow access to schema-explorer from other computers. Listens on localhost by default only allow connections from this machine.")
	flag.BoolVar(&Options.Live, "live", false, "Update html templates & schema information on from every page load. (Row counts and data are always updated).")
	flag.StringVar(&Options.ConnectionDisplayName, "display-name", "", "A display name for this connection.")
	flag.StringVar(&Options.PeekConfigPath, "peek-config-path", "", "Path to peek configuration file. Defaults to the file included with schema explorer.")

	for _, driver := range drivers.Drivers {
		for key, driverOpt := range driver.Options {
			flag.StringVar(driverOpt.Value, fmt.Sprintf("%s-%s", driver.Name, key), "", driverOpt.Description)
		}
	}
}

func ReadArgsAndEnv() {
	flag.Parse()

	if Options.Driver == "" && os.Getenv("schemaexplorer_driver") != "" {
		envDriver := os.Getenv("schemaexplorer_driver")
		Options.Driver = envDriver
	}
	if Options.ListenOnAddress == "" && os.Getenv("schemaexplorer_listen_on_address") != "" {
		envAddress := os.Getenv("schemaexplorer_listen_on_address")
		Options.ListenOnAddress = envAddress
	}
	if Options.ListenOnPort == "" && os.Getenv("schemaexplorer_listen_on_port") != "" {
		envPort := os.Getenv("schemaexplorer_listen_on_port")
		Options.ListenOnPort = envPort
	}
	if !Options.Live && os.Getenv("schemaexplorer_live") != "" {
		envLive := os.Getenv("schemaexplorer_live")
		boolLive, err := strconv.ParseBool(envLive)
		if err != nil {
			panic(err)
		}
		Options.Live = boolLive
	}
	if Options.ConnectionDisplayName == "" && os.Getenv("schemaexplorer_display_name") != "" {
		envName := os.Getenv("schemaexplorer_display_name")
		Options.ConnectionDisplayName = envName
	}
	if Options.PeekConfigPath == "" && os.Getenv("schemaexplorer_peek_config_path") != "" {
		envPeek := os.Getenv("schemaexplorer_Peek")
		Options.PeekConfigPath = envPeek
	}

	for _, driver := range drivers.Drivers {
		for key, driverOpt := range driver.Options {
			if *driverOpt.Value != "" {
				continue // command line flags take precedence over environment
			}
			envKey := fmt.Sprintf("schemaexplorer_%s_%s", driver.Name, strings.Replace(key, "-", "_", -1))
			if os.Getenv(envKey) != "" {
				envValue := os.Getenv(envKey)
				*driverOpt.Value = envValue
			}
		}
	}
}

func (options SseOptions) IsConfigured() bool {
	return options.Driver != ""
}
