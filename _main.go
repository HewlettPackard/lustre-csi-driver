/*
 * Copyright 2021, 2022 Hewlett Packard Enterprise Development LP
 * Other additional copyright holders may be indicated within.
 *
 * The entirety of this work is licensed under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 *
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"flag"

	"github.com/rexray/gocsi"

	"github.com/HewlettPackard/lustre-csi-driver/pkg/driver"
	"github.com/HewlettPackard/lustre-csi-driver/pkg/lustre-driver"
	"github.com/HewlettPackard/lustre-csi-driver/pkg/mock-driver"
)

// main is ignored when this package is built as a go plug-in.
func main() {
	var d = flag.String("driver", "lustre", "the Lustre CSI driver to execute: [\"lustre\", \"mock\"]")
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
