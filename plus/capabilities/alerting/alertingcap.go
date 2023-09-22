package alertingcap

import (
	"plugin"

	"github.com/riportdev/riport/plus/validator"
	"github.com/riportdev/riport/share/logger"
)

const (
	InitAlertingServiceCapabilityEx = "InitAlertingServiceCapabilityEx"
)

type Capability struct {
	Provider CapabilityEx

	Config *Config
	*logger.Logger
}

func (cap *Capability) GetInitFuncName() (name string) {
	return InitAlertingServiceCapabilityEx
}

func (cap *Capability) InitProvider(sym plugin.Symbol) {
	fn := sym.(func(cap *Capability) (capProvider CapabilityEx))
	cap.Provider = fn(cap)
}

func (cap *Capability) GetAlertingCapabilityEx() (capEx CapabilityEx) {
	return cap.Provider
}

func (cap *Capability) GetConfigValidator() (v validator.Validator) {
	return nil
}
