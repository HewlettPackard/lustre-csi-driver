/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/HewlettPackard/lustre-csi-driver/pkg/hpelustre"
	"k8s.io/klog/v2"
)

var NnfDriverName = "lustre-csi.hpe.com"

var (
	endpoint                 = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID                   = flag.String("nodeid", "", "node id")
	version                  = flag.Bool("version", false, "Print the version and exit.")
	driverName               = flag.String("drivername", NnfDriverName, "name of the driver")
	enableHpeLustreMockMount = flag.Bool("enable-hpelustre-mock-mount", false, "Whether enable mock mount(only for testing)")
	workingMountDir          = flag.String("working-mount-dir", "/tmp", "working directory for provisioner to mount lustre filesystems temporarily")
)

func main() {
	klog.InitFlags(nil)
	err := flag.Set("logtostderr", "true")
	if err != nil {
		klog.Fatalln(err)
	}
	flag.Parse()
	if *version {
		info, err := hpelustre.GetVersionYAML(*driverName)
		if err != nil {
			klog.Fatalln(err)
		}
		klog.V(2).Info(info)
		fmt.Println(info) //nolint:forbidigo // Print version info to stdout for access through kubectl exec
		os.Exit(0)
	}

	handle()
	os.Exit(0)
}

func handle() {
	driverOptions := hpelustre.DriverOptions{
		NodeID:                   *nodeID,
		DriverName:               *driverName,
		EnableHpeLustreMockMount: *enableHpeLustreMockMount,
		WorkingMountDir:          *workingMountDir,
	}
	driver := hpelustre.NewDriver(&driverOptions)
	if driver == nil {
		klog.Fatalln("Failed to initialize HPE Lustre CSI driver")
	}
	driver.Run(*endpoint, false)
}
