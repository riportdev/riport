package auditlog

import (
	"net/http/httptest"
	"testing"

	"github.com/riportdev/riport/server/api/users"
	"github.com/riportdev/riport/server/auditlog/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riportdev/riport/db/migration/auditlog"
	"github.com/riportdev/riport/db/sqlite"
	"github.com/riportdev/riport/server/clients/clientdata"
)

var DataSourceOptions = sqlite.DataSourceOptions{WALEnabled: false}

func TestNotEnabled(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	mockProvider := &mockProvider{}
	auditLog, err := New(nil, nil, "", config.Config{Enable: false}, DataSourceOptions)
	require.NoError(t, err)
	auditLog.provider = mockProvider

	// Call with all methods to make sure it doesn't panic if not initialized
	e := auditLog.Entry(ApplicationAuthUser, ActionCreate).
		WithID(123).
		WithHTTPRequest(req).
		WithRequest(map[string]interface{}{}).
		WithResponse(map[string]interface{}{}).
		WithClient(&clientdata.Client{}).
		WithClientID("123")

	e.Save()
	e.SaveForMultipleClients([]*clientdata.Client{&clientdata.Client{}})

	assert.Equal(t, 0, len(mockProvider.entries))
}

func TestIPObfuscation(t *testing.T) {
	testCases := []struct {
		RemoteIP            string
		ExpectedObfuscation string
	}{
		{
			RemoteIP:            "192.0.2.123",
			ExpectedObfuscation: "192.0.2.x",
		}, {
			RemoteIP:            "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			ExpectedObfuscation: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		}, {
			RemoteIP:            "example.com",
			ExpectedObfuscation: "example.com",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.RemoteIP, func(t *testing.T) {
			t.Parallel()

			t.Run("with obfuscation", func(t *testing.T) {
				mockProvider := &mockProvider{}
				auditLog := &AuditLog{
					config: config.Config{
						Enable:           true,
						UseIPObfuscation: true,
					},
					provider: mockProvider,
				}

				e := auditLog.Entry("", "")
				e.RemoteIP = tc.RemoteIP
				e.Save()

				assert.Equal(t, tc.ExpectedObfuscation, mockProvider.entries[0].RemoteIP)
			})

			t.Run("without obfuscation", func(t *testing.T) {
				mockProvider := &mockProvider{}
				auditLog := &AuditLog{
					config: config.Config{
						Enable:           true,
						UseIPObfuscation: false,
					},
					provider: mockProvider,
				}

				e := auditLog.Entry("", "")
				e.RemoteIP = tc.RemoteIP
				e.Save()

				assert.Equal(t, tc.RemoteIP, mockProvider.entries[0].RemoteIP)
			})
		})
	}
}

func TestList(t *testing.T) {
	var tests = []struct {
		desc              string
		user              users.User
		filter            string
		expectedError     string
		expectedResultLen int
	}{
		{
			"Admin success",
			users.User{Username: "Admin", Groups: []string{"Administrators"}},
			"filter[application]=library.script&sort=-timestamp&page[limit]=1&page[offset]=1",
			"",
			1,
		},
		{
			"No Admin denied",
			users.User{Username: "Loser", Groups: []string{"Losers"}},
			"filter[username]=Admin",
			"only members of group Administrators can filter by usernames",
			0,
		},
		{
			"No Admin No results",
			users.User{Username: "Loser", Groups: []string{"Losers"}},
			"",
			"",
			0,
		},
	}
	db, err := sqlite.New(":memory:", auditlog.AssetNames(), auditlog.Asset, DataSourceOptions)
	require.NoError(t, err)
	dbProv := &SQLiteProvider{
		db: db,
	}
	auditLog := &AuditLog{
		config: config.Config{
			Enable: true,
		},
		provider: dbProv,
	}
	defer auditLog.Close()

	auditLog.Entry(ApplicationLibraryScript, ActionCreate).Save()
	auditLog.Entry(ApplicationLibraryScript, ActionUpdate).Save()
	auditLog.Entry(ApplicationLibraryCommand, ActionCreate).Save()

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/auditlog?"+tc.filter, nil)
			result, err := auditLog.List(r, &tc.user)
			if tc.expectedError != "" {
				assert.Error(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
				entries := result.Data.([]*Entry)
				assert.Equal(t, tc.expectedResultLen, len(entries))
				if tc.expectedResultLen > 0 {
					assert.Equal(t, 2, result.Meta.Count)
					assert.Equal(t, ApplicationLibraryScript, entries[0].Application)
					assert.Equal(t, ActionCreate, entries[0].Action)
				}
			}
		})
	}

}
