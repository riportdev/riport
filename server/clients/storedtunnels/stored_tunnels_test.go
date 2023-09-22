package storedtunnels

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riportdev/riport/db/migration/clients"
	"github.com/riportdev/riport/db/sqlite"
	"github.com/riportdev/riport/share/query"
	"github.com/riportdev/riport/share/random"
)

var DataSourceOptions = sqlite.DataSourceOptions{WALEnabled: false}

func TestStoredTunnels(t *testing.T) {
	ctx := context.Background()
	client1, err := random.UUID4()
	require.NoError(t, err)
	client2, err := random.UUID4()
	require.NoError(t, err)
	db, err := sqlite.New(":memory:", clients.AssetNames(), clients.Asset, DataSourceOptions)
	require.NoError(t, err)
	tunnel := &StoredTunnel{}
	options := &query.ListOptions{}

	manager := New(db)

	// no results initially
	results, err := manager.List(ctx, options, client1)
	require.NoError(t, err)
	assert.Equal(t, 0, results.Meta.Count)

	_, err = manager.Create(ctx, client1, tunnel)
	require.NoError(t, err)

	// client1 has one stored tunnel
	results, err = manager.List(ctx, options, client1)
	require.NoError(t, err)
	assert.Equal(t, 1, results.Meta.Count)

	// client2 has no stored tunnels
	results, err = manager.List(ctx, options, client2)
	require.NoError(t, err)
	assert.Equal(t, 0, results.Meta.Count)

	err = manager.Delete(ctx, client1, tunnel.ID)
	require.NoError(t, err)

	// client1 doesn't have any stored tunnel anymore
	results, err = manager.List(ctx, options, client1)
	require.NoError(t, err)
	assert.Equal(t, 0, results.Meta.Count)
}
