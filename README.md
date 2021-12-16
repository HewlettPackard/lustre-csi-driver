# NNF CSI Driver

CSI Driver for Near-Node Flash

## Deployment

Assuming the nnf-sos kind cluster is created...

Deploy the {lustre,mock} DaemonSet and CSIDriver on the NNF Nodes
```
$ ./node.sh create {lustre,mock}
```

## Examples

Deploy an example mock pod with volume mount /mnt/nnf
```
$ kustomize build config/examples/mock | kubectl create -f -
```

# Manual Testing

## Start NNF Driver
```
$ CSI_ENDPOINT=tcp://127.0.0.1:10000 ./nnf-csi-driver
```

## Test NNF Driver using csc
Get csc tool from https://github.com/rexray/gocsi/tree/master/csc

### Get plugin info
```
$ csc identity plugin-info --endpoint tcp://127.0.0.1:10000
"nnf-csi-driver"	"v0.0.1"
```


### NodePublish
```
$ csc node publish --cap ACCESS_MODE,ACCESS_TYPE[,FS_TYPE,MOUNT_FLAGS] --target-path TARGET_PATH VOLUME_ID [VOLUME_ID...]
```
For example
```
$ csc node publish --cap MULTI_NODE_MULTI_WRITER,mount,lustre --target-path=/mnt/fs1 rabbit-dev-01@tcp:/fs1 --endpoint tcp://127.0.0.1:10000
```