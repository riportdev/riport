package chserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riportdev/riport/db/migration/api_token"
	"github.com/riportdev/riport/db/sqlite"
	"github.com/riportdev/riport/server/api"
	"github.com/riportdev/riport/server/api/authorization"
	"github.com/riportdev/riport/server/api/users"
	"github.com/riportdev/riport/server/bearer"
	"github.com/riportdev/riport/server/chconfig"
	"github.com/riportdev/riport/share/ptr"
	"github.com/riportdev/riport/share/random"
	"github.com/riportdev/riport/share/security"
)

func TestAPITokenOps(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	user := &users.User{
		Username: "test-user",
	}
	mockUsersService := &MockUsersService{
		UserService: users.NewAPIService(users.NewStaticProvider([]*users.User{user}), false, 0, -1),
	}

	// database
	apiTokenDb, err := sqlite.New(":memory:", api_token.AssetNames(), api_token.Asset, DataSourceOptions)
	require.NoError(err)
	defer apiTokenDb.Close()
	tokenProvider := authorization.NewSqliteProvider(apiTokenDb)
	mockTokenManager := authorization.NewManager(tokenProvider)

	uuid := "mynicefi-xedl-enth-long-livedpasswor"
	oldUUID := random.UUID4
	random.UUID4 = func() (string, error) {
		return uuid, nil
	}
	defer func() {
		random.UUID4 = oldUUID
	}()

	MyalphaNumNewPrefix := "theprefi"
	oldAlphaNum := random.AlphaNum
	random.AlphaNum = func(int) string {
		return MyalphaNumNewPrefix
	}
	defer func() {
		random.AlphaNum = oldAlphaNum
	}()

	expirationDate, _ := time.Date(2025, 1, 1, 2, 0, 0, 0, time.UTC).UTC().MarshalText()
	updateExpirationDate, _ := time.Date(2026, 3, 10, 5, 0, 0, 0, time.UTC).UTC().MarshalText()

	testCases := []struct {
		descr string // Test Case Description

		clientAuthWrite bool
		requestMethod   string
		requestURL      string
		requestBody     io.Reader

		wantStatusCode int
		wantJSON       string
		wantErrCode    string
		wantErrTitle   string
		wantErrDetail  string
	}{
		{
			descr:          "create token bad json in request body",
			requestMethod:  http.MethodPost,
			requestURL:     "/api/v1/me/tokens",
			requestBody:    strings.NewReader(`}`),
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Invalid JSON data",
		},
		{
			descr:          "create token empty request body",
			requestMethod:  http.MethodPost,
			requestURL:     "/api/v1/me/tokens",
			requestBody:    nil,
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Invalid JSON data",
			wantErrDetail:  "Missing body with json data.",
			wantErrCode:    "",
		},
		{
			descr:          "new token bad scope creation",
			requestMethod:  http.MethodPost,
			requestURL:     "/api/v1/me/tokens",
			requestBody:    strings.NewReader(`{"scope": "reads"}`),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    "",
			wantErrTitle:   "missing or invalid scope.",
		},
		{
			descr:          "new token no scope provided",
			requestMethod:  http.MethodPost,
			requestURL:     "/api/v1/me/tokens",
			requestBody:    strings.NewReader(`{"name": "This is a test name"}`),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    "",
			wantErrTitle:   "missing or invalid scope.",
		},
		{
			descr:          "new token no name provided",
			requestMethod:  http.MethodPost,
			requestURL:     "/api/v1/me/tokens",
			requestBody:    strings.NewReader(`{"scope": "` + string(authorization.APITokenRead) + `"}`),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    "",
			wantErrTitle:   "missing or invalid name.",
		},
		{
			descr:          "delete a token, prefix wrong len",
			requestMethod:  http.MethodDelete,
			requestURL:     "/api/v1/me/tokens/hjk",
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    "",
			wantErrTitle:   "missing or invalid token prefix.",
		},
		{
			descr:          "new token read creation, too long of a name",
			requestMethod:  http.MethodPost,
			requestURL:     "/api/v1/me/tokens",
			requestBody:    strings.NewReader(`{"scope": "` + string(authorization.APITokenRead) + `", "name":"` + strings.Repeat("I am a token and this is my name", 10) + `"}`),
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "missing or invalid name.",
			wantErrDetail:  "field name is required, 250 characters max",
		},
		{
			descr:          "new token read creation",
			requestMethod:  http.MethodPost,
			requestURL:     "/api/v1/me/tokens",
			requestBody:    strings.NewReader(`{"scope": "` + string(authorization.APITokenRead) + `", "expires_at": "` + string(expirationDate) + `",  "name": "This is my name"}`),
			wantStatusCode: http.StatusOK,
			wantJSON:       `{"data":{"expires_at":"2025-01-01T02:00:00Z", "scope":"` + string(authorization.APITokenRead) + `", "token":"theprefi_mynicefi-xedl-enth-long-livedpasswor", "prefix": "theprefi"}}`,
		},
		{
			descr:          "new token same name",
			requestMethod:  http.MethodPost,
			requestURL:     "/api/v1/me/tokens",
			requestBody:    strings.NewReader(`{"scope": "` + string(authorization.APITokenRead) + `", "name": "This is my name"}`),
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "A token with the same name already exists",
		},
		{
			descr:          "new token read+write creation",
			requestMethod:  http.MethodPost,
			requestURL:     "/api/v1/me/tokens",
			requestBody:    strings.NewReader(`{"scope": "` + string(authorization.APITokenReadWrite) + `", "name": "This is my name 2", "expires_at": "` + string(expirationDate) + `"}`),
			wantStatusCode: http.StatusOK,
			wantJSON:       `{"data":{"expires_at":"2025-01-01T02:00:00Z", "scope":"` + string(authorization.APITokenReadWrite) + `", "token":"theprefi_mynicefi-xedl-enth-long-livedpasswor", "prefix":"theprefi"}}`,
		},
		{
			descr:          "update token name collision",
			requestMethod:  http.MethodPut,
			requestURL:     "/api/v1/me/tokens/theprefi",
			requestBody:    strings.NewReader(`{"name": "This is my name 2"}`),
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "A token with the same name already exists",
		},
		{
			descr:          "token update with expires_at",
			requestMethod:  http.MethodPut,
			requestURL:     "/api/v1/me/tokens/theprefi",
			requestBody:    strings.NewReader(`{"expires_at": "` + string(updateExpirationDate) + `"}`),
			wantStatusCode: http.StatusOK,
			wantJSON:       `{"data":{"expires_at":"2026-03-10T05:00:00Z", "prefix":"theprefi" }}`,
		},
		{
			descr:          "token update with name",
			requestMethod:  http.MethodPut,
			requestURL:     "/api/v1/me/tokens/theprefi",
			requestBody:    strings.NewReader(`{"name": "new name"}`),
			wantStatusCode: http.StatusOK,
			wantJSON:       `{"data":{"name": "new name", "prefix":"theprefi" }}`,
		},
		{
			descr:          "delete a token ",
			requestMethod:  http.MethodDelete,
			requestURL:     "/api/v1/me/tokens/" + MyalphaNumNewPrefix,
			wantStatusCode: http.StatusNoContent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.descr, func(t *testing.T) {
			// given
			al := APIListener{
				Logger:           testLog,
				insecureForTests: true,
				Server: &Server{
					config: &chconfig.Config{
						API: chconfig.APIConfig{
							MaxRequestBytes: 1024 * 1024,
						},
					},
				},
				tokenManager: mockTokenManager,
				userService:  mockUsersService,
			}

			al.initRouter()
			req := httptest.NewRequest(tc.requestMethod, tc.requestURL, tc.requestBody)

			// when
			w := httptest.NewRecorder()
			ctx := api.WithUser(req.Context(), user.Username)
			req = req.WithContext(ctx)
			al.router.ServeHTTP(w, req)
			t.Logf("Got response %s", w.Body)

			// then
			require.Equal(tc.wantStatusCode, w.Code)
			if tc.wantStatusCode/100 == 2 { // any 2** status code is a success test case
				// success case
				if tc.wantJSON == "" {
					assert.Empty(w.Body.String())
				} else {
					assert.JSONEq(tc.wantJSON, w.Body.String())

				}
			} else {
				// failure case

				// decode response
				tResponse := api.NewErrAPIPayloadFromMessage("", "", "")
				err = json.Unmarshal(w.Body.Bytes(), &tResponse)
				require.NoError(err)

				// compare items of struct error, one by one,
				if tc.wantErrTitle != "" {
					assert.Equal(tc.wantErrTitle, tResponse.Errors[0].Title)
				}
				if tc.wantErrDetail != "" {
					assert.Equal(tc.wantErrDetail, tResponse.Errors[0].Detail)
				}
			}
		})
	}
}

func CommonAPITokenTestDb(t *testing.T, username, prefix, name string, scope authorization.APITokenScope, token string) *authorization.SqliteProvider {
	db, err := sqlite.New(":memory:", api_token.AssetNames(), api_token.Asset, DataSourceOptions)
	require.NoError(t, err)
	dbProv := authorization.NewSqliteProvider(db)

	ctx := context.Background()
	itemToSave := authorization.APIToken{
		Username:  username,
		Prefix:    prefix,
		Name:      name,
		CreatedAt: ptr.Time(time.Date(2001, 1, 1, 1, 0, 0, 0, time.UTC)),
		ExpiresAt: ptr.Time(time.Date(2051, 1, 1, 2, 0, 0, 0, time.UTC)),
		Scope:     scope,
		Token:     token,
	}
	err = dbProv.Save(ctx, &itemToSave)
	require.NoError(t, err)

	itemToSave = authorization.APIToken{
		Username:  username,
		Prefix:    "expired1",
		Name:      "another name",
		CreatedAt: ptr.Time(time.Date(2001, 1, 1, 1, 0, 0, 0, time.UTC)),
		ExpiresAt: ptr.Time(time.Date(2001, 1, 1, 2, 0, 0, 0, time.UTC)),
		Scope:     scope,
		Token:     token,
	}
	err = dbProv.Save(ctx, &itemToSave)
	require.NoError(t, err)

	return dbProv
}

type MockUsersService struct {
	UserService

	ChangeUser     *users.User
	ChangeUsername string
}

func (s *MockUsersService) Change(user *users.User, username string) error {
	s.ChangeUser = user
	s.ChangeUsername = username
	return nil
}

func TestPostToken(t *testing.T) {
	user := &users.User{
		Username: "user1",
		Password: "$2y$05$ep2DdPDeLDDhwRrED9q/vuVEzRpZtB5WHCFT7YbcmH9r9oNmlsZOm",
	}
	userWithoutToken := &users.User{
		Username: "user2",
		Password: "$2y$05$ep2DdPDeLDDhwRrED9q/vuVEzRpZtB5WHCFT7YbcmH9r9oNmlsZOm",
	}

	// database for tokenManager, creates a token read+write
	apiTokenDb, err := sqlite.New(":memory:", api_token.AssetNames(), api_token.Asset, DataSourceOptions)
	require.NoError(t, err)
	defer apiTokenDb.Close()
	tokenProvider := authorization.NewSqliteProvider(apiTokenDb)
	mockTokenManager := authorization.NewManager(tokenProvider)

	uuid := "mynicefi-xedl-enth-long-livedpasswor"
	oldUUID := random.UUID4
	random.UUID4 = func() (string, error) {
		return uuid, nil
	}
	defer func() {
		random.UUID4 = oldUUID
	}()

	MyalphaNumNewPrefix := "theprefi"
	oldAlphaNum := random.AlphaNum
	random.AlphaNum = func(int) string {
		return MyalphaNumNewPrefix
	}
	defer func() {
		random.AlphaNum = oldAlphaNum
	}()

	al := APIListener{
		Logger:      testLog,
		bannedUsers: security.NewBanList(0),
		apiSessions: newEmptyAPISessionCache(t),
		Server: &Server{
			config: &chconfig.Config{
				API: chconfig.APIConfig{
					MaxRequestBytes: 1024 * 1024,
				},
			},
		},
		tokenManager: mockTokenManager,
		userService:  users.NewAPIService(users.NewStaticProvider([]*users.User{user, userWithoutToken}), false, 0, -1),
	}
	expirationDate, _ := time.Date(2025, 1, 1, 2, 0, 0, 0, time.UTC).UTC().MarshalText()

	al.initRouter()
	req := httptest.NewRequest("POST", "/api/v1/me/tokens", strings.NewReader(`{"name": "token name", "scope": "`+string(authorization.APITokenReadWrite)+`", "expires_at": "`+string(expirationDate)+`"}`))
	w := httptest.NewRecorder()
	ctxUser1 := api.WithUser(req.Context(), user.Username)
	req = req.WithContext(ctxUser1)
	req.SetBasicAuth(user.Username, "pwd")
	al.router.ServeHTTP(w, req)
	expectedJSON := `{"data":{"scope":"` + string(authorization.APITokenReadWrite) + `", "token":"theprefi_mynicefi-xedl-enth-long-livedpasswor", "prefix":"theprefi", "expires_at": "` + string(expirationDate) + `"}}`
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, expectedJSON, w.Body.String())
}
func TestWrapWithAuthMiddleware(t *testing.T) {
	ctx := context.Background()

	user := &users.User{
		Username: "user1",
		Password: "$2y$05$ep2DdPDeLDDhwRrED9q/vuVEzRpZtB5WHCFT7YbcmH9r9oNmlsZOm",
	}
	userWithoutToken := &users.User{
		Username: "user2",
		Password: "$2y$05$ep2DdPDeLDDhwRrED9q/vuVEzRpZtB5WHCFT7YbcmH9r9oNmlsZOm",
	}
	mockTokenManager := authorization.NewManager(
		CommonAPITokenTestDb(t, "user1", "theprefi", "the name", authorization.APITokenReadWrite, "mynicefi-xedl-enth-long-livedpasswor")) // APIToken database

	al := APIListener{
		Logger:      testLog,
		apiSessions: newEmptyAPISessionCache(t),
		bannedUsers: security.NewBanList(0),
		Server: &Server{
			config: &chconfig.Config{
				API: chconfig.APIConfig{
					MaxRequestBytes: 1024 * 1024,
				},
			},
		},
		tokenManager: mockTokenManager,
		userService:  users.NewAPIService(users.NewStaticProvider([]*users.User{user, userWithoutToken}), false, 0, -1),
	}

	al.initRouter()
	jwt, err := bearer.CreateAuthToken(ctx, al.apiSessions, al.config.API.JWTSecret, time.Hour, user.Username, []bearer.Scope{}, "", "")
	require.NoError(t, err)

	testCases := []struct {
		Name           string
		Username       string
		Password       string
		EnableTwoFA    bool
		Bearer         string
		ExpectedStatus int
	}{
		{
			Name:           "no auth",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "basic auth with password",
			Username:       user.Username,
			Password:       "pwd",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "basic auth with password, no password",
			Username:       user.Username,
			Password:       "",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "basic auth with password, wrong password",
			Username:       user.Username,
			Password:       "wrong",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "basic auth with password, 2fa enabled",
			Username:       user.Username,
			Password:       "pwd",
			EnableTwoFA:    true,
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "basic auth with token",
			Username:       user.Username,
			Password:       "theprefi_mynicefi-xedl-enth-long-livedpasswor",
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "basic auth with an expired token",
			Username:       user.Username,
			Password:       "expired1_mynicefi-xedl-enth-long-livedpasswor",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "basic auth with token, 2fa enabled",
			Username:       user.Username,
			Password:       "theprefi_mynicefi-xedl-enth-long-livedpasswor",
			EnableTwoFA:    true,
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "basic auth with token, wrong token",
			Username:       user.Username,
			Password:       "wrong-token",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "basic auth with token, user has no token",
			Username:       userWithoutToken.Username,
			Password:       "",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:           "bearer token",
			ExpectedStatus: http.StatusOK,
			Bearer:         jwt,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			twoFATokenDelivery := ""
			if tc.EnableTwoFA {
				twoFATokenDelivery = "smtp"
			}
			al.config.API.TwoFATokenDelivery = twoFATokenDelivery

			handler := al.wrapWithAuthMiddleware(false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, user.Username, api.GetUser(r.Context(), nil))
			}))

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/some-endpoint", nil)
			if tc.Username != "" {
				req.SetBasicAuth(tc.Username, tc.Password)
			}

			if tc.Bearer != "" {
				req.Header.Set("Authorization", "Bearer "+tc.Bearer)
			}

			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.ExpectedStatus, w.Code)
		})
	}
}

func TestAPISessionUpdates(t *testing.T) {
	ctx := context.Background()

	user := &users.User{
		Username: "user1",
		Password: "$2y$05$ep2DdPDeLDDhwRrED9q/vuVEzRpZtB5WHCFT7YbcmH9r9oNmlsZOm",
	}
	userWithoutToken := &users.User{
		Username: "user2",
		Password: "$2y$05$ep2DdPDeLDDhwRrED9q/vuVEzRpZtB5WHCFT7YbcmH9r9oNmlsZOm",
	}
	mockTokenManager := authorization.NewManager(
		CommonAPITokenTestDb(t, "user1", "theprefi", "the name", authorization.APITokenReadWrite, "mynicefi-xedl-enth-long-livedpasswor")) // APIToken database

	al := APIListener{
		Logger:      testLog,
		apiSessions: newEmptyAPISessionCache(t),
		bannedUsers: security.NewBanList(0),
		Server: &Server{
			config: &chconfig.Config{
				API: chconfig.APIConfig{
					MaxRequestBytes: 1024 * 1024,
				},
			},
		},
		tokenManager: mockTokenManager,
		userService:  users.NewAPIService(users.NewStaticProvider([]*users.User{user, userWithoutToken}), false, 0, -1),
	}
	al.initRouter()

	testIPAddress := "1.2.3.4"
	testUserAgent := "Chrome"

	jwt, err := bearer.CreateAuthToken(ctx, al.apiSessions, al.config.API.JWTSecret, time.Hour, user.Username, []bearer.Scope{}, testUserAgent, testIPAddress)
	require.NoError(t, err)

	testCases := []struct {
		Name            string
		Username        string
		Password        string
		EnableTwoFA     bool
		BearerUsername  string
		Bearer          string
		ExpectedSession bool
		ExpectedStatus  int
	}{
		{
			Name:           "no session",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:            "user auth, regular login, existing token",
			Username:        user.Username,
			Password:        "pwd",
			ExpectedSession: true,
			ExpectedStatus:  http.StatusOK,
		},
		{
			Name:            "user auth, regular login, no token",
			Username:        userWithoutToken.Username,
			Password:        "pwd",
			ExpectedSession: false,
			ExpectedStatus:  http.StatusOK,
		},
		{
			Name:            "user auth, existing token, 2fa, bad password",
			Username:        user.Username,
			Password:        "pwd",
			EnableTwoFA:     true,
			ExpectedSession: true,
			ExpectedStatus:  http.StatusUnauthorized,
		},
		{
			Name:            "user auth, existing token, 2fa, good password",
			Username:        user.Username,
			Password:        "theprefi_mynicefi-xedl-enth-long-livedpasswor",
			EnableTwoFA:     true,
			ExpectedSession: true,
			ExpectedStatus:  http.StatusOK,
		},
		{
			Name:            "existing bearer token only",
			BearerUsername:  user.Username,
			Bearer:          jwt,
			ExpectedSession: true,
			ExpectedStatus:  http.StatusOK,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			twoFATokenDelivery := ""
			if tc.EnableTwoFA {
				twoFATokenDelivery = "smtp"
			}
			al.config.API.TwoFATokenDelivery = twoFATokenDelivery

			testRunTime := time.Now()

			handler := al.wrapWithAuthMiddleware(tc.Bearer != "")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				username := tc.Username
				if tc.BearerUsername != "" {
					username = tc.BearerUsername
				}
				sessions, err := al.apiSessions.GetAllByUser(ctx, username)
				require.NoError(t, err)

				if tc.ExpectedSession {
					require.NotNil(t, sessions)
					require.Greater(t, len(sessions), 0)

					if sessions != nil {
						session := sessions[0]
						assert.Equal(t, username, session.Username)
						assert.Greater(t, time.Now(), session.LastAccessAt)
						assert.Equal(t, testIPAddress, session.IPAddress)
						assert.Equal(t, testUserAgent, session.UserAgent)
						// if ok access and we used a bearer token
						if tc.ExpectedStatus == http.StatusOK && tc.Bearer != "" {
							// then test run time should be before last access time
							assert.Less(t, testRunTime, session.LastAccessAt)
						}
					}

				}
			}))

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/some-endpoint", nil)
			req.RemoteAddr = testIPAddress
			req.Header.Set("User-Agent", testUserAgent)

			if tc.Username != "" {
				req.SetBasicAuth(tc.Username, tc.Password)
			}
			if tc.Bearer != "" {
				req.Header.Set("Authorization", "Bearer "+tc.Bearer)
			}

			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.ExpectedStatus, w.Code)
		})
	}
}

func TestHandleGetLogin(t *testing.T) {
	authHeader := "Authentication-IsAuthenticated"
	userHeader := "Authentication-User"
	userGroup := "Administrators"
	user := &users.User{
		Username: "user1",
		Password: "$2y$05$ep2DdPDeLDDhwRrED9q/vuVEzRpZtB5WHCFT7YbcmH9r9oNmlsZOm",
	}
	mockUsersService := &MockUsersService{
		UserService: users.NewAPIService(users.NewStaticProvider([]*users.User{user}), false, 0, -1),
	}
	mockTokenManager := authorization.NewManager(
		CommonAPITokenTestDb(t, "user1", "theprefi", "the name", authorization.APITokenReadWrite, "mynicefi-xedl-enth-long-livedpasswor")) // APIToken database

	al := APIListener{
		Logger: testLog,
		Server: &Server{
			config: &chconfig.Config{
				API: chconfig.APIConfig{
					DefaultUserGroup: userGroup,
					MaxRequestBytes:  1024 * 1024,
				},
			},
		},
		bannedUsers:  security.NewBanList(0),
		userService:  mockUsersService,
		apiSessions:  newEmptyAPISessionCache(t),
		tokenManager: mockTokenManager,
	}
	al.initRouter()

	testCases := []struct {
		Name              string
		BasicAuthPassword string
		HeaderAuthUser    string
		HeaderAuthEnabled bool
		CreateMissingUser bool
		ExpectedStatus    int
	}{
		{
			Name:           "no auth",
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:              "basic auth",
			BasicAuthPassword: "pwd",
			ExpectedStatus:    http.StatusOK,
		},
		{
			Name:           "header auth - disabled",
			HeaderAuthUser: user.Username,
			ExpectedStatus: http.StatusUnauthorized,
		},
		{
			Name:              "header auth - enabled",
			HeaderAuthUser:    user.Username,
			HeaderAuthEnabled: true,
			ExpectedStatus:    http.StatusOK,
		},
		{
			Name:              "header auth + invalid basic auth",
			HeaderAuthUser:    user.Username,
			HeaderAuthEnabled: true,
			BasicAuthPassword: "invalid",
			ExpectedStatus:    http.StatusOK,
		},
		{
			Name:              "header auth - unknown user",
			HeaderAuthUser:    "unknown",
			HeaderAuthEnabled: true,
			CreateMissingUser: true,
			ExpectedStatus:    http.StatusOK,
		}, {
			Name:              "header auth - create missing user",
			HeaderAuthUser:    "new-user",
			HeaderAuthEnabled: true,
			CreateMissingUser: true,
			ExpectedStatus:    http.StatusOK,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			if tc.HeaderAuthEnabled {
				al.config.API.AuthHeader = authHeader
				al.config.API.UserHeader = userHeader
			} else {
				al.config.API.AuthHeader = ""
			}
			al.config.API.CreateMissingUsers = tc.CreateMissingUser

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/v1/login", nil)
			if tc.BasicAuthPassword != "" {
				req.SetBasicAuth(user.Username, tc.BasicAuthPassword)
			}
			if tc.HeaderAuthUser != "" {
				req.Header.Set(authHeader, "1")
				req.Header.Set(userHeader, tc.HeaderAuthUser)
			}

			al.router.ServeHTTP(w, req)

			assert.Equal(t, tc.ExpectedStatus, w.Code)
			if tc.ExpectedStatus == http.StatusOK {
				assert.Contains(t, w.Body.String(), `{"data":{"token":"`)
			}
			if tc.CreateMissingUser {
				assert.Equal(t, tc.HeaderAuthUser, mockUsersService.ChangeUser.Username)
				assert.Equal(t, userGroup, mockUsersService.ChangeUser.Groups[0])
				assert.NotEmpty(t, mockUsersService.ChangeUser.Password)
			}
		})
	}
}
