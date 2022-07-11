# Helm Charts

This directory contains charts to deploy the lustre CSI driver with Helm.
This deploys the [CSIDriver](lustre-csi-driver/templates/driver.yaml), 
[DaemonSet](lustre-csi-driver/templates/plugin.yaml), and [Namespace](lustre-csi-driver/templates/namespace.yaml)
Kubernetes resources.

## Usage

- To deploy: `helm install lustre-csi-driver lustre-csi-driver/ --values lustre-csi-driver/values.yaml`
- To shut down: `helm delete lustre-csi-driver`