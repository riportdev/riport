package clientdata

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"

	"github.com/riportdev/riport/server/api/users"
	"github.com/riportdev/riport/server/cgroups"
)

func NewTestClient(id string, address string, hostname string, clientAuthID string, connection ssh.Conn) (c *Client) {
	c = &Client{
		ID:           id,
		Address:      address,
		Hostname:     hostname,
		ClientAuthID: clientAuthID,
		Connection:   connection,
	}
	return c
}

func TestClientBelongsToGroup(t *testing.T) {
	data1 := []byte(`["Linux", "Tag1", "Data*", "Some Tag", "AB*"]`)
	data2 := []byte(`["Some Tag", "AB*"]`)
	dataEmpty := []byte(`[]`)

	c1 := &Client{
		ID:           "test-client-id-1",
		Name:         "Random Rport Client 1",
		OS:           "Linux alpine-3-10-tk-03 4.19.80-0-virt #1-Alpine SMP Fri Oct 18 11:51:24 UTC 2019 x86_64 Linux",
		OSArch:       "amd64",
		OSFamily:     "alpine",
		OSKernel:     "linux",
		Hostname:     "alpine-1-10-tk-01",
		IPv4:         []string{"192.168.122.113", "192.168.122.114"},
		IPv6:         []string{"fe80::b84f:aff:fe59:a0b3"},
		Tags:         []string{"Linux", "Datacenter 3", "TAG1", "Tag2", "Tag3"},
		Version:      "0.1.12",
		Address:      "88.198.189.163:50078",
		ClientAuthID: "client-auth-1",
	}
	c2 := &Client{
		ID:      "test-client-id-1",
		OS:      "Linux alpine-3-10-tk-03 4.19.80-0-virt #1-Alpine SMP Fri Oct 18 11:51:24 UTC 2019 x86_64 Linux",
		IPv4:    []string{"192.168.122.113", "192.168.122.114"},
		Version: "0.1.12",
	}
	g1 := &cgroups.ClientGroup{
		ID: "group-1",
		Params: &cgroups.ClientParams{
			ClientID:        &cgroups.ParamValues{"test-client-id-1", "test-client-id-2"},
			Name:            &cgroups.ParamValues{"random riport client*", "My Client*"},
			OS:              &cgroups.ParamValues{"linux*"},
			OSArch:          &cgroups.ParamValues{"amd64", "darwin", "windows"},
			OSFamily:        &cgroups.ParamValues{"alpine", "win*"},
			OSKernel:        &cgroups.ParamValues{"LINUX", "solaris"},
			Hostname:        &cgroups.ParamValues{"a*", "l*", "w*"},
			IPv4:            &cgroups.ParamValues{"192.168.122.121", "192.168.122.11*"},
			IPv6:            &cgroups.ParamValues{"fe80::b84f:aff:fe59:a0b3"},
			Tag:             (*json.RawMessage)(&data1),
			Version:         &cgroups.ParamValues{"0.1.1*"},
			Address:         &cgroups.ParamValues{"88.198.189.163*"},
			ClientAuthID:    &cgroups.ParamValues{"client-auth-1", "client-auth-2", "client-auth-3*"},
			ConnectionState: &cgroups.ParamValues{"connected"},
		},
	}
	g2 := &cgroups.ClientGroup{
		ID: "group-1",
		Params: &cgroups.ClientParams{
			ClientID:     &cgroups.ParamValues{"test-client-id-1", "test-client-id-2"},
			OS:           &cgroups.ParamValues{"Linux*"},
			IPv4:         &cgroups.ParamValues{"192.168.122.121", "192.168.122.11*"},
			Version:      &cgroups.ParamValues{"0.1.1*"},
			ClientAuthID: &cgroups.ParamValues{"client-auth-1", "client-auth-2", "client-auth-3*"},
		},
	}
	g3 := &cgroups.ClientGroup{
		ID: "group-1",
		Params: &cgroups.ClientParams{
			ClientID: &cgroups.ParamValues{"test-client-id-1", "test-client-id-2"},
			OS:       &cgroups.ParamValues{"Linux*"},
			Version:  &cgroups.ParamValues{"0.1.1*"},
		},
	}

	testCases := []struct {
		name string

		client *Client
		group  *cgroups.ClientGroup

		wantRes bool
	}{
		{
			name:    "all group param, all client params",
			client:  c1,
			group:   g1,
			wantRes: true,
		},

		{
			name:    "all group params, not all client params",
			client:  c2,
			group:   g1,
			wantRes: false,
		},

		{
			name:    "not all group params, not all client params, extra group param",
			client:  c2,
			group:   g2,
			wantRes: false,
		},

		{
			name:    "not all group params, not all client params, extra client param",
			client:  c2,
			group:   g3,
			wantRes: true,
		},

		{
			name:    "not all group params, all client params",
			client:  c1,
			group:   g2,
			wantRes: true,
		},

		{
			name:   "one param does not match",
			client: c1,
			group: &cgroups.ClientGroup{
				ID: "group-1",
				Params: &cgroups.ClientParams{
					ClientID: &cgroups.ParamValues{"test-client-id-1", "test-client-id-2"},
					Name:     &cgroups.ParamValues{"Random Rport Client*", "My Client*"},
					OS:       &cgroups.ParamValues{"Linux*"},
					Tag:      (*json.RawMessage)(&data2),
				},
			},
			wantRes: false,
		},

		{
			name: "no group params, one client param",
			client: &Client{
				ID: "test-client-id-1",
			},
			group: &cgroups.ClientGroup{
				ID:          "empty-group",
				Description: "Group with no params",
				Params:      &cgroups.ClientParams{},
			},
			wantRes: false,
		},

		{
			name: "group with no tags, client with empty tag",
			client: &Client{
				ID:   "test-client-id-1",
				Tags: []string{""},
			},
			group: &cgroups.ClientGroup{
				ID: "no tags",
				Params: &cgroups.ClientParams{
					ClientID: &cgroups.ParamValues{"*"},
					Tag:      (*json.RawMessage)(&dataEmpty),
				},
			},
			wantRes: false,
		},

		{
			name: "group with no tags, client with nonempty tag",
			client: &Client{
				ID:   "test-client-id-1",
				Tags: []string{"tag1"},
			},
			group: &cgroups.ClientGroup{
				ID: "no tags",
				Params: &cgroups.ClientParams{
					ClientID: &cgroups.ParamValues{"*"},
					Tag:      (*json.RawMessage)(&dataEmpty),
				},
			},
			wantRes: false,
		},

		{
			name: "group with unset tags, client with tags",
			client: &Client{
				ID:   "test-client-id-1",
				Tags: []string{"tag1"},
			},
			group: &cgroups.ClientGroup{
				ID: "no tags",
				Params: &cgroups.ClientParams{
					ClientID: &cgroups.ParamValues{"*"},
					Tag:      nil,
				},
			},
			wantRes: true,
		},

		{
			name: "group with unset tags, client with empty tag",
			client: &Client{
				ID:   "test-client-id-1",
				Tags: []string{""},
			},
			group: &cgroups.ClientGroup{
				ID: "no tags",
				Params: &cgroups.ClientParams{
					ClientID: &cgroups.ParamValues{"*"},
					Tag:      nil,
				},
			},
			wantRes: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			gotRes := tc.client.BelongsTo(tc.group)

			// then
			assert.Equal(t, tc.wantRes, gotRes)
		})
	}
}

func TestClientBelongsToGroupLogicalOps(t *testing.T) {
	c1 := &Client{
		ID:           "test-client-id-1",
		Name:         "Random Rport Client 1",
		OS:           "Linux alpine-3-10-tk-03 4.19.80-0-virt #1-Alpine SMP Fri Oct 18 11:51:24 UTC 2019 x86_64 Linux",
		OSArch:       "amd64",
		OSFamily:     "alpine",
		OSKernel:     "linux",
		Hostname:     "alpine-1-10-tk-01",
		IPv4:         []string{"192.168.122.113", "192.168.122.114"},
		IPv6:         []string{"fe80::b84f:aff:fe59:a0b3"},
		Tags:         []string{"Linux", "Datacenter 3", "TAG1", "Tag2", "Tag3"},
		Version:      "0.1.12",
		Address:      "88.198.189.163:50078",
		ClientAuthID: "client-auth-1",
	}

	testCases := []struct {
		name     string
		client   *Client
		jsonData string
		wantRes  bool
	}{
		{
			name:     "match with or no logical operator",
			client:   c1,
			jsonData: ` { "tag": ["Linux", "Datacenter 3"] } `,
			wantRes:  true,
		},

		{
			name:     "match with and 1",
			client:   c1,
			jsonData: `{ "tag": { "and": ["Linux", "Datacenter 2"] } }`,
			wantRes:  false,
		},

		{
			name:     "match with and 2",
			client:   c1,
			jsonData: ` { "tag": { "and": ["Linux", "Datacenter 3"] } }`,
			wantRes:  true,
		},

		{
			name:     "match with or",
			client:   c1,
			jsonData: ` {"tag": { "or": ["Linux", "Datacenter 2"] } }`,
			wantRes:  true,
		},

		{
			name:     "match with and with wildcards",
			client:   c1,
			jsonData: `{ "tag": { "and": ["T*", "Datacenter 2", "Datacenter 5"] } }`,
			wantRes:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var group cgroups.ClientGroup
			var basicTestCase cgroups.ClientParams
			err := json.Unmarshal([]byte(tc.jsonData), &basicTestCase)
			assert.NoError(t, err)
			group.ID = "group1"
			group.Params = &basicTestCase
			// when
			gotRes := tc.client.BelongsTo(&group)

			// then
			assert.Equal(t, tc.wantRes, gotRes)
		})
	}
}
func TestHasAccess(t *testing.T) {
	testCases := []struct {
		name string

		client     *Client
		userGroups []string

		wantRes bool
	}{
		{
			name: "empty acl, empty user groups",
			client: &Client{
				AllowedUserGroups: nil,
			},
			userGroups: nil,
			wantRes:    false,
		},
		{
			name: "non-empty acl, empty user groups",
			client: &Client{
				AllowedUserGroups: []string{"group1"},
			},
			userGroups: nil,
			wantRes:    false,
		},
		{
			name: "empty acl, non-empty user groups",
			client: &Client{
				AllowedUserGroups: nil,
			},
			userGroups: []string{"group1"},
			wantRes:    false,
		},
		{
			name: "acl with no explicit admin, user is admin",
			client: &Client{
				AllowedUserGroups: []string{"group1"},
			},
			userGroups: []string{users.Administrators},
			wantRes:    true,
		},
		{
			name: "empty acl, user is admin",
			client: &Client{
				AllowedUserGroups: nil,
			},
			userGroups: []string{users.Administrators},
			wantRes:    true,
		},
		{
			name: "acl contains user group",
			client: &Client{
				AllowedUserGroups: []string{"group2"},
			},
			userGroups: []string{"group1", "group2", "group3"},
			wantRes:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			gotRes := tc.client.HasAccessViaUserGroups(tc.userGroups)

			// then
			assert.Equal(t, tc.wantRes, gotRes)
		})
	}
}

func TestToCalculatedForGroups(t *testing.T) {
	data4 := []byte(`["AB*"]`)
	data5 := []byte(`"[Other]"`)
	client := &Client{
		Name: "abc",
		Tags: []string{"ABC"},
	}
	groups := []*cgroups.ClientGroup{
		{
			ID: "group1",
			Params: &cgroups.ClientParams{
				Name: &cgroups.ParamValues{"abc"},
			},
		},
		{
			ID: "group2",
			Params: &cgroups.ClientParams{
				Tag: (*json.RawMessage)(&data4),
			},
		},
		{
			ID: "group3",
			Params: &cgroups.ClientParams{
				Tag: (*json.RawMessage)(&data5),
			},
		},
	}

	calculated := client.ToCalculated(groups)
	assert.Equal(t, client, calculated.Client)
	assert.Equal(t, []string{"group1", "group2"}, calculated.Groups)
}

func TestToCalculatedWhenConnected(t *testing.T) {
	client := &Client{
		Name:           "abc",
		DisconnectedAt: nil,
	}
	groups := []*cgroups.ClientGroup{}

	calculated := client.ToCalculated(groups)
	assert.Equal(t, client, calculated.Client)
	assert.Equal(t, "connected", string(calculated.ConnectionState))
}

func TestToCalculatedWhenDisconnected(t *testing.T) {
	now := time.Now()
	client := &Client{
		Name:           "abc",
		DisconnectedAt: &now,
	}
	groups := []*cgroups.ClientGroup{}

	calculated := client.ToCalculated(groups)
	assert.Equal(t, client, calculated.Client)
	assert.Equal(t, "disconnected", string(calculated.ConnectionState))
}
