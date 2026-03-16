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

## Local Testing, Without a Real Lustre Filesystem

The PV's `.spec.csi.volumeHandle` should refer to a Lustre nid, per usual. Then, commandline arguments may be specified to tell the driver to mount a local filesystem instead of the Lustre filesystem. This swap of volume arguments will happen immediately prior to calling into the `mount` library routine.

For example, maybe on the node where the CSI driver is running, and where the example application will run, there is an XFS filesystem on `/dev/vdb` which is not yet mounted. Edit the DaemonSet to add the `--swap-source-*` commandline arguments to the `csi-node-driver` container:

```console
kubectl edit ds -n lustre-csi-system lustre-csi-node
```

Note the `--swap-source-from` argument matches the `.spec.csi.volumeHandle` specified in the PV, and that the local XFS filesystem which is not yet mounted is on `/dev/vdb`. This is a DaemonSet, so these args will apply to all nodes:

```yaml
      containers:
      - args:
        - -v=5
        - --endpoint=$(CSI_ENDPOINT)
        - --nodeid=$(KUBE_NODE_NAME)
        - --swap-source-from=10.1.1.113@tcp:/lushtx
        - --swap-source-to=/dev/vdb
        - --swap-source-to-fstype=xfs
```

The application will refer to the original PV which specifies the Lustre nid, but will end up with the /dev/vdb XFS filesystem mounted, instead.

## Read-Only Mount

When considering read-only mounts, recall that on a single host, Linux does not allow the same volume to be mounted "rw" on one mountpoint and "ro" on another mountpoint.

Details:

- Pod `.spec.volumes[].persistentVolumeClaim.readOnly`:
  Volume is mounted with "ro" mount option. This affects all containers in the pod. CRI-O knows it's read-only and doesn't attempt the selinux relabel.

- Pod `.spec.containers[].volumeMounts[].readOnly`:
  Volume is mounted with "rw" mount option. But it's read-only in this individual container.

- PV `.spec.csi.readOnly`:
  This is passed to the ControllerPublishVolumeRequest endpoint in the CSI driver. This CSI driver does not support this endpoint.

- PV `.spec.mountOptions`
  Additional mount options. Supported with csi, iscsi, and nfs. If "ro" is specified, then the volume is mounted with "ro" mount option. CRI-O doesn't know it's read-only and wants to do the selinux relabel, but cannot write to the volume, and it fails to setup the container.

- PVC does not have an equivalent of PV's `.spec.mountOptions`.

- PV `.spec.accessModes` does not control or constrain the mount options. This is used to
advise the k8s scheduler about pod placement.

- PVC `.spec.accessModes` is loosely used to match a PV. The PV access mode is what matters.


## OpenShift Lustre Client Install
RedHat CoreOS (RHCOS) is an immutable OS that requires the use of a Kernel Module Manager (KMM) operator to orchestrate the building and insertion of 3rd party kernel modules (i.e. Lustre client kernel modules). So, users will need to install a KMM operator prior to following these instructions. Additionally, the KMM build must include entitlement keys associated with your Red Hat subscription to access required RHEL content during the build process.

### Including entitlement keys as build secret:
Copy the entitlement secret from the openshift-config-managed namespace to the namespace of the build (openshift-kmm):

```console
$ cat << EOF > secret-template.txt
kind: Secret
apiVersion: v1
metadata:
  name: etc-pki-entitlement
type: Opaque
data: {{ range \$key, \$value := .data }}
  {{ \$key }}: {{ \$value }} {{ end }}
EOF
$ oc get secret etc-pki-entitlement -n openshift-config-managed -o=go-template-file --template=secret-template.txt | oc apply -f
```

### Create Builder Image:
Builder images are created using 2 primary commands: `oc new-build` and `oc start-build`. Where `new-build` generates a BuildConfig used in `start-build` to define and create an image it pushes to registry: `image-registry.openshift-image-registry.svc:5000/<namespace>/<image-name>:<tag>`. In our case, the image will be packaged with the Lustre source code and all required build dependencies. The Driver ToolKit (DTK) is an ideal base image for our purposes and comes packaged with commonly required dependencies to build or install kernel modules. To define a DTK as the base image we will need to leverage the docker build strategy and provide a *Dockerfile*:

Example *Dockerfile*:
```yaml
    # DTK build image for OpenShift 4.X (RHEL 8) latest:
    FROM registry.redhat.io/openshift4/driver-toolkit-rhel8:latest

    # Install Lustre client && lnet build dependencies
    RUN dnf install -y \
    gcc \
    make \
    binutils \
    rpm-build \
    kernel-devel \
    kernel-headers \
    elfutils-libelf-devel \
    autoconf \
    automake \
    libtool \
    rdma-core-devel \
    libibverbs-devel \
    && dnf clean all

# Copy the lustre-release git repo (build context) into the image
COPY . /opt/app-root/src
WORKDIR /opt/app-root/src
```
In the following section we show two methods for how this *Dockerfile* can be used. For the *Binary source* build option, this Dockerfile is copied into the root of the Lustre source tree (i.e. /lustre-release/) where the docker build strategy automatically detects and applies the file. Alternatively, for the git source build it is executed as a inline dockerfile. 

*NOTE:* Verify you’re in the “openshift-kmm” namespace prior to creating the build image: `oc project openshift-kmm` 

There are multiple ways to include application source into a build image. Below, are **two** of the most common methods:

1.	*Binary (local) source*:
      - Clone the Lustre source and checkout relevant branch:
         ```console
         git clone -b <branch> https://github.com/lustre/lustre-release.git
         ```
      - Copy Lustre source code into designated build directory (referenced by --from-dir).
      - Copy *Dockerfile* (from previous section) into the root of the Lustre project (lustre-release).
      - Create new BuildConfig and configure it to take binary source as input utilizing the docker strategy. 
         ```console
         oc new-build --binary --name= <build-name> -n openshift-kmm --strategy=docker
         ```
      - Start build: 
         - update with build directory containing lustre source:
         ```console
         oc start-build <build-name> --from-dir=/<build-dir>/ --follow
         ```

2. *Git Source*:
     - Create BuildConfig using Lustre github repo as source with relevant branch. 
        ```console
        oc new-build https://github.com/lustre/lustre-release.git#<branch> \
          --name=<build-name> \
          -n openshift-kmm \
          --strategy=docker \
          --dockerfile="$(cat </path/to/Dockerfile>)"
        ```
     - The source definition in the BuildConfig should now contain:
         ```yaml
         source:
           git: 
             uri: "https://github.com/lustre/lustre-release.git"
             ref: "branch-name" 
         ``` 
     - You can check the contents of the BuildConfig by executing:
           `oc describe bc <build-name> -n openshift-kmm`
    - Start build:
      ```console
      oc start-build <build-name> --follow
      ```
Now the builder image should be successfully built and pushed to OpenShift’s internal registry. 

### Configure KMM:
To build and deploy the module you'll need to provide both a ConfigMap and Module Custom Resource Definition (CRD) to KMM. The ConfigMap instructs KMM on how to build the lustre-client and the Module CRD defines various required build parameters. 

*Example ConfigMap*: (Update \<place-holders\>)
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: <name>
  namespace: openshift-kmm
data:
  dockerfile: |
    # Replace <image-name> with relevant image name from previous section:
    FROM image-registry.openshift-image-registry.svc:5000/openshift-kmm/<image-name>:latest AS builder
    ARG KERNEL_VERSION

    # lustre-release source location:  
    WORKDIR /opt/app-root/src
    
    # Configure and build Lustre RPMs
    # NOTE: Uncomment --with-o2ib if using InfiniBand
    RUN ./configure \
        --with-linux=/lib/modules/${KERNEL_VERSION}/build \
        --with-linux-obj=/lib/modules/${KERNEL_VERSION}/build \
        --disable-server \
        --disable-tests \
        --enable-client \
        # --with-o2ib
    
    # Build RPMs
    RUN make rpms
    
    # Now package Lustre modules in minimal image for KMM
    FROM registry.access.redhat.com/ubi8/ubi-minimal
    
    ARG KERNEL_VERSION
    
    # Install kmod and rpm utilities
    RUN microdnf install -y kmod rpm
    
    # Copy built RPMs from builder stage into ubi8 min image
    COPY --from=builder /opt/app-root/src/*.rpm /tmp/
    
    # Install both kernel module and userspace client RPMs
    RUN rpm -ivh /tmp/lustre-client-*.rpm /tmp/kmod-lustre-client-*.rpm || \
    (rpm2cpio /tmp/kmod-lustre-client-*.rpm | cpio -idmv)
    
    # Copy Lustre/LNet modules to KMM expected location
    # Note: The following 'find' statement might need to be modified 
    # depending on where kernel mods where installed on your system. 
    RUN mkdir -p /opt/lib/modules/${KERNEL_VERSION}/extra && \
        find /lib/modules/${KERNEL_VERSION} \( -path "*/lustre/*.ko" -o -path "*/lnet/*.ko" -o -name "ko2iblnd.ko" -o -path "*/lustre-client/*.ko" \) -exec cp {} /opt/lib/modules/${KERNEL_VERSION}/extra/ \;
    
    # Generate module dependencies
    RUN depmod -b /opt ${KERNEL_VERSION}
```

*Template Module CRD*: (Update \<place-holders\>)
```yaml
apiVersion: kmm.sigs.x-k8s.io/v1beta1
kind: Module
metadata:
  name: <name>
  namespace: openshift-kmm
spec:
  moduleLoader:
    container:
      modprobe:
        moduleName: lustre
      kernelMappings:
      - regexp: '^.+$'
        containerImage: image-registry.openshift-image-registry.svc:5000/openshift-kmm/<image-name>:${KERNEL_VERSION}
        build:
          dockerfileConfigMap:
            name: <configmap>
          buildArgs:
          - name: KERNEL_VERSION
            value: ${KERNEL_VERSION}
  selector:
    node-role.kubernetes.io/worker: ""
```

### Configure LNet (InfiniBand/o2ib)

If you're using InfiniBand in your OpenShift cluster you'll also need to configure LNet once Lustre client modules have been successfully loaded. Below is an example DaemonSet that applies the runtime LNet network configuration (for example: binding `o2ib` to `ib0`) on each node.

- Determine your node kernel version:

   ```console
   oc debug node/<worker-node-name> -- chroot /host uname -r
   ```

- Create a ServiceAccount and allow it to run privileged pods (required for `hostNetwork` + `modprobe` + `lnetctl`):

   ```console
   oc create sa lnet-configuration -n openshift-kmm
   oc adm policy add-scc-to-user privileged -z lnet-configuration -n openshift-kmm
   ```

- Apply the DaemonSet below.
   Replace `<KERNEL_VERSION>` with the output from `uname -r`.

   *Example DaemonSet* (Update \<place-holders\>)
   ```yaml
   apiVersion: apps/v1
   kind: DaemonSet
   metadata:
     name: lnet-configuration
     namespace: openshift-kmm
   spec:
     selector:
       matchLabels:
         app: lnet-configuration
     template:
       metadata:
         labels:
           app: lnet-configuration
       spec:
         hostNetwork: true
         nodeSelector:
           node-role.kubernetes.io/worker: ""
         tolerations:
         - operator: Exists
         serviceAccountName: lnet-configuration
         containers:
         - name: lnet-configuration
           image: image-registry.openshift-image-registry.svc:5000/openshift-kmm/<image-name>:<KERNEL_VERSION>
           imagePullPolicy: IfNotPresent
           securityContext:
             privileged: true
             runAsUser: 0
             allowPrivilegeEscalation: true
             seLinuxOptions:
               type: spc_t
           env:
           - name: NET_TYPE
             value: "o2ib"
           - name: NET_IFACE
             value: <Network Interface name (i.e. ib0)>
           command:
           - /bin/sh
           - -c
           - |
             set -eu
             echo "Waiting for interface ${NET_IFACE}..."
             while [ ! -d "/sys/class/net/${NET_IFACE}" ]; do
               sleep 5
             done

             # Load LNet/Lustre dependencies (best-effort)
             modprobe ko2iblnd || true
             modprobe lnet || true

             # Initialize LNet and add o2ib network (safe to run repeatedly)
             lnetctl lnet configure || true
             lnetctl net add --net "${NET_TYPE}" --if "${NET_IFACE}" || true

             echo "LNet configuration complete; sleeping"
             trap : TERM INT
             while true; do sleep 3600; done
   ```


