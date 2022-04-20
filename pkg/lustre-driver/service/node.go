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

package service

import (
	"os"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/mount"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func (s *service) NodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest) (
	*csi.NodeStageVolumeResponse, error) {

	return nil, nil
}

func (s *service) NodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest) (
	*csi.NodeUnstageVolumeResponse, error) {

	return nil, nil
}

func (s *service) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	// 1. Validate request
	if req.GetVolumeId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "NodePublishVolume - VolumeID is required")
	}

	if req.GetTargetPath() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "NodePublishVolume - TargetPath is required")
	}

	// ??? req.GetVolumeCapability()
	// TODO: Check the FsType is supported by the driver

	// This targetpath is deep down in /var/lib/kubelet/pods/....
	// As k8s starts a pod that references this FS, that pod will have
	// a spec.containers[].volumeMounts that tells k8s where to bind mount
	// it into the pod's namespace.
	err := os.MkdirAll(req.GetTargetPath(), 0755)
	if err != nil && err != os.ErrExist {
		return nil, status.Errorf(codes.Internal, "NodePublishVolume - Mountpoint mkdir Failed: Error %v", err)
	}

	// 2. Perform the mount
	mounter := mount.New("")
	err = mounter.Mount(
		req.GetVolumeId(),
		req.GetTargetPath(),
		req.GetVolumeCapability().GetMount().GetFsType(),
		req.GetVolumeCapability().GetMount().GetMountFlags())

	if err != nil {
		return nil, status.Errorf(codes.Internal, "NodePublishVolume - Mount Failed: Error %v", err)
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (s *service) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	if req.GetVolumeId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "NodeUnpublishVolume - VolumeID is required")
	}

	if req.GetTargetPath() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "NodeUnpublishVolume - TargetPath is required")
	}

	mounter := mount.New("")
	err := mounter.Unmount(req.GetTargetPath())

	if err != nil {
		return nil, status.Errorf(codes.Internal, "NodeUnpublishVolume - Unmount Failed: Error %v", err)
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (s *service) NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest) (
	*csi.NodeGetVolumeStatsResponse, error) {

	return nil, nil
}

func (s *service) NodeExpandVolume(
	ctx context.Context,
	req *csi.NodeExpandVolumeRequest) (
	*csi.NodeExpandVolumeResponse, error) {

	return nil, nil
}

func (s *service) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	/*
		return &csi.NodeGetCapabilitiesResponse{
			Capabilities: []*csi.NodeServiceCapability{
				{
					Type: &csi.NodeServiceCapability_Rpc{
						Rpc: &csi.NodeServiceCapability_RPC{
							Type: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
						},
					},
				},
			},
		}, nil
	*/
	return &csi.NodeGetCapabilitiesResponse{}, nil
}

func (s *service) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest) (
	*csi.NodeGetInfoResponse, error) {

	return &csi.NodeGetInfoResponse{
		NodeId: os.Getenv("KUBE_NODE_NAME"),
	}, nil
}
