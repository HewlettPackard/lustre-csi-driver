# Lustre CSI Driver

## Overview

This repository provides a [Lustre](https://www.lustre.org/) Container Storage Interface ([CSI](https://github.com/container-storage-interface/spec/blob/master/spec.md)), allowing Container Orchestration (CO)
frameworks to mount and unmount Lustre filesystems to/from containers in their purview.

## Features

- **Static Provisioning** - Mount and unmount externally-created Lustre volumes within containers using Persistent
  Volumes ([PV](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)) and Persistent Volume Claims 
  ([PVC](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#PersistentVolumeClaim:~:text=PersistentVolumeClaim%20(PVC))).

## Kubernetes Compatibility Matrix

| Lustre CSI Driver / Kubernetes Version | v1.12 | v1.13 | v1.14 | v1.15 | v1.16 | v1.17 | v1.18+ |
|----------------------------------------|-------|-------|-------|-------|-------|-------|--------|
| v0.0.1                                 | no    | yes   | yes   | yes   | yes   | yes   | yes    |


## Deployment

### Helm

You can use Helm to manage the lustre CSI driver components:
- To deploy: `cd charts/ && helm install lustre-csi-driver lustre-csi-driver/ --values lustre-csi-driver/values.yaml`
- To shut down: `helm delete lustre-csi-driver`

### Kind

Assuming the nnf-sos kind cluster is created...

Deploy the {lustre,mock} DaemonSet and CSIDriver on the NNF Nodes

```bash
./node.sh create {lustre,mock}
```

## Examples

Deploy an example mock pod with volume mount /mnt/nnf

```bash
kustomize build config/examples/mock | kubectl create -f -
```

## Manual Testing

Start NNF Driver:

```bash
CSI_ENDPOINT=tcp://127.0.0.1:10000 ./nnf-csi-driver
```

Test NNF Driver using csc: `https://github.com/rexray/gocsi/tree/master/csc`

Get plugin info:

```bash
csc identity plugin-info --endpoint tcp://127.0.0.1:10000
"nnf-csi-driver" "v0.0.1"
```

### NodePublish

```bash
csc node publish --cap ACCESS_MODE,ACCESS_TYPE[,FS_TYPE,MOUNT_FLAGS] --target-path TARGET_PATH VOLUME_ID [VOLUME_ID...]
```

Example

```bash
csc node publish --cap MULTI_NODE_MULTI_WRITER,mount,lustre --target-path=/mnt/fs1 rabbit-dev-01@tcp:/fs1 --endpoint tcp://127.0.0.1:10000
```
