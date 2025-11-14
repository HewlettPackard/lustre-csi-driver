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
	"bytes"
	"context"
	"flag"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

func TestParseEndpoint(t *testing.T) {
	// Valid unix domain socket endpoint
	sockType, addr, err := ParseEndpoint("unix://fake.sock")
	require.NoError(t, err)
	assert.Equal(t, "unix", sockType)
	assert.Equal(t, "fake.sock", addr)

	sockType, addr, err = ParseEndpoint("unix:///fakedir/fakedir/fake.sock")
	require.NoError(t, err)
	assert.Equal(t, "unix", sockType)
	assert.Equal(t, "/fakedir/fakedir/fake.sock", addr)

	// Valid unix domain socket with uppercase
	sockType, addr, err = ParseEndpoint("UNIX://fake.sock")
	require.NoError(t, err)
	assert.Equal(t, "UNIX", sockType)
	assert.Equal(t, "fake.sock", addr)

	// Valid TCP endpoint with ip
	sockType, addr, err = ParseEndpoint("tcp://127.0.0.1:80")
	require.NoError(t, err)
	assert.Equal(t, "tcp", sockType)
	assert.Equal(t, "127.0.0.1:80", addr)

	// Valid TCP endpoint with uppercase
	sockType, addr, err = ParseEndpoint("TCP://127.0.0.1:80")
	require.NoError(t, err)
	assert.Equal(t, "TCP", sockType)
	assert.Equal(t, "127.0.0.1:80", addr)

	// Valid TCP endpoint with hostname
	sockType, addr, err = ParseEndpoint("tcp://fakehost:80")
	require.NoError(t, err)
	assert.Equal(t, "tcp", sockType)
	assert.Equal(t, "fakehost:80", addr)

	_, _, err = ParseEndpoint("unix:/fake.sock/")
	require.Error(t, err)

	_, _, err = ParseEndpoint("fake.sock")
	require.Error(t, err)

	_, _, err = ParseEndpoint("unix://")
	require.Error(t, err)

	_, _, err = ParseEndpoint("://")
	require.Error(t, err)

	_, _, err = ParseEndpoint("")
	require.Error(t, err)
}

func TestLogGRPC(t *testing.T) {
	// SET UP
	klog.InitFlags(nil)
	if e := flag.Set("logtostderr", "false"); e != nil {
		t.Error(e)
	}
	if e := flag.Set("alsologtostderr", "false"); e != nil {
		t.Error(e)
	}
	if e := flag.Set("v", "100"); e != nil {
		t.Error(e)
	}
	flag.Parse()

	buf := new(bytes.Buffer)
	klog.SetOutput(buf)

	handler := func(_ context.Context, _ any) (any, error) { return nil, nil }
	info := grpc.UnaryServerInfo{
		FullMethod: "fake",
	}

	tests := []struct {
		name   string
		req    any
		expStr string
	}{
		{
			"with secrets",
			&csi.NodeStageVolumeRequest{
				VolumeId: "vol_1",
				Secrets: map[string]string{
					"account_name": "k8s",
					"account_key":  "testkey",
				},
			},
			`GRPC request: {"secrets":"***stripped***","volume_id":"vol_1"}`,
		},
		{
			"without secrets",
			&csi.ListSnapshotsRequest{
				StartingToken: "testtoken",
			},
			`GRPC request: {"starting_token":"testtoken"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// EXECUTE
			_, err := logGRPC(context.Background(), test.req, &info, handler)
			require.NoError(t, err)
			klog.Flush()

			// ASSERT
			assert.Contains(t, buf.String(), "GRPC call: fake")
			assert.Contains(t, buf.String(), test.expStr)
			assert.Contains(t, buf.String(), "GRPC response: null")

			// CLEANUP
			buf.Reset()
		})
	}
}

func TestNewControllerServiceCapability(t *testing.T) {
	tests := []struct {
		cap csi.ControllerServiceCapability_RPC_Type
	}{
		{
			cap: csi.ControllerServiceCapability_RPC_UNKNOWN,
		},
		{
			cap: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		},
		{
			cap: csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		},
		{
			cap: csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
		},
		{
			cap: csi.ControllerServiceCapability_RPC_GET_CAPACITY,
		},
	}
	for _, test := range tests {
		resp := NewControllerServiceCapability(test.cap)
		assert.NotNil(t, resp)
		assert.Equal(t, test.cap, resp.GetRpc().GetType())
	}
}

func TestNewNodeServiceCapability(t *testing.T) {
	tests := []struct {
		cap csi.NodeServiceCapability_RPC_Type
	}{
		{
			cap: csi.NodeServiceCapability_RPC_UNKNOWN,
		},
		{
			cap: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
		},
		{
			cap: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
		},
		{
			cap: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
		},
	}
	for _, test := range tests {
		resp := NewNodeServiceCapability(test.cap)
		assert.NotNil(t, resp)
		assert.Equal(t, test.cap, resp.GetRpc().GetType())
	}
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		method string
		level  int32
	}{
		{
			method: "/csi.v1.Identity/Probe",
			level:  6,
		},
		{
			method: "/csi.v1.Node/NodeGetCapabilities",
			level:  6,
		},
		{
			method: "/csi.v1.Node/NodeGetVolumeStats",
			level:  6,
		},
		{
			method: "",
			level:  2,
		},
		{
			method: "unknown",
			level:  2,
		},
	}

	for _, test := range tests {
		level := getLogLevel(test.method)
		if level != test.level {
			t.Errorf("returned level: (%v), expected level: (%v)", level, test.level)
		}
	}
}
