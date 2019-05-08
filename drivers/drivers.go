package drivers

import "bitbucket.org/timabell/sql-data-viewer/driver_interface"

type Driver struct {
	Name         string
	FullName     string
	CreateReader CreateReader // factory method for creating this driver's DbReader implementation
	Options      interface{}  // todo: remove
	NewOptions   DriverOpts
}

// The list of options a driver supports
// Key is the name of the option
type DriverOpts map[string]DriverOpt

type DriverOpt struct {
	Description string  // set by the driver and used to build UI
	Value       *string // set by the UI for use by the driver
}

var Drivers = make(map[string]*Driver)

type CreateReader func() driver_interface.DbReader
