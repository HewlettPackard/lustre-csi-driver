package main

import (
	"context"
	"flag"

	"github.com/rexray/gocsi"

	"github.hpe.com/hpe/hpc-rabsw-lustre-csi-driver/pkg/driver"
	"github.hpe.com/hpe/hpc-rabsw-lustre-csi-driver/pkg/lustre-driver"
	"github.hpe.com/hpe/hpc-rabsw-lustre-csi-driver/pkg/mock-driver"
)

// main is ignored when this package is built as a go plug-in.
func main() {
	var d = flag.String("driver", "lustre", "the nnf csi driver {lustre,mock} to execute")
	flag.Parse()

	drvr := newDriver(*d)

	gocsi.Run(
		context.Background(),
		drvr.Name(),
		"A description of the SP",
		"",
		drvr.Provider()(),
	)
}

func newDriver(driver string) driver.DriverApi {
	switch driver {
	case "lustre":
		return lustre.NewLustreDriver()
	case "mock":
		return mock.NewMockDriver()
	default:
		panic("Unrecognized driver type " + driver)
	}
}
