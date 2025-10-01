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
	"fmt"
	"strings"
	"sync"

	csicommon "github.com/HewlettPackard/lustre-csi-driver/pkg/csi-common"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog/v2"
	mount "k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
)

const (
	DefaultLustreFsName = "lustrefs"
	separator           = "#"

	podNameKey            = "csi.storage.k8s.io/pod.name"
	podNamespaceKey       = "csi.storage.k8s.io/pod.namespace"
	podUIDKey             = "csi.storage.k8s.io/pod.uid"
	serviceAccountNameKey = "csi.storage.k8s.io/serviceaccount.name"
	pvcNameKey            = "csi.storage.k8s.io/pvc/name"
	pvcNamespaceKey       = "csi.storage.k8s.io/pvc/namespace"
	pvNameKey             = "csi.storage.k8s.io/pv/name"

	podNameMetadata            = "${pod.metadata.name}"
	podNamespaceMetadata       = "${pod.metadata.namespace}"
	podUIDMetadata             = "${pod.metadata.uid}"
	serviceAccountNameMetadata = "${serviceAccount.metadata.name}"
	pvcNameMetadata            = "${pvc.metadata.name}"
	pvcNamespaceMetadata       = "${pvc.metadata.namespace}"
	pvNameMetadata             = "${pv.metadata.name}"
)

var (
	controllerServiceCapabilities = []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
	}

	volumeCapabilities = []csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_SINGLE_WRITER,
		csi.VolumeCapability_AccessMode_SINGLE_NODE_MULTI_WRITER,
		csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
		csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	}

	nodeServiceCapabilities = []csi.NodeServiceCapability_RPC_Type{
		csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
		csi.NodeServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
	}
)

type lustreVolume struct {
	name          string
	id            string
	mgsIPAddress  string
	hpeLustreName string
	subDir        string
}

// DriverOptions defines driver parameters specified in driver deployment
type DriverOptions struct {
	DriverName               string
	EnableHpeLustreMockMount bool
	WorkingMountDir          string
}

// Driver implements all interfaces of CSI drivers
type Driver struct {
	csicommon.CSIDriver
	csicommon.DefaultIdentityServer
	csicommon.DefaultControllerServer
	csicommon.DefaultNodeServer
	// enableHpeLustreMockMount is only for testing, DO NOT set as true in non-testing scenario
	enableHpeLustreMockMount bool
	mounter                  *mount.SafeFormatAndMount
	forceMounter             *mount.MounterForceUnmounter
	// Directory to temporarily mount to for subdirectory creation
	workingMountDir  string
	kernelModuleLock sync.Mutex
}

// NewDriver Creates a NewCSIDriver object. Assumes vendor version is equal to driver version &
// does not support optional driver plugin info manifest field. Refer to CSI spec for more details.
func NewDriver(options *DriverOptions) *Driver {
	d := Driver{
		enableHpeLustreMockMount: options.EnableHpeLustreMockMount,
		workingMountDir:          options.WorkingMountDir,
	}
	d.Name = options.DriverName
	d.Version = driverVersion

	d.DefaultControllerServer.Driver = &d.CSIDriver
	d.DefaultIdentityServer.Driver = &d.CSIDriver
	d.DefaultNodeServer.Driver = &d.CSIDriver

	return &d
}

// Run driver initialization
func (d *Driver) Run(endpoint string, testBool bool) {
	versionMeta, err := GetVersionYAML(d.Name)
	if err != nil {
		klog.Fatalf("%v", err)
	}
	klog.Infof("\nDRIVER INFORMATION:\n-------------------\n%s\n\nStreaming logs below:", versionMeta)

	d.mounter = &mount.SafeFormatAndMount{
		Interface: mount.New(""),
		Exec:      utilexec.New(),
	}
	forceUnmounter, ok := d.mounter.Interface.(mount.MounterForceUnmounter)
	if ok {
		klog.V(4).Infof("Using force unmounter interface")
		d.forceMounter = &forceUnmounter
	} else {
		klog.Fatalf("Mounter does not support force unmount")
	}

	// TODO_JUSJIN: revisit these caps
	// Initialize default library driver
	// TODO_CHYIN: move this to {service}.go
	d.AddControllerServiceCapabilities(controllerServiceCapabilities)
	d.AddVolumeCapabilityAccessModes(volumeCapabilities)
	d.AddNodeServiceCapabilities(nodeServiceCapabilities)

	s := csicommon.NewNonBlockingGRPCServer()
	// Driver d act as IdentityServer, ControllerServer and NodeServer
	s.Start(endpoint, d, d, d, testBool)
	s.Wait()
}

func IsCorruptedDir(dir string) bool {
	_, pathErr := mount.PathExists(dir)
	return pathErr != nil && mount.IsCorruptedMnt(pathErr)
}

func getLustreVolFromID(id string) (*lustreVolume, error) {
	segments := strings.Split(id, separator)
	if len(segments) < 3 {
		return nil, fmt.Errorf("could not split volume ID %q into lustre name and ip address", id)
	}

	name := segments[0]
	vol := &lustreVolume{
		name:          name,
		id:            id,
		hpeLustreName: DefaultLustreFsName,
		mgsIPAddress:  segments[2],
	}

	if len(segments) >= 4 {
		vol.subDir = strings.Trim(segments[3], "/")
	}

	return vol, nil
}
