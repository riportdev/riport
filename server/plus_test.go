package chserver

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rportplus "github.com/riportdev/riport/plus"
	"github.com/riportdev/riport/plus/capabilities/oauth"
	"github.com/riportdev/riport/plus/validator"
	"github.com/riportdev/riport/server/chconfig"
	"github.com/riportdev/riport/share/logger"
)

const (
	defaultPluginPath = "../riport-plus/riport-plus.so"
)

var defaultValidMinServerConfig = chconfig.ServerConfig{
	URL:          []string{"http://localhost/"},
	DataDir:      "./",
	Auth:         "abc:def",
	UsedPortsRaw: []string{"10-20"},
}

type mockValidator struct{}

func (m *mockValidator) ValidateConfig() (err error) {
	return nil
}

type plusManagerMock struct {
	CapabilityCount int
	Caps            map[string]rportplus.Capability

	rportplus.ManagerProvider
}

func (pm *plusManagerMock) RegisterCapability(capName string, newCap rportplus.Capability) (cap rportplus.Capability, err error) {
	pm.CapabilityCount++
	if pm.Caps == nil {
		pm.Caps = make(map[string]rportplus.Capability, 0)
	}
	pm.Caps[capName] = newCap
	return newCap, nil
}

func (pm *plusManagerMock) GetConfigValidator(capName string) (v validator.Validator) {
	return &mockValidator{}
}

// Checks that the expected plugins are loaded using using mock interfaces.
// Does not require a working plugin.
func TestShouldRegisterPlusCapabilities(t *testing.T) {
	plusLog := logger.NewLogger("riport-plus", logger.LogOutput{File: os.Stdout}, logger.LogLevelDebug)

	config := &chconfig.Config{
		Server: defaultValidMinServerConfig,
		PlusConfig: rportplus.PlusConfig{
			PluginConfig: &rportplus.PluginConfig{
				PluginPath: defaultPluginPath,
			},
			OAuthConfig: &oauth.Config{
				Provider: oauth.GitHubOAuthProvider,
			},
		},
	}

	plus := &plusManagerMock{}
	plus.InitPlusManager(&config.PlusConfig, nil, plusLog)
	require.NotNil(t, plus)

	// register the capabilities with the plus manager partial mock. the purpose
	// of the test is to check whether the expected capabilities are being
	// requested, not to test the plugin manager.
	err := RegisterPlusCapabilities(plus, config, testLog)
	assert.NoError(t, err)

	// this check will flag when additional capabilities have been registered but the test
	// not updated
	assert.Equal(t, 5, plus.CapabilityCount)

	// additional capabilities should be checked here to see that the server has
	// registered them
	assert.NotNil(t, plus.Caps[rportplus.PlusOAuthCapability])
	assert.NotNil(t, plus.Caps[rportplus.PlusStatusCapability])
	assert.NotNil(t, plus.Caps[rportplus.PlusAlertingCapability])
}
