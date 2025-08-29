module github.com/HewlettPackard/lustre-csi-driver

go 1.23.9

require (
	github.com/container-storage-interface/spec v1.5.0
	github.com/rexray/gocsi v1.2.2
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/net v0.33.0
	google.golang.org/grpc v1.26.0
	k8s.io/mount-utils v0.24.2
)

require google.golang.org/protobuf v1.26.0 // indirect

require (
	github.com/akutz/gosync v0.1.0 // indirect
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/utils v0.0.0-20230406110748-d93618cff8a2 // indirect
)
