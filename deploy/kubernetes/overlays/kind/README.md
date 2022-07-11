# Kind Configuration

This directory contains Kustomization overlays for a [Kubernetes Kind](https://kind.sigs.k8s.io/) environment,
allowing the lustre CSI driver to be started in mock mode for testing.

Patches are applied to the [base configuration files](../../base), to delete the following volumeMounts and volumes:
- `/dev`
- `/mnt`
- `/bin`
- `/sbin`
- `/usr/bin`
- `/usr/sbin`
- `/usr/lib`
- `/usr/lib64`

These are only used by a production lustre driver and may not be present in a testing environment.
A patch is also applied to the container runtime argument for the `csi-node-driver`, setting `--driver=lustre` to 
`driver=mock`.