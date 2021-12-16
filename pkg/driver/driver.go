package driver

import (
	"github.com/rexray/gocsi"
)

type DriverApi interface {
	// The name of the driver
	Name() string

	// The storage provider allocator method of the driver
	Provider() func() gocsi.StoragePluginProvider
}
