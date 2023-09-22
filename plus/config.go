package rportplus

import (
	"github.com/riportdev/riport/plus/capabilities/oauth"
	"github.com/riportdev/riport/plus/license"
)

// PluginConfig contains the config related to the plugin itself
type PluginConfig struct {
	PluginPath string `mapstructure:"plugin_path"`
}

// PlusConfig contains the overall config for the riport-plus plugin. note that
// each capability should have it's own config section in the config file.
type PlusConfig struct {
	PluginConfig  *PluginConfig   `mapstructure:"plus-plugin"`
	OAuthConfig   *oauth.Config   `mapstructure:"plus-oauth"`
	LicenseConfig *license.Config `mapstructure:"plus-license"`
}
