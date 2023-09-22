package chshare

import (
	"encoding/json"
	"fmt"

	"github.com/riportdev/riport/share/clientconfig"
	"github.com/riportdev/riport/share/models"
)

// ConnectionRequest represents configuration options when initiating client-server connection
type ConnectionRequest struct {
	ID                     string
	Name                   string
	SessionID              string
	OS                     string
	OSFullName             string
	OSVersion              string
	OSVirtualizationSystem string
	OSVirtualizationRole   string
	OSArch                 string
	OSFamily               string
	OSKernel               string
	Version                string
	Hostname               string
	CPUFamily              string
	CPUModel               string
	CPUModelName           string
	CPUVendor              string
	NumCPUs                int
	MemoryTotal            uint64
	Timezone               string
	IPv4                   []string
	IPv6                   []string
	Tags                   []string
	Labels                 map[string]string
	Remotes                []*models.Remote
	ClientConfiguration    *clientconfig.Config
}

func DecodeConnectionRequest(b []byte) (*ConnectionRequest, error) {
	c := &ConnectionRequest{}
	err := json.Unmarshal(b, c)
	if err != nil {
		return nil, fmt.Errorf("Invalid JSON config")
	}
	return c, nil
}

func EncodeConnectionRequest(c *ConnectionRequest) ([]byte, error) {
	return json.Marshal(c)
}
