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

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func (s *service) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	log.Tracef(">>> NodeStageVolume: VolumeId: %s TargetPath: %s", req.GetVolumeId(), "")
	defer log.Tracef("<<< NodeStageVolume")
	return nil, nil
}

func (s *service) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	log.Tracef(">>> NodeUnstageVolume: VolumeId: %s TargetPath: %s", req.GetVolumeId(), "")
	defer log.Tracef("<<< NodeUnstageVolume")

	return nil, nil
}

func (s *service) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	log.Tracef(">>> NodePublishVolume: VolumeId: %s TargetPath: %s", req.GetVolumeId(), req.GetTargetPath())
	defer log.Tracef("<<< NodePublishVolume")

	// 1. Validate request
	if req.GetVolumeId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "NodePublishVolume - VolumeID is required")
	}

	if req.GetTargetPath() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "NodePublishVolume - TargetPath is required")
	}

	// ??? req.GetVolumeCapability()
	// TODO: Check the FsType is supported by the driver

	// 2. Perform the mount

	return &csi.NodePublishVolumeResponse{}, nil
}

func (s *service) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	log.Tracef(">>> NodeUnpublishVolume: VolumeId: %s TargetPath: %s", req.GetVolumeId(), req.GetTargetPath())
	defer log.Tracef("<<< NodeUnpublishVolume")

	// 1. Validate request
	if req.GetVolumeId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "NodeUnpublishVolume - VolumeID is required")
	}

	if req.GetTargetPath() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "NodeUnpublishVolume - TargetPath is required")
	}

	// 2. Perform the unmount

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (s *service) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	log.Tracef(">>> NodeGetVolumeStats: VolumeId: %s TargetPath: %s", req.GetVolumeId(), "")
	defer log.Tracef("<<< NodeGetVolumeStats")

	return nil, nil
}

func (s *service) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	log.Tracef(">>> NodeExpandVolume: VolumeId: %s TargetPath: %s", req.GetVolumeId(), "")
	defer log.Tracef("<<< NodeExpandVolume")

	return nil, nil
}

func (s *service) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	log.Tracef(">>> NodeGetCapabilities:")
	defer log.Tracef("<<< NodeGetCapabilities")

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_SINGLE_NODE_MULTI_WRITER,
					},
				},
			},
		},
	}, nil
}

func (s *service) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	log.Tracef(">>> NodeGetInfo: NodeId: %s", os.Getenv("KUBE_NODE_NAME"))
	defer log.Tracef("<<< NodeGetInfo")

	return &csi.NodeGetInfoResponse{
		NodeId: os.Getenv("KUBE_NODE_NAME"),
	}, nil
}
