package options

import "github.com/jessevdk/go-flags"

func init() {
	ArgParser.EnvNamespace = "schemaexplorer"
	ArgParser.NamespaceDelimiter = "-"
}

type SseOptions struct {
	Driver                *string `short:"d" long:"driver" description:"Driver to use" choice:"mssql" choice:"pg" choice:"sqlite" env:"schemaexplorer_driver"`
	Live                  *bool   `short:"l" long:"live" description:"update html templates & schema information from disk on every page load" env:"schemaexplorer_live"`
	ConnectionDisplayName *string `short:"n" long:"display-name" description:"A display name for this connection" env:"schemaexplorer_display_name"`
	ListenOnAddress       *string `short:"a" long:"listen-on-address" description:"address to listen on" default:"localhost" env:"schemaexplorer_listen_on_address"` // localhost so that it's secure by default, only listen for local connections
	ListenOnPort          *int    `short:"p" long:"listen-on-port" description:"port to listen on" env:"schemaexplorer_listen_on_port"`
	PeekConfigPath        *string `long:"peek-config-path" description:"path to peek configuration file" default:"config/peek-config.txt" env:"schemaexplorer_peek_config_path"`
}

var Options = &SseOptions{}
var ArgParser = flags.NewParser(Options, flags.Default)
