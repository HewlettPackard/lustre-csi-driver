# Kind Configuration

This directory contains Kustomization overlays for a [Kubernetes Kind](https://kind.sigs.k8s.io/) environment,
allowing the lustre CSI driver to be started in mock mode for testing.

A patch is applied to the container runtime argument for the `csi-node-driver` in [plugin.yaml](../../base/plugin.yaml), 
setting `--driver=lustre` to `driver=mock`.