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

package hpelustre

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	volumehelper "github.com/HewlettPackard/lustre-csi-driver/pkg/util"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/volume"
	mount "k8s.io/mount-utils"
)

// NodePublishVolume mount the volume from staging to target path
func (d *Driver) NodePublishVolume(
	_ context.Context,
	req *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {

	volCap := req.GetVolumeCapability()
	if volCap == nil {
		return nil, status.Error(codes.InvalidArgument,
			"Volume capability missing in request")
	}
	userMountFlags := volCap.GetMount().GetMountFlags()
	volumeType := volCap.GetMount().GetFsType()

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}

	target := req.GetTargetPath()
	if len(target) == 0 {
		return nil, status.Error(codes.InvalidArgument,
			"Target path not provided")
	}

	context := req.GetVolumeContext()
	if context == nil {
		return nil, status.Error(codes.InvalidArgument,
			"Volume context must be provided")
	}

	vol, err := getVolume(volumeID, context)
	if err != nil {
		return nil, err
	}

	//source := getSourceString(vol.mgsIPAddress, vol.hpeLustreName)
	source := volumeID

	mountOptions, readOnly := getMountOptions(req, userMountFlags)

	if len(vol.subDir) > 0 && !d.enableHpeLustreMockMount {
		interpolatedSubDir := interpolateSubDirVariables(context, vol)

		if isSubpath := ensureStrictSubpath(interpolatedSubDir); !isSubpath {
			return nil, status.Error(
				codes.InvalidArgument,
				"Context sub-dir must be strict subpath",
			)
		}

		if readOnly {
			klog.V(2).Info("NodePublishVolume: not attempting to create sub-dir on read-only volume, assuming existing path")
		} else {
			klog.V(2).Infof(
				"NodePublishVolume: sub-dir will be created at %q",
				interpolatedSubDir,
			)

			if err = d.createSubDir(vol, target, interpolatedSubDir, mountOptions); err != nil {
				return nil, err
			}
		}

		source = filepath.Join(source, interpolatedSubDir)
		klog.V(2).Infof(
			"NodePublishVolume: full mount source with sub-dir: %q",
			source,
		)
	}

	mnt, err := d.ensureMountPoint(target)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"Could not mount target %q: %v",
			target,
			err)
	}
	if mnt {
		klog.V(2).Infof(
			"NodePublishVolume: volume %s is already mounted on %s",
			volumeID,
			target,
		)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	//klog.V(2).Infof(
	//	"NodePublishVolume: volume %s mounting %s at %s with mountOptions: %v",
	//	volumeID, source, target, mountOptions,
	//)
	klog.V(2).Infof(
		"NodePublishVolume: volume %s mounting at %s, type %s, with mountOptions: %v",
		source, target, volumeType, mountOptions,
	)
	if d.enableHpeLustreMockMount {
		klog.Warningf(
			"NodePublishVolume: mock mount on volumeID(%s), this is only for"+
				"TESTING!!!",
			volumeID,
		)
		if err := volumehelper.MakeDir(target); err != nil {
			klog.Errorf("MakeDir failed on target: %s (%v)", target, err)
			return nil, err
		}
		return &csi.NodePublishVolumeResponse{}, nil
	}

	err = mountVolumeAtPath(d, source, target, volumeType, mountOptions)
	if err != nil {
		if removeErr := os.Remove(target); removeErr != nil {
			return nil, status.Errorf(
				codes.Internal,
				"Could not remove mount target %q: %v",
				target,
				removeErr,
			)
		}
		return nil, status.Errorf(codes.Internal,
			"Could not mount %q at %q: %v", source, target, err)
	}

	//klog.V(2).Infof(
	//	"NodePublishVolume: volume %s mount %s at %s successfully",
	//	volumeID,
	//	source,
	//	target,
	//)
	klog.V(2).Infof(
		"NodePublishVolume: volume mount %s at %s successfully",
		source,
		target,
	)

	return &csi.NodePublishVolumeResponse{}, nil
}

func interpolateSubDirVariables(context map[string]string, vol *lustreVolume) string {
	subDirReplaceMap := map[string]string{}

	// get metadata values
	for k, v := range context {
		switch strings.ToLower(k) {
		case podNameKey:
			subDirReplaceMap[podNameMetadata] = v
		case podNamespaceKey:
			subDirReplaceMap[podNamespaceMetadata] = v
		case podUIDKey:
			subDirReplaceMap[podUIDMetadata] = v
		case serviceAccountNameKey:
			subDirReplaceMap[serviceAccountNameMetadata] = v
		case pvcNamespaceKey:
			subDirReplaceMap[pvcNamespaceMetadata] = v
		case pvcNameKey:
			subDirReplaceMap[pvcNameMetadata] = v
		case pvNameKey:
			subDirReplaceMap[pvNameMetadata] = v
		}
	}

	interpolatedSubDir := volumehelper.ReplaceWithMap(vol.subDir, subDirReplaceMap)
	return interpolatedSubDir
}

func getMountOptions(req *csi.NodePublishVolumeRequest, userMountFlags []string) ([]string, bool) {
	readOnly := false
	mountOptions := []string{}
	if req.GetReadonly() {
		readOnly = true
		mountOptions = append(mountOptions, "ro")
	}
	for _, userMountFlag := range userMountFlags {
		if userMountFlag == "ro" {
			readOnly = true

			if req.GetReadonly() {
				continue
			}
		}
		mountOptions = append(mountOptions, userMountFlag)
	}
	return mountOptions, readOnly
}

func getVolume(volumeID string, context map[string]string) (*lustreVolume, error) {
	return &lustreVolume{}, nil
}

func xx_getVolume(volumeID string, context map[string]string) (*lustreVolume, error) {
	volName := ""

	volFromID, err := getLustreVolFromID(volumeID)
	if err != nil {
		klog.Warningf("error parsing volume ID '%v'", err)
	} else {
		volName = volFromID.name
	}

	vol, err := newLustreVolume(volumeID, volName, context)
	if err != nil {
		return nil, err
	}

	if volFromID != nil && *volFromID != *vol {
		klog.Warningf("volume context does not match values in volume ID for volumeID %v", volumeID)
	}

	return vol, nil
}

func mountVolumeAtPath(d *Driver, source, target string, volumeType string, mountOptions []string) error {
	d.kernelModuleLock.Lock()
	defer d.kernelModuleLock.Unlock()
	klog.Warningf("DEANDEAN mountoptions are: %v", mountOptions)
	klog.Warningf("DEANDEAN source is: %s", source)
	klog.Warningf("DEANDEAN target is: %s", target)
	klog.Warningf("DEANDEAN volumeType is: %s", volumeType)
	if d.swapSourceFrom != "" && source == d.swapSourceFrom {
		klog.Warningf("Swapping PV source '%s' to '%s' (%s) in mountVolumeAtPath", d.swapSourceFrom, d.swapSourceTo, d.swapSourceToType)
		source = d.swapSourceTo
		volumeType = d.swapSourceToType
	}
	klog.Warningf("DEANDEAN 2 source is: %s", source)
	klog.Warningf("DEANDEAN 2 volumeType is: %s", volumeType)
	err := d.mounter.MountSensitiveWithoutSystemdWithMountFlags(
		source,
		target,
		volumeType, // "lustre",
		mountOptions,
		nil,
		[]string{"--no-mtab"},
	)
	return err
}

// NodeUnpublishVolume unmount the volume from the target path
func (d *Driver) NodeUnpublishVolume(
	_ context.Context,
	req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}

	targetPath := req.GetTargetPath()
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument,
			"Target path missing in request")
	}

	klog.V(2).Infof("NodeUnpublishVolume: unmounting volume %s on %s",
		volumeID, targetPath)
	err := unmountVolumeAtPath(d, targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"failed to unmount target %q: %v", targetPath, err)
	}
	klog.V(2).Infof(
		"NodeUnpublishVolume: unmount volume %s on %s successfully",
		volumeID,
		targetPath,
	)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func unmountVolumeAtPath(d *Driver, targetPath string) error {
	shouldUnmountBadPath := false

	d.kernelModuleLock.Lock()
	defer d.kernelModuleLock.Unlock()

	parent := filepath.Dir(targetPath)
	klog.V(2).Infof("Listing dir: %s", parent)
	entries, err := os.ReadDir(parent)
	if err != nil {
		klog.Warningf("could not list directory %s, will explicitly unmount path before cleanup %s: %q", parent, targetPath, err)
		shouldUnmountBadPath = true
	}

	for _, e := range entries {
		if e.Name() == filepath.Base(targetPath) {
			_, err := e.Info()
			if err != nil {
				klog.Warningf("could not get info for entry %s, will explicitly unmount path before cleanup %s: %q", e.Name(), targetPath, err)
				shouldUnmountBadPath = true
			}
		}
	}

	if shouldUnmountBadPath {
		// In these cases, if we only ran mount.CleanupMountWithForce,
		// it would have issues trying to stat the directory before
		// cleanup, so we need to explicitly unmount the path, with
		// force if necessary. Then the directory can be cleaned up
		// by the mount.CleanupMountWithForce call.
		klog.V(4).Infof("unmounting bad mount: %s)", targetPath)
		forceUnmounter := *d.forceMounter
		if err := forceUnmounter.UnmountWithForce(targetPath, 30*time.Second); err != nil {
			klog.Warningf("couldn't unmount %s: %q", targetPath, err)
		}
	}

	err = mount.CleanupMountWithForce(targetPath, *d.forceMounter,
		true /*extensiveMountPointCheck*/, 10*time.Second)
	return err
}

// Staging and Unstaging is not able to be supported with how Lustre is mounted
//
// This was discovered during a proof of concept implementation and the issue
// is as follows:
//
// When the kubelet process attempts to unstage / unmount a Lustre mount that
// has been staged to a global mount point, it performs extra checks to ensure
// that the same device is not mounted anywhere else in the filesystem. For
// usual configurations, this would be a reasonable check to ensure that we
// aren't trying to remove something that is still in use elsewhere in the
// system. However, the way Lustre mounts are configured is not compatible
// with the check it performs.
//
// The kubelet process does this by checking all of the mount points on the
// node to see if any have the following:
// 1) The same 'root' value of the mount that is being cleaned
// 2) The same device number of the mount that is being cleaned
// And that those mounts are in a different path tree.
// If so, it returns this error: "the device mount path %q is still mounted
// by other references %v", deviceMountPath, refs) and fails the unmount.
// See pkg/volume/util/operationexecutor/operation_generator.go
// calling GetDeviceMountRefs(deviceMountPath) around line 947.
//
// All Lustre mounts on a system, no matter where in the lustrefs they are
// mounted to, all have '/' as the root and they all have the same major and
// minor device numbers, so as far as this check is concerned, every lustre
// mount is the same device, even though individual Lustre mount points can
// be unmounted without affecting others and should not be a concern.
//
// With a single Lustre volume mount, this works fine. It stages to a
// globalpath dir, pods can bind mount into that, and when the last pod is
// done, unstage is called and the global mount point can be cleaned up,
// because that is the only lustre mount so kubelet has no issue with
// 'other mounts' on the same node.
//
// The problem occurs when two different volumes are trying to mount a
// Lustre cluster. In that case, pods for the first volume can come up
// as expected with their global mount path, then pods for the second
// volume with their global mount path. The error occurs when the pods
// for one of these volumes are deleted and an unstage action should occur,
// because the other volume has its own Lustre mount, so it fails this
// check. For example, it's trying to unmount
// /var/...<firstvolume>.../globalpath, but there's another volume at
// /var/...<secondvolume>.../globalpath with the same root '/' and major
// and minor device numbers.
//
// It errors out, fails the unmount, and never calls unstage, even
// though all of the pods using that volume have already been deleted.
// This leaves the box with as many global mount directories still mounted
// to the Lustre cluster as you've ever staged, but without any way to see
// this other than looking at the mounts on the node or in the kubelet logs.
func (d *Driver) NodeStageVolume(_ context.Context, _ *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// Staging and Unstaging is not able to be supported with how Lustre is mounted
//
// See NodeStageVolume for more details
func (d *Driver) NodeUnstageVolume(_ context.Context, _ *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// NodeGetCapabilities return the capabilities of the Node plugin
func (d *Driver) NodeGetCapabilities(
	_ context.Context, _ *csi.NodeGetCapabilitiesRequest,
) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: d.NSCap,
	}, nil
}

// NodeGetInfo return info of the node on which this plugin is running
func (d *Driver) NodeGetInfo(
	_ context.Context,
	_ *csi.NodeGetInfoRequest,
) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId: d.NodeID,
	}, nil
}

// NodeGetVolumeStats get volume stats
func (d *Driver) NodeGetVolumeStats(
	_ context.Context,
	req *csi.NodeGetVolumeStatsRequest,
) (*csi.NodeGetVolumeStatsResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument,
			"NodeGetVolumeStats volume ID was empty")
	}
	volumePath := req.GetVolumePath()
	if len(volumePath) == 0 {
		return nil, status.Error(codes.InvalidArgument,
			"NodeGetVolumeStats volume path was empty")
	}

	if _, err := os.Lstat(volumePath); err != nil {
		if os.IsNotExist(err) {
			return nil, status.Errorf(codes.NotFound,
				"path %s does not exist", volumePath)
		}
		return nil, status.Errorf(codes.Internal,
			"failed to stat file %s: %v", volumePath, err)
	}

	volumeMetrics, err := volume.NewMetricsStatFS(volumePath).GetMetrics()
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"failed to get metrics: %v", err)
	}

	available, ok := volumeMetrics.Available.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal,
			"failed to transform volume available size(%v)",
			volumeMetrics.Available)
	}
	capacity, ok := volumeMetrics.Capacity.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal,
			"failed to transform volume capacity size(%v)",
			volumeMetrics.Capacity)
	}
	used, ok := volumeMetrics.Used.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal,
			"failed to transform volume used size(%v)", volumeMetrics.Used)
	}

	inodesFree, ok := volumeMetrics.InodesFree.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal,
			"failed to transform disk inodes free(%v)",
			volumeMetrics.InodesFree)
	}
	inodes, ok := volumeMetrics.Inodes.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal,
			"failed to transform disk inodes(%v)", volumeMetrics.Inodes)
	}
	inodesUsed, ok := volumeMetrics.InodesUsed.AsInt64()
	if !ok {
		return nil, status.Errorf(codes.Internal,
			"failed to transform disk inodes used(%v)",
			volumeMetrics.InodesUsed)
	}

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Unit:      csi.VolumeUsage_BYTES,
				Available: available,
				Total:     capacity,
				Used:      used,
			},
			{
				Unit:      csi.VolumeUsage_INODES,
				Available: inodesFree,
				Total:     inodes,
				Used:      inodesUsed,
			},
		},
	}, nil
}

// ensureMountPoint: create mount point if not exists
// return <true, nil> if it's already a mounted point
// otherwise return <false, nil>
func (d *Driver) ensureMountPoint(target string) (bool, error) {
	notMnt, err := d.mounter.IsLikelyNotMountPoint(target)
	if err != nil && !os.IsNotExist(err) {
		if IsCorruptedDir(target) {
			notMnt = false
			klog.Warningf("detected corrupted mount for targetPath [%s]",
				target)
		} else {
			return !notMnt, err
		}
	}

	if !notMnt {
		// testing original mount point, make sure the mount link is valid
		_, err := os.ReadDir(target)
		if err == nil {
			klog.V(2).Infof("already mounted to target %s", target)
			return !notMnt, nil
		}
		// mount link is invalid, now unmount and remount later
		klog.Warningf("ReadDir %s failed with %v, unmount this directory",
			target, err)
		if err := d.mounter.Unmount(target); err != nil {
			klog.Errorf("Unmount directory %s failed with %v", target, err)
			return !notMnt, err
		}
		notMnt = true
		return !notMnt, err
	}
	if err := volumehelper.MakeDir(target); err != nil {
		klog.Errorf("MakeDir failed on target: %s (%v)", target, err)
		return !notMnt, err
	}
	return !notMnt, nil
}

func (d *Driver) createSubDir(vol *lustreVolume, mountPath, subDirPath string, mountOptions []string) error {
	if err := d.internalMount(vol, mountPath, mountOptions); err != nil {
		return err
	}

	defer func() {
		if err := d.internalUnmount(mountPath); err != nil {
			klog.Warningf("failed to unmount lustre server: %v", err.Error())
		}
	}()

	internalVolumePath, err := getInternalVolumePath(d.workingMountDir, mountPath, subDirPath)
	if err != nil {
		return err
	}

	klog.V(2).Infof("Making subdirectory at %q", internalVolumePath)

	if err := os.MkdirAll(internalVolumePath, 0o775); err != nil {
		return status.Errorf(codes.Internal, "failed to make subdirectory: %v", err.Error())
	}

	return nil
}

func getSourceString(mgsIPAddress, lustreName string) string {
	if lustreName[0] != '/' {
		lustreName = "/" + lustreName
	}
	var src string
	if strings.Contains(mgsIPAddress, "@") {
		src = fmt.Sprintf("%s:%s", mgsIPAddress, lustreName)
	} else {
		src = fmt.Sprintf("%s@tcp:%s", mgsIPAddress, lustreName)
	}
	return src
}

func getInternalMountPath(workingMountDir, mountPath string) (string, error) {
	mountPath = strings.Trim(mountPath, "/")

	if isSubpath := ensureStrictSubpath(mountPath); !isSubpath {
		return "", status.Errorf(
			codes.Internal,
			"invalid mount path %q",
			mountPath,
		)
	}

	return filepath.Join(workingMountDir, mountPath), nil
}

func getInternalVolumePath(workingMountDir, mountPath, subDirPath string) (string, error) {
	internalMountPath, err := getInternalMountPath(workingMountDir, mountPath)
	if err != nil {
		return "", err
	}

	if isSubpath := ensureStrictSubpath(subDirPath); !isSubpath {
		return "", status.Errorf(
			codes.InvalidArgument,
			"sub-dir %q must be strict subpath",
			subDirPath,
		)
	}

	return filepath.Join(internalMountPath, subDirPath), nil
}

func (d *Driver) internalMount(vol *lustreVolume, mountPath string, mountOptions []string) error {
	source := getSourceString(vol.mgsIPAddress, vol.hpeLustreName)

	target, err := getInternalMountPath(d.workingMountDir, mountPath)
	if err != nil {
		return err
	}

	klog.V(4).Infof("internally mounting %v", target)

	mnt, err := d.ensureMountPoint(target)
	if err != nil {
		return status.Errorf(codes.Internal,
			"Could not mount target %q: %v",
			target,
			err)
	}

	if mnt {
		klog.Warningf(
			"volume %q is already mounted on %q",
			vol.id,
			target,
		)

		err = d.internalUnmount(mountPath)
		if err != nil {
			return status.Errorf(codes.Internal,
				"Could not unmount existing volume at %q: %v",
				target,
				err)
		}
	}

	klog.V(2).Infof(
		"volume %q mounting %q at %q with mountOptions: %v",
		vol.id, source, target, mountOptions,
	)

	err = mountVolumeAtPath(d, source, target, "lustre", mountOptions)
	if err != nil {
		if removeErr := os.Remove(target); removeErr != nil {
			return status.Errorf(
				codes.Internal,
				"Could not remove mount target %q: %v",
				target,
				removeErr,
			)
		}

		return status.Errorf(codes.Internal,
			"Could not mount %q at %q: %v", source, target, err)
	}

	return nil
}

func (d *Driver) internalUnmount(mountPath string) error {
	target, err := getInternalMountPath(d.workingMountDir, mountPath)
	if err != nil {
		return err
	}

	klog.V(4).Infof("internally unmounting %v", target)

	err = mount.CleanupMountWithForce(target, *d.forceMounter, true, 10*time.Second)
	if err != nil {
		err = status.Errorf(codes.Internal, "failed to unmount staging target %q: %v", target, err)
	}

	return err
}

// Ensures that the given subpath, when joined with any base path, will be a path
// within the given base path, and not equal to it. This ensures that this
// subpath value can be safely created or deleted under the base path without
// affecting other data in the base path.
func ensureStrictSubpath(subPath string) bool {
	return filepath.IsLocal(subPath) && filepath.Clean(subPath) != "."
}

// Convert context parameters to a lustreVolume
func newLustreVolume(volumeID, volumeName string, params map[string]string) (*lustreVolume, error) {
	var mgsIPAddress, subDir string

	// validate parameters (case-insensitive).
	for k, v := range params {
		switch strings.ToLower(k) {
		case VolumeContextMGSIPAddress:
			mgsIPAddress = v
		case VolumeContextSubDir:
			subDir = v
			subDir = strings.Trim(subDir, "/")

			if len(subDir) == 0 {
				return nil, status.Error(
					codes.InvalidArgument,
					"Context sub-dir must not be empty or root if provided",
				)
			}
		}
	}

	if len(mgsIPAddress) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"Context mgs-ip-address must be provided",
		)
	}

	vol := &lustreVolume{
		name:          volumeName,
		mgsIPAddress:  mgsIPAddress,
		hpeLustreName: volumeID, // DefaultLustreFsName,
		subDir:        subDir,
		id:            volumeID,
	}

	return vol, nil
}
