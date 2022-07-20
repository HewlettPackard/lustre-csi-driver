# Base Configuration

Provides the base configuration .yamls for a Lustre CSI driver as a [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/).
Assumes lustre client capabilities are installed on the underlying host wherever the driver containers are launched.

- **driver.yaml** - Defines a [CSIDriver](https://kubernetes-csi.github.io/docs/csi-driver-object.html) object that allows Kubernetes to discover CSI Drivers on the cluster,
and defines the driver's supported features.
- **namespace.yaml** - Defines a [Namespace](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/) resource to isolate the Lustre CSI driver resources to the `system` namespace.
- **plugin.yaml** - Defines a DaemonSet for the Lustre CSI driver container, and a sidecar registrar container.
- **example_pv.yaml** - Example [PersistentVolume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) for a lustre filesystem.
- **example_pvc.yaml** - Example [PersistentVolumeClaim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#lifecycle-of-a-volume-and-claim)
for a lustre filesystem.
- **example_app.yaml** - Example dummy application which uses a lustre filesystem through a PersistentVolumeClaim.
- **kustomization.yaml** - Config file defining resources for the [Kustomize](https://kustomize.io/) tool
