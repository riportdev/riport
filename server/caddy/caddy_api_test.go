//go:build linux
// +build linux

package caddy_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riportdev/riport/server/caddy"
	"github.com/riportdev/riport/server/chconfig"
)

func TestShouldAddRouteToCaddyServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer withWait(cancel)

	s := setupNewCaddyServer(ctx, t)

	nrr := makeTestNewRouteRequest()

	res, err := s.AddRoute(ctx, nrr)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestShouldDeleteRouteFromCaddyServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer withWait(cancel)

	s := setupNewCaddyServer(ctx, t)

	nrr := makeTestNewRouteRequest()

	res, err := s.AddRoute(ctx, nrr)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	res, err = s.DeleteRoute(ctx, nrr.RouteID)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func makeTestNewRouteRequest() (nrr *caddy.NewRouteRequest) {
	nrr = &caddy.NewRouteRequest{
		RouteID:                   "1111",
		TargetTunnelHost:          "127.0.0.1",
		TargetTunnelPort:          "5555",
		DownstreamProxySubdomain:  "1111",
		DownstreamProxyBaseDomain: "tunnels.rpdev",
	}

	return nrr
}

func setupNewCaddyServer(ctx context.Context, t *testing.T) (cs *caddy.Server) {
	t.Helper()

	cfg := &caddy.Config{
		ExecPath:         "/usr/bin/caddy",
		DataDir:          "/tmp",
		BaseConfFilename: "caddy-base.conf",
		HostAddress:      "0.0.0.0:8443",
		BaseDomain:       "tunnels.riport.test",
		CertFile:         "../../testdata/certs/tunnels.riport.test.crt",
		KeyFile:          "../../testdata/certs/tunnels.riport.test.key",
	}

	chCfg := &chconfig.Config{
		API: chconfig.APIConfig{
			Address:  "0.0.0.0:3000",
			CertFile: cfg.CertFile,
			KeyFile:  cfg.KeyFile,
		},
	}

	if !caddyAvailable(t, cfg) {
		t.Skip("caddy server not available")
	}

	bc, err := chCfg.WriteCaddyBaseConfig(cfg)
	require.NoError(t, err)
	caddy.HostDomainSocket = bc.GlobalSettings.AdminSocket

	// ensure no existing admin socket file
	os.Remove(caddy.HostDomainSocket)

	cs = caddy.NewCaddyServer(cfg, testLog)
	err = cs.Start(ctx)
	require.NoError(t, err)

	// allow time for the server start to settle
	time.Sleep(200 * time.Millisecond)

	return cs
}

func withWait(cancel context.CancelFunc) {
	cancel()
	// give the server time to receive the cancel
	time.Sleep(50 * time.Millisecond)
}
