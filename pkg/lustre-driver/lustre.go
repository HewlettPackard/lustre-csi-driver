package lustre

import (
	"github.com/rexray/gocsi"

	"github.hpe.com/hpe/hpc-rabsw-lustre-csi-driver/pkg/driver"
	"github.hpe.com/hpe/hpc-rabsw-lustre-csi-driver/pkg/lustre-driver/provider"
	"github.hpe.com/hpe/hpc-rabsw-lustre-csi-driver/pkg/lustre-driver/service"
)

func NewLustreDriver() driver.DriverApi {
	return &lustreDriver{}
}

type lustreDriver struct{}

func (lustreDriver) Name() string                                 { return service.Name }
func (lustreDriver) Provider() func() gocsi.StoragePluginProvider { return provider.New }
