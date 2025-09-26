# Lustre CSI Driver

- [Overview](#overview)
- [Features](#features)
- [Kubernetes Compatibility Matrix](#kubernetes-compatibility-matrix)
- [Deployment](#deployment)
  - [Helm](#helm)
  - [Kubernetes](#kubernetes)
  - [Kind](#kind)
- [Usage](#usage)

## Overview

This repository provides a [Lustre](https://www.lustre.org/) Container Storage Interface ([CSI](https://github.com/container-storage-interface/spec/blob/master/spec.md)), allowing Container Orchestration (CO)
frameworks to mount and unmount Lustre filesystems to/from containers in their purview.

## Features

- **Static Provisioning** - Mount and unmount externally-created Lustre volumes within containers using Persistent
  Volumes ([PV](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)) and Persistent Volume Claims 
  ([PVC](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#PersistentVolumeClaim:~:text=PersistentVolumeClaim%20(PVC))).

## Kubernetes Compatibility Matrix

| Lustre CSI Driver / Kubernetes Version | v1.29-v1.34
|----------------------------------------|------------
| v0.1.8+  | yes

## Deployment

This describes methods of deploying the Lustre CSI driver in various environments.

### Helm

You can use Helm to manage the lustre CSI driver components:

- To pick a release: `git tag`. Then pick a tag with `git checkout $RELEASE_TAG`
- To deploy: `cd charts/ && helm install lustre-csi-driver lustre-csi-driver/ --values lustre-csi-driver/values.yaml`
- To shut down: `helm delete lustre-csi-driver`

For a development build, to install a specific image tag, use the following:

- `helm install lustre-csi-driver lustre-csi-driver/ --values lustre-csi-driver/values.yaml --set deployment.tag=0.0.0.126-4fee`

### Kubernetes

Deployment uses [Kustomize](https://kustomize.io/) to configure the deployment YAMLs in the [kustomization base](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#bases-and-overlays)
[deploy/kubernetes/base](./deploy/kubernetes/base).

- To deploy using the Makefile: `make deploy`
- To undeploy using the Makefile: `make undeploy`

To deploy a specific [overlay](./deploy/kubernetes/overlays):

- `make deploy OVERLAY=overlays/<overlay>`

### Kind

This assumes your [Kind](https://kind.sigs.k8s.io/) environment is already set up and ready for a deployment.
A Kind [kustomization overlay](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#bases-and-overlays) is defined by the YAMLs in [deploy/kubernetes/overlays/kind](./deploy/kubernetes/overlays/kind).

- To deploy using the Makefile: `make kind-push && make kind-deploy`
- To undeploy using the Makefile: `make kind-undeploy`

## Usage

This section provides examples for consuming a Lustre filesystem via a Kubernetes [PersistentVolume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
(PV) and [PersistentVolumeClaim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#lifecycle-of-a-volume-and-claim) (PVC),
and finally an example of using the PVC in a simple application deployed as a Pod.

It assumed that a Lustre filesystem is already created, and that the Lustre CSI
driver is deployed on your Kubernetes cluster wherever the application pods are running (see [Deployment](#deployment) for instructions).

Inspect the `example_*.yaml` Kubernetes resources under [deploy/kubernetes/base](./deploy/kubernetes/base), then:

1. Update [example_pv.yaml](./deploy/kubernetes/base/example_pv.yaml)'s `volumeHandle` value to the NID list of your Lustre filesystem's MGS.
2. Deploy the PV:  `kubectl apply -f deploy/kubernetes/base/example_pv.yaml`
3. Deploy the PVC: `kubectl apply -f deploy/kubernetes/base/example_pvc.yaml`
4. Deploy the app: `kubectl apply -f deploy/kubernetes/base/example_app.yaml`
   - Note: The lustre filesystem defaults to being mounted at `/mnt/lus` within the container. Update this in example_app.yaml if you desire a different location.

## Steps for Releasing a Version

To perform a release, please use the tools and documentation described in [Releasing NNF Software](https://nearnodeflash.github.io/latest/repo-guides/release-nnf-sw/release-all/#nnf-software-overview). The steps and tools in that guide will ensure that the new release is properly configured to self-identify and to package properly with new releases of the NNF software stack.

The following old steps can be used if this project is ever disassociated from the NNF software stack:

1. Checkout the project at the commit you wish to release
2. Create a local annotated tag: `git tag -a <tag> -m <message>`
3. Push this tag to remote: `git push origin <tag>`
   - This will trigger a package build with the `<tag>` version
4. Go to [GitHub releases](https://github.com/HewlettPackard/lustre-csi-driver/releases) and **Draft a New Release**
5. Use the `tag` corresponding to the release and fill out Title/Features/Bugfixes/etc.
