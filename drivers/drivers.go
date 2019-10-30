package drivers

import "github.com/timabell/schema-explorer/driver_interface"

type Driver struct {
	Name         string
	FullName     string
	CreateReader CreateReader // factory method for creating this driver's DbReader implementation
	Options      DriverOpts
}

// The list of options a driver supports
// Key is the name of the option
type DriverOpts map[string]DriverOpt

type DriverOpt struct {
	Description string  // set by the driver and used to build UI - user friendly explanation of this option
	Value       *string // set by the UI for use by the driver - the actual configured value once the application has been configured in some way
}

var Drivers = make(map[string]*Driver)

type CreateReader func() driver_interface.DbReader
