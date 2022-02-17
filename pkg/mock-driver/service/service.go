package service

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
)

const (
	// Name is the name of this CSI SP.
	Name = "lustre-csi.nnf.cray.hpe.com"

	// VendorVersion is the version of this CSP SP.
	VendorVersion = "v0.0.1"
)

// Service is a CSI SP and idempotency.Provider.
type Service interface {
	csi.ControllerServer
	csi.IdentityServer
	csi.NodeServer
}

type service struct{}

// New returns a new Service.
func New() Service {
	return &service{}
}
