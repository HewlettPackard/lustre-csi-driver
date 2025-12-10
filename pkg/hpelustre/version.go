/*
Copyright 2019 The Kubernetes Authors.

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

package hpelustre

import (
	"runtime"
	"strings"

	"sigs.k8s.io/yaml"
)

var (
	driverVersion = "N/A"
)

type VersionInfo struct {
	DriverName    string `json:"Driver Name"`
	DriverVersion string `json:"Driver Version"`
	GoVersion     string `json:"Go Version"`
}

func GetVersion(driverName string) VersionInfo {
	return VersionInfo{
		DriverName:    driverName,
		DriverVersion: driverVersion,
		GoVersion:     runtime.Version(),
	}
}

// GetVersionYAML returns the version information of the driver
// in YAML format.
func GetVersionYAML(driverName string) (string, error) {
	info := GetVersion(driverName)
	marshalled, err := yaml.Marshal(&info)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(marshalled)), nil
}
