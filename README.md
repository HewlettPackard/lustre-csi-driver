# Lustre CSI Driver

- [Overview](#overview)
- [Features](#features)
- [Kubernetes Compatibility Matrix](#kubernetes-compatibility-matrix)
- [Deployment](#deployment)
  * [Helm](#helm)
  * [Kubernetes](#kubernetes)
  * [Kind](#kind)
- [Usage](#usage)

## Overview

This repository provides a [Lustre](https://www.lustre.org/) Container Storage Interface ([CSI](https://github.com/container-storage-interface/spec/blob/master/spec.md)), allowing Container Orchestration (CO)
frameworks to mount and unmount Lustre filesystems to/from containers in their purview.

## Features

- **Static Provisioning** - Mount and unmount externally-created Lustre volumes within containers using Persistent
  Volumes ([PV](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)) and Persistent Volume Claims 
  ([PVC](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#PersistentVolumeClaim:~:text=PersistentVolumeClaim%20(PVC))).

## Kubernetes Compatibility Matrix

| Lustre CSI Driver / Kubernetes Version | v1.13-1.18 | v1.25 | v1.27 | v1.28-v1.29 | v1.32
|----------------------------------------|------------|-------|-------|------|-----
| v0.0.10 | yes | yes | yes
| v0.1.0  |     |     |     | yes | yes


## Deployment

This describes methods of deploying the Lustre CSI driver in various environments.

### Helm

You can use Helm to manage the lustre CSI driver components:
- To deploy: `cd charts/ && helm install lustre-csi-driver lustre-csi-driver/ --values lustre-csi-driver/values.yaml`
- To shut down: `helm delete lustre-csi-driver`

### Kubernetes

Deployment uses [Kustomize](https://kustomize.io/) to configure the deployment YAMLs in the [kustomization base](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#bases-and-overlays) 
[deploy/kubernetes/base](./deploy/kubernetes/base).
- To deploy using the Makefile: `make deploy`
- To undeploy using the Makefile: `make undeploy`

To deploy a specific [overlay](./deploy/kubernetes/overlays):
- `make deploy OVERLAY=overlays/<overlay>`

Otherwise, you can just use the pre-built .yaml files in [deploy/kubernetes](./deploy/kubernetes):
- `kubectl apply -k 'https://github.com/HewlettPackard/lustre-csi-driver.git/deploy/kubernetes/overlays/kind/?ref=master'`

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

1. Checkout the project at the commit you wish to release
2. Create a local annotated tag: `git tag -a <tag> -m <message>`
4. Push this tag to remote: `git push origin <tag>`
   - This will trigger a package build with the `<tag>` version
5. Go to [GitHub releases](https://github.com/HewlettPackard/lustre-csi-driver/releases) and **Draft a New Release**
6. Use the `tag` corresponding to the release and fill out Title/Features/Bugfixes/etc.
