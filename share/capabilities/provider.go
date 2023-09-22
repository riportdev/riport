package capabilities

import (
	"github.com/riportdev/riport/server/chconfig"
	chshare "github.com/riportdev/riport/share"
	"github.com/riportdev/riport/share/models"
)

func NewServerCapabilities(cfg *chconfig.MonitoringConfig) *models.Capabilities {
	caps := models.Capabilities{
		ServerVersion:      chshare.BuildVersion,
		MonitoringVersion:  chshare.MonitoringVersion,
		IPAddressesVersion: chshare.IPAddressesVersion,
	}

	if !cfg.Enabled {
		caps.MonitoringVersion = 0
	}
	return &caps
}
