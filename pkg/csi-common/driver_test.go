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

package csicommon

import (
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	fakeDriverName = "fake"
	fakeNodeID     = "fakeNodeID"
)

var vendorVersion = "0.3.0"

func TestNewCSIDriver(t *testing.T) {
	name := ""
	str := ""
	nodeID := ""
	assert.Nil(t, NewCSIDriver(name, str, nodeID))
	name = "unit-test"
	assert.Nil(t, NewCSIDriver(name, str, nodeID))
	nodeID = "unit-test"
	driver := CSIDriver{
		Name:    name,
		NodeID:  nodeID,
		Version: str,
	}
	assert.Equal(t, &driver, NewCSIDriver(name, str, nodeID))
}

func NewFakeDriver() *CSIDriver {
	driver := NewCSIDriver(fakeDriverName, vendorVersion, fakeNodeID)

	return driver
}

func TestNewFakeDriver(t *testing.T) {
	// Test New fake driver with invalid arguments.
	d := NewCSIDriver("", vendorVersion, fakeNodeID)
	assert.Nil(t, d)
}

func TestAddControllerServiceCapabilities(t *testing.T) {
	d := NewFakeDriver()
	cl := []csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_UNKNOWN}
	d.AddControllerServiceCapabilities(cl)
	assert.Len(t, d.Cap, 1)
	assert.Equal(t, csi.ControllerServiceCapability_RPC_UNKNOWN, d.Cap[0].GetRpc().GetType())
}

func TestAddNodeServiceCapabilities(t *testing.T) {
	d := NewFakeDriver()

	nl := []csi.NodeServiceCapability_RPC_Type{csi.NodeServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER}
	d.AddNodeServiceCapabilities(nl)
	assert.Len(t, d.NSCap, 1)
	assert.Equal(t, csi.NodeServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER, d.NSCap[0].GetRpc().GetType())
}

func TestGetVolumeCapabilityAccessModes(t *testing.T) {
	d := NewFakeDriver()

	// Test no volume access modes.
	// REVISIT: Do we need to support any default access modes.
	c := d.GetVolumeCapabilityAccessModes()
	assert.Empty(t, c)

	// Test driver with access modes.
	d.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})
	modes := d.GetVolumeCapabilityAccessModes()
	assert.Len(t, modes, 1)
	assert.Equal(t, csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER, modes[0].GetMode())
}

func TestValidateControllerServiceRequest(t *testing.T) {
	d := NewFakeDriver()

	// Valid requests which require no capabilities
	err := d.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_UNKNOWN)
	require.NoError(t, err)

	// Test controller service publish/unpublish not supported
	err = d.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME)
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, s.Code())

	// Add controller service publish & unpublish request
	d.AddControllerServiceCapabilities(
		[]csi.ControllerServiceCapability_RPC_Type{
			csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
			csi.ControllerServiceCapability_RPC_GET_CAPACITY,
			csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
		})

	// Test controller service publish/unpublish is supported
	err = d.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME)
	require.NoError(t, err)

	// Test controller service create/delete is supported
	err = d.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME)
	require.NoError(t, err)

	// Test controller service list volumes is supported
	err = d.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_LIST_VOLUMES)
	require.NoError(t, err)

	// Test controller service get capacity is supported
	err = d.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_GET_CAPACITY)
	require.NoError(t, err)
}
