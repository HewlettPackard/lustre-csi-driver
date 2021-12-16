package mock

import (
	"github.com/rexray/gocsi"

	"github.hpe.com/hpe/hpc-rabsw-lustre-csi-driver/pkg/driver"
	"github.hpe.com/hpe/hpc-rabsw-lustre-csi-driver/pkg/mock-driver/provider"
	"github.hpe.com/hpe/hpc-rabsw-lustre-csi-driver/pkg/mock-driver/service"
)

func NewMockDriver() driver.DriverApi {
	return &mockDriver{}
}

type mockDriver struct{}

func (mockDriver) Name() string                                 { return service.Name }
func (mockDriver) Provider() func() gocsi.StoragePluginProvider { return provider.New }
