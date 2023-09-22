package chserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/riportdev/riport/db/migration/client_groups"
	jobsmigration "github.com/riportdev/riport/db/migration/jobs"
	"github.com/riportdev/riport/db/sqlite"
	"github.com/riportdev/riport/server/api"
	"github.com/riportdev/riport/server/api/authorization"
	"github.com/riportdev/riport/server/api/jobs"
	"github.com/riportdev/riport/server/api/jobs/schedule"
	"github.com/riportdev/riport/server/api/users"
	"github.com/riportdev/riport/server/cgroups"
	"github.com/riportdev/riport/server/chconfig"
	"github.com/riportdev/riport/server/clients"
	"github.com/riportdev/riport/server/clients/clientdata"
	"github.com/riportdev/riport/server/test/jb"
	"github.com/riportdev/riport/share/comm"
	"github.com/riportdev/riport/share/logger"
	"github.com/riportdev/riport/share/models"
	"github.com/riportdev/riport/share/query"
	"github.com/riportdev/riport/share/random"
	"github.com/riportdev/riport/share/security"
	"github.com/riportdev/riport/share/test"
	"github.com/riportdev/riport/share/ws"
)

type JobProviderMock struct {
	JobProvider
	ReturnJob     *models.Job
	ReturnJobList []*models.Job
	ReturnErr     error

	InputCID       string
	InputJID       string
	InputSaveJob   *models.Job
	InputCreateJob *models.Job
}

func NewJobProviderMock() *JobProviderMock {
	return &JobProviderMock{}
}

func (p *JobProviderMock) GetByJID(cid, jid string) (*models.Job, error) {
	p.InputCID = cid
	p.InputJID = jid
	return p.ReturnJob, p.ReturnErr
}

func (p *JobProviderMock) List(ctx context.Context, opts *query.ListOptions) ([]*models.Job, error) {
	p.InputCID = opts.Filters[0].Values[0]
	return p.ReturnJobList, p.ReturnErr
}

func (p *JobProviderMock) Count(ctx context.Context, opts *query.ListOptions) (int, error) {
	return len(p.ReturnJobList), p.ReturnErr
}

func (p *JobProviderMock) SaveJob(job *models.Job) error {
	p.InputSaveJob = job
	return p.ReturnErr
}

func (p *JobProviderMock) CreateJob(job *models.Job) error {
	p.InputCreateJob = job
	return p.ReturnErr
}

func (p *JobProviderMock) Close() error {
	return nil
}

func TestHandlePostCommand(t *testing.T) {
	var testJID string
	generateNewJobID = func() (string, error) {
		uuid, err := random.UUID4()
		testJID = uuid
		return uuid, err
	}
	testUser := "test-user"

	defaultTimeout := 60
	gotCmd := "/bin/date;foo;whoami"
	gotCmdTimeoutSec := 30
	validReqBody := `{"command": "` + gotCmd + `","timeout_sec": ` + strconv.Itoa(gotCmdTimeoutSec) + `}`

	connMock := test.NewConnMock()
	// by default set to return success
	connMock.ReturnOk = true
	sshSuccessResp := comm.RunCmdResponse{Pid: 123, StartedAt: time.Date(2020, 10, 10, 10, 10, 10, 0, time.UTC)}
	sshRespBytes, err := json.Marshal(sshSuccessResp)
	require.NoError(t, err)
	connMock.ReturnResponsePayload = sshRespBytes

	c1 := clients.New(t).Connection(connMock).Logger(testLog).Build()
	c2 := clients.New(t).DisconnectedDuration(5 * time.Minute).Logger(testLog).Build()

	testCases := []struct {
		name string

		cid             string
		requestBody     string
		jpReturnSaveErr error
		connReturnErr   error
		connReturnNotOk bool
		connReturnResp  []byte
		runningJob      *models.Job
		clients         []*clientdata.Client

		wantStatusCode  int
		wantTimeout     int
		wantErrCode     string
		wantErrTitle    string
		wantErrDetail   string
		wantInterpreter string
	}{
		{
			name:           "valid cmd",
			requestBody:    validReqBody,
			cid:            c1.GetID(),
			clients:        []*clientdata.Client{c1},
			wantStatusCode: http.StatusOK,
			wantTimeout:    gotCmdTimeoutSec,
		},
		{
			name:            "valid cmd with interpreter",
			requestBody:     `{"command": "` + gotCmd + `","interpreter": "powershell"}`,
			cid:             c1.GetID(),
			clients:         []*clientdata.Client{c1},
			wantStatusCode:  http.StatusOK,
			wantTimeout:     defaultTimeout,
			wantInterpreter: "powershell",
		},
		{
			name:           "invalid interpreter",
			requestBody:    `{"command": "` + gotCmd + `","interpreter": "unsupported"}`,
			cid:            c1.GetID(),
			clients:        []*clientdata.Client{c1},
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Invalid interpreter.",
			wantErrDetail:  "expected interpreter to be one of: [cmd powershell tacoscript], actual: unsupported",
		},
		{
			name:           "valid cmd with no timeout",
			requestBody:    `{"command": "/bin/date;foo;whoami"}`,
			cid:            c1.GetID(),
			clients:        []*clientdata.Client{c1},
			wantTimeout:    defaultTimeout,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "valid cmd with 0 timeout",
			requestBody:    `{"command": "/bin/date;foo;whoami", "timeout_sec": 0}`,
			cid:            c1.GetID(),
			clients:        []*clientdata.Client{c1},
			wantTimeout:    defaultTimeout,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "empty cmd",
			requestBody:    `{"command": "", "timeout_sec": 30}`,
			cid:            c1.GetID(),
			clients:        []*clientdata.Client{c1},
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Command cannot be empty.",
		},
		{
			name:           "no cmd",
			requestBody:    `{"timeout_sec": 30}`,
			cid:            c1.GetID(),
			clients:        []*clientdata.Client{c1},
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Command cannot be empty.",
		},
		{
			name:           "empty body",
			requestBody:    "",
			cid:            c1.GetID(),
			clients:        []*clientdata.Client{c1},
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Missing body with json data.",
		},
		{
			name:           "invalid request body",
			requestBody:    "sdfn fasld fasdf sdlf jd",
			cid:            c1.GetID(),
			clients:        []*clientdata.Client{c1},
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Invalid JSON data.",
			wantErrDetail:  "invalid character 's' looking for beginning of value",
		},
		{
			name:           "invalid request body: unknown param",
			requestBody:    `{"command": "/bin/date;foo;whoami", "timeout": 30}`,
			cid:            c1.GetID(),
			clients:        []*clientdata.Client{c1},
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Invalid JSON data.",
			wantErrDetail:  "json: unknown field \"timeout\"",
		},
		{
			name:           "no active client",
			requestBody:    validReqBody,
			cid:            c1.GetID(),
			clients:        []*clientdata.Client{},
			wantStatusCode: http.StatusNotFound,
			wantErrTitle:   fmt.Sprintf("Active client with id=%q not found.", c1.GetID()),
		},
		{
			name:           "disconnected client",
			requestBody:    validReqBody,
			cid:            c2.GetID(),
			clients:        []*clientdata.Client{c1, c2},
			wantStatusCode: http.StatusNotFound,
			wantErrTitle:   fmt.Sprintf("Active client with id=%q not found.", c2.GetID()),
		},
		{
			name:            "error on save job",
			requestBody:     validReqBody,
			jpReturnSaveErr: errors.New("save fake error"),
			cid:             c1.GetID(),
			clients:         []*clientdata.Client{c1},
			wantStatusCode:  http.StatusInternalServerError,
			wantErrTitle:    "Failed to persist a new job.",
			wantErrDetail:   "save fake error",
		},
		{
			name:           "error on send request",
			requestBody:    validReqBody,
			connReturnErr:  errors.New("send fake error"),
			cid:            c1.GetID(),
			clients:        []*clientdata.Client{c1},
			wantStatusCode: http.StatusInternalServerError,
			wantErrTitle:   "Failed to execute remote command.",
			wantErrDetail:  "failed to send request: send fake error",
		},
		{
			name:           "invalid ssh response format",
			requestBody:    validReqBody,
			connReturnResp: []byte("invalid ssh response data"),
			cid:            c1.GetID(),
			clients:        []*clientdata.Client{c1},
			wantStatusCode: http.StatusConflict,
			wantErrTitle:   "invalid client response format: failed to decode response into *comm.RunCmdResponse: invalid character 'i' looking for beginning of value",
		},
		{
			name:            "failure response on send request",
			requestBody:     validReqBody,
			connReturnNotOk: true,
			connReturnResp:  []byte("fake failure msg"),
			cid:             c1.GetID(),
			clients:         []*clientdata.Client{c1},
			wantStatusCode:  http.StatusConflict,
			wantErrTitle:    "client error: fake failure msg",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			clientService := clients.NewClientService(nil, nil, clients.NewClientRepository(tc.clients, &hour, testLog), testLog, nil)
			al := APIListener{
				insecureForTests: true,
				Server: &Server{
					clientService: clientService,
					config: &chconfig.Config{
						Server: chconfig.ServerConfig{
							RunRemoteCmdTimeoutSec: defaultTimeout,
						},
						API: chconfig.APIConfig{
							MaxRequestBytes: 1024 * 1024,
						},
					},
				},
				Logger: testLog,
			}
			al.initRouter()

			jp := NewJobProviderMock()
			jp.ReturnErr = tc.jpReturnSaveErr
			al.jobProvider = jp

			connMock.ReturnErr = tc.connReturnErr
			connMock.ReturnOk = !tc.connReturnNotOk
			if len(tc.connReturnResp) > 0 {
				connMock.ReturnResponsePayload = tc.connReturnResp // override stubbed success payload
			}

			ctx := api.WithUser(context.Background(), testUser)
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/clients/%s/commands", tc.cid), strings.NewReader(tc.requestBody))
			req = req.WithContext(ctx)

			// when
			w := httptest.NewRecorder()
			al.router.ServeHTTP(w, req)

			// then
			assert.Equal(t, tc.wantStatusCode, w.Code)
			if tc.wantErrTitle == "" {
				// success case
				assert.Equal(t, fmt.Sprintf("{\"data\":{\"jid\":\"%s\"}}", testJID), w.Body.String())
				gotRunningJob := jp.InputCreateJob
				assert.NotNil(t, gotRunningJob)
				assert.Equal(t, testJID, gotRunningJob.JID)
				assert.Equal(t, models.JobStatusRunning, gotRunningJob.Status)
				assert.Nil(t, gotRunningJob.FinishedAt)
				assert.Equal(t, tc.cid, gotRunningJob.ClientID)
				assert.Equal(t, gotCmd, gotRunningJob.Command)
				assert.Equal(t, tc.wantInterpreter, gotRunningJob.Interpreter)
				assert.Equal(t, &sshSuccessResp.Pid, gotRunningJob.PID)
				assert.Equal(t, sshSuccessResp.StartedAt, gotRunningJob.StartedAt)
				assert.Equal(t, testUser, gotRunningJob.CreatedBy)
				assert.Equal(t, tc.wantTimeout, gotRunningJob.TimeoutSec)
				assert.Nil(t, gotRunningJob.Result)
			} else {
				// failure case
				wantResp := api.NewErrAPIPayloadFromMessage(tc.wantErrCode, tc.wantErrTitle, tc.wantErrDetail)
				wantRespBytes, err := json.Marshal(wantResp)
				require.NoError(t, err)
				require.Equal(t, string(wantRespBytes), w.Body.String())
			}
		})
	}
}

func TestHandleGetCommand(t *testing.T) {
	wantJob := jb.New(t).ClientID("cid-1234").JID("jid-1234").Build()
	wantJobResp := api.NewSuccessPayload(wantJob)
	b, err := json.Marshal(wantJobResp)
	require.NoError(t, err)
	wantJobRespJSON := string(b)

	testCases := []struct {
		name string

		jpReturnErr error
		jpReturnJob *models.Job

		wantStatusCode int
		wantErrCode    string
		wantErrTitle   string
		wantErrDetail  string
	}{
		{
			name:           "job found",
			jpReturnJob:    wantJob,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "not found",
			jpReturnJob:    nil,
			wantStatusCode: http.StatusNotFound,
			wantErrTitle:   fmt.Sprintf("Job[id=%q] not found.", wantJob.JID),
		},
		{
			name:           "error on get job",
			jpReturnErr:    errors.New("get job fake error"),
			wantStatusCode: http.StatusInternalServerError,
			wantErrTitle:   fmt.Sprintf("Failed to find a job[id=%q].", wantJob.JID),
			wantErrDetail:  "get job fake error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			al := APIListener{
				insecureForTests: true,
				Logger:           testLog,
				Server: &Server{
					config: &chconfig.Config{
						API: chconfig.APIConfig{
							MaxRequestBytes: 1024 * 1024,
						},
					},
				},
			}
			al.initRouter()

			jp := NewJobProviderMock()
			jp.ReturnErr = tc.jpReturnErr
			jp.ReturnJob = tc.jpReturnJob
			al.jobProvider = jp

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/clients/%s/commands/%s", wantJob.ClientID, wantJob.JID), nil)

			// when
			w := httptest.NewRecorder()
			al.router.ServeHTTP(w, req)

			// then
			assert.Equal(t, tc.wantStatusCode, w.Code)
			if tc.wantErrTitle == "" {
				// success case
				assert.Equal(t, wantJobRespJSON, w.Body.String())
				assert.Equal(t, wantJob.ClientID, jp.InputCID)
				assert.Equal(t, wantJob.JID, jp.InputJID)
			} else {
				// failure case
				wantResp := api.NewErrAPIPayloadFromMessage(tc.wantErrCode, tc.wantErrTitle, tc.wantErrDetail)
				wantRespBytes, err := json.Marshal(wantResp)
				require.NoError(t, err)
				require.Equal(t, string(wantRespBytes), w.Body.String())
			}
		})
	}
}

func TestHandleGetCommands(t *testing.T) {
	ft := time.Date(2020, 10, 10, 10, 10, 10, 0, time.UTC)
	testCID := "cid-1234"
	jb := jb.New(t).ClientID(testCID)
	job1 := jb.Status(models.JobStatusSuccessful).FinishedAt(ft).Build()
	job2 := jb.Status(models.JobStatusUnknown).FinishedAt(ft.Add(-time.Hour)).Build()
	job3 := jb.Status(models.JobStatusFailed).FinishedAt(ft.Add(time.Minute)).Build()
	job4 := jb.Status(models.JobStatusRunning).Build()
	wantResp1 := fmt.Sprintf(
		`{"data":[{"jid":"%s"},{"jid":"%s"},{"jid":"%s"},{"jid":"%s"}], "meta": {"count": 4}}`,
		job1.JID,
		job2.JID,
		job3.JID,
		job4.JID,
	)
	wantResp2 := fmt.Sprintf(
		`{"data":[{"jid":"%s", "finished_at": "%s", "status": "%s", "result":{"summary":"%s"}}], "meta": {"count": 1}}`,
		job1.JID,
		job1.FinishedAt.Format(time.RFC3339),
		job1.Status,
		job1.Result.Summary,
	)

	testCases := []struct {
		name   string
		params string

		jpReturnErr  error
		jpReturnJobs []*models.Job

		wantStatusCode  int
		wantSuccessResp string
		wantErrCode     string
		wantErrTitle    string
		wantErrDetail   string
	}{
		{
			name:            "found few jobs, jid only",
			params:          "fields[commands]=jid",
			jpReturnJobs:    []*models.Job{job1, job2, job3, job4},
			wantSuccessResp: wantResp1,
			wantStatusCode:  http.StatusOK,
		},
		{
			name:            "found one job, default fields",
			jpReturnJobs:    []*models.Job{job1},
			wantSuccessResp: wantResp2,
			wantStatusCode:  http.StatusOK,
		},
		{
			name:            "not found",
			jpReturnJobs:    []*models.Job{},
			wantSuccessResp: `{"data":[], "meta": {"count": 0}}`,
			wantStatusCode:  http.StatusOK,
		},
		{
			name:           "error on get job list",
			jpReturnErr:    errors.New("get job list fake error"),
			wantStatusCode: http.StatusInternalServerError,
			wantErrTitle:   fmt.Sprintf("Failed to get client jobs: client_id=%q.", testCID),
			wantErrDetail:  "get job list fake error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			al := APIListener{
				insecureForTests: true,
				Logger:           testLog,
				Server: &Server{
					config: &chconfig.Config{
						API: chconfig.APIConfig{
							MaxRequestBytes: 1024 * 1024,
						},
					},
				},
			}
			al.initRouter()

			jp := NewJobProviderMock()
			jp.ReturnErr = tc.jpReturnErr
			jp.ReturnJobList = tc.jpReturnJobs
			al.jobProvider = jp

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/clients/%s/commands?%s", testCID, tc.params), nil)

			// when
			w := httptest.NewRecorder()
			al.router.ServeHTTP(w, req)

			// then
			assert.Equal(t, tc.wantStatusCode, w.Code)
			if tc.wantErrTitle == "" {
				// success case
				assert.JSONEq(t, tc.wantSuccessResp, w.Body.String())
				assert.Equal(t, testCID, jp.InputCID)
			} else {
				// failure case
				wantResp := api.NewErrAPIPayloadFromMessage(tc.wantErrCode, tc.wantErrTitle, tc.wantErrDetail)
				wantRespBytes, err := json.Marshal(wantResp)
				require.NoError(t, err)
				require.Equal(t, string(wantRespBytes), w.Body.String())
			}
		})
	}
}

func TestHandlePostMultiClientCommand(t *testing.T) {
	testUser := "test-user"
	curUser := &users.User{
		Username: testUser,
		Groups:   []string{users.Administrators},
	}

	connMock1 := test.NewConnMock()
	// by default set to return success
	connMock1.ReturnOk = true
	sshSuccessResp1 := comm.RunCmdResponse{Pid: 1, StartedAt: time.Date(2020, 10, 10, 10, 10, 1, 0, time.UTC)}
	sshRespBytes1, err := json.Marshal(sshSuccessResp1)
	require.NoError(t, err)
	connMock1.ReturnResponsePayload = sshRespBytes1

	connMock2 := test.NewConnMock()
	// by default set to return success
	connMock2.ReturnOk = true
	sshSuccessResp2 := comm.RunCmdResponse{Pid: 2, StartedAt: time.Date(2020, 10, 10, 10, 10, 2, 0, time.UTC)}
	sshRespBytes2, err := json.Marshal(sshSuccessResp2)
	require.NoError(t, err)
	connMock2.ReturnResponsePayload = sshRespBytes2

	c1 := clients.New(t).ID("client-1").Connection(connMock1).Logger(testLog).Build()
	c2 := clients.New(t).ID("client-2").Connection(connMock2).Logger(testLog).Build()
	c3 := clients.New(t).ID("client-3").DisconnectedDuration(5 * time.Minute).Logger(testLog).Build()

	c1.Logger = testLog
	c2.Logger = testLog
	c3.Logger = testLog

	defaultTimeout := 60
	gotCmd := "/bin/date;foo;whoami"
	gotCmdTimeoutSec := 30
	validReqBody := `{"command": "` + gotCmd +
		`","timeout_sec": ` + strconv.Itoa(gotCmdTimeoutSec) +
		`,"client_ids": ["` + c1.GetID() + `", "` + c2.GetID() + `"]` +
		`,"abort_on_error": false` +
		`,"execute_concurrently": false` +
		`}`

	testCases := []struct {
		name string

		requestBody string

		connReturnErr error

		wantStatusCode int
		wantErrCode    string
		wantErrTitle   string
		wantErrDetail  string
		wantJobStatus  []string
		wantJobErr     string
	}{
		{
			name:           "valid cmd",
			requestBody:    validReqBody,
			wantStatusCode: http.StatusOK,
			wantJobStatus:  []string{models.JobStatusRunning, models.JobStatusRunning},
		},
		{
			name: "no targeting params provided",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30
		}`,
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Missing targeting parameters.",
			wantErrDetail:  ErrRequestMissingTargetingParams.Error(),
		},
		{
			name: "disconnected client",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30,
			"client_ids": ["client-3", "client-1"]
		}`,
			wantStatusCode: http.StatusOK,
			wantJobStatus:  []string{models.JobStatusFailed, models.JobStatusRunning},
			wantJobErr:     "client is not connected",
		},
		{
			name: "client not found",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30,
			"client_ids": ["client-1", "client-4"]
		}`,
			wantStatusCode: http.StatusNotFound,
			wantErrTitle:   fmt.Sprintf("Client with id=%q not found.", "client-4"),
		},
		{
			name:           "error on send request",
			requestBody:    validReqBody,
			connReturnErr:  errors.New("send fake error"),
			wantStatusCode: http.StatusOK,
			wantJobStatus:  []string{models.JobStatusFailed, models.JobStatusRunning},
			wantJobErr:     "failed to send request: send fake error",
		},
		{
			name: "error on send request, abort on err",
			requestBody: `
			{
				"command": "/bin/date;foo;whoami",
				"timeout_sec": 30,
				"client_ids": ["client-1", "client-2"],
				"execute_concurrently": false,
				"abort_on_error": true
			}`,
			connReturnErr:  errors.New("send fake error"),
			wantStatusCode: http.StatusOK,
			wantJobStatus:  []string{models.JobStatusFailed},
			wantJobErr:     "failed to send request: send fake error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			al := APIListener{
				insecureForTests: true,
				Server: &Server{
					clientService: clients.NewClientService(nil, nil, clients.NewClientRepository([]*clientdata.Client{c1, c2, c3}, &hour, testLog), testLog, nil),
					config: &chconfig.Config{
						Server: chconfig.ServerConfig{
							RunRemoteCmdTimeoutSec: defaultTimeout,
						},
						API: chconfig.APIConfig{
							MaxRequestBytes: 1024 * 1024,
						},
					},
					jobsDoneChannel: jobResultChanMap{
						m: make(map[string]chan *models.Job),
					},
					clientGroupProvider: mockClientGroupProvider{},
				},
				userService: users.NewAPIService(users.NewStaticProvider([]*users.User{curUser}), false, 0, -1),
				Logger:      testLog,
			}
			var done chan bool
			if tc.wantStatusCode == http.StatusOK {
				done = make(chan bool)
				al.testDone = done
			}

			al.initRouter()

			jobsDB, err := sqlite.New(
				":memory:",
				jobsmigration.AssetNames(),
				jobsmigration.Asset,
				DataSourceOptions,
			)
			require.NoError(t, err)
			jp := jobs.NewSqliteProvider(jobsDB, testLog)
			defer jp.Close()
			al.jobProvider = jp

			connMock1.ReturnErr = tc.connReturnErr

			ctx := api.WithUser(context.Background(), testUser)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/commands", strings.NewReader(tc.requestBody))
			req = req.WithContext(ctx)

			// when
			w := httptest.NewRecorder()
			al.router.ServeHTTP(w, req)

			// then
			assert.Equal(t, tc.wantStatusCode, w.Code)
			if tc.wantStatusCode == http.StatusOK {
				// wait until async task executeMultiClientJob finishes
				<-al.testDone
				// success case
				assert.Contains(t, w.Body.String(), `{"data":{"jid":`)
				gotResp := api.NewSuccessPayload(newJobResponse{})
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &gotResp))
				gotPropMap, ok := gotResp.Data.(map[string]interface{})
				require.True(t, ok)
				jidObj, found := gotPropMap["jid"]
				require.True(t, found)
				gotJID, ok := jidObj.(string)
				require.True(t, ok)
				require.NotEmpty(t, gotJID)

				gotMultiJob, err := jp.GetMultiJob(ctx, gotJID)
				require.NoError(t, err)
				require.NotNil(t, gotMultiJob)
				require.Len(t, gotMultiJob.Jobs, len(tc.wantJobStatus))
				for i := range gotMultiJob.Jobs {
					assert.Equal(t, tc.wantJobStatus[i], gotMultiJob.Jobs[i].Status)
					if tc.wantJobStatus[i] == models.JobStatusFailed {
						assert.Equal(t, tc.wantJobErr, gotMultiJob.Jobs[i].Error)
					}
				}
			} else {
				// failure case
				wantResp := api.NewErrAPIPayloadFromMessage(tc.wantErrCode, tc.wantErrTitle, tc.wantErrDetail)
				wantRespBytes, err := json.Marshal(wantResp)
				require.NoError(t, err)
				require.Equal(t, string(wantRespBytes), w.Body.String())
			}
		})
	}
}

func TestHandlePostMultiClientCommandWithPausedClient(t *testing.T) {
	testUser := "test-user"
	curUser := &users.User{
		Username: testUser,
		Groups:   []string{users.Administrators},
	}

	connMock1 := test.NewConnMock()
	// by default set to return success
	connMock1.ReturnOk = true
	sshSuccessResp1 := comm.RunCmdResponse{Pid: 1, StartedAt: time.Date(2020, 10, 10, 10, 10, 1, 0, time.UTC)}
	sshRespBytes1, err := json.Marshal(sshSuccessResp1)
	require.NoError(t, err)
	connMock1.ReturnResponsePayload = sshRespBytes1

	connMock2 := test.NewConnMock()
	// by default set to return success
	connMock2.ReturnOk = true
	sshSuccessResp2 := comm.RunCmdResponse{Pid: 2, StartedAt: time.Date(2020, 10, 10, 10, 10, 2, 0, time.UTC)}
	sshRespBytes2, err := json.Marshal(sshSuccessResp2)
	require.NoError(t, err)
	connMock2.ReturnResponsePayload = sshRespBytes2

	c1 := clients.New(t).ID("client-1").Connection(connMock1).Logger(testLog).Build()
	c1.Logger = testLog
	c1.SetPaused(true, clientdata.PausedDueToMaxClientsExceeded)

	c2 := clients.New(t).ID("client-2").Connection(connMock2).Logger(testLog).Build()
	c2.Logger = testLog

	defaultTimeout := 60
	gotCmd := "/bin/date;foo;whoami"
	gotCmdTimeoutSec := 30

	c1ValidReqBody := `{"command": "` + gotCmd +
		`","timeout_sec": ` + strconv.Itoa(gotCmdTimeoutSec) +
		`,"client_ids": ["` + c1.GetID() + `"]` +
		`,"abort_on_error": false` +
		`,"execute_concurrently": false` +
		`}`

	c2ValidReqBody := `{"command": "` + gotCmd +
		`","timeout_sec": ` + strconv.Itoa(gotCmdTimeoutSec) +
		`,"client_ids": ["` + c2.GetID() + `"]` +
		`,"abort_on_error": false` +
		`,"execute_concurrently": false` +
		`}`

	testCases := []struct {
		name   string
		client *clientdata.Client

		requestBody string
		abortOnErr  bool

		jobFailedErr error

		wantErrCode   string
		wantErrTitle  string
		wantErrDetail string
		wantJobErr    string
	}{
		{
			name:         "valid cmd with paused client",
			client:       c1,
			requestBody:  c1ValidReqBody,
			jobFailedErr: errors.New("client is paused (reason = unlicensed)"),
		},
		{
			name:        "valid cmd with ok client",
			client:      c2,
			requestBody: c2ValidReqBody,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given
			al := APIListener{
				insecureForTests: true,
				Server: &Server{
					clientService: clients.NewClientService(nil, nil, clients.NewClientRepository([]*clientdata.Client{c1, c2}, &hour, testLog), testLog, nil),
					config: &chconfig.Config{
						Server: chconfig.ServerConfig{
							RunRemoteCmdTimeoutSec: defaultTimeout,
						},
						API: chconfig.APIConfig{
							MaxRequestBytes: 1024 * 1024,
						},
					},
					jobsDoneChannel: jobResultChanMap{
						m: make(map[string]chan *models.Job),
					},
					clientGroupProvider: mockClientGroupProvider{},
				},
				userService: users.NewAPIService(users.NewStaticProvider([]*users.User{curUser}), false, 0, -1),
				Logger:      testLog,
			}

			done := make(chan bool)
			al.testDone = done

			al.initRouter()

			jobsDB, err := sqlite.New(
				":memory:",
				jobsmigration.AssetNames(),
				jobsmigration.Asset,
				DataSourceOptions,
			)
			require.NoError(t, err)
			jp := jobs.NewSqliteProvider(jobsDB, testLog)
			defer jp.Close()
			al.jobProvider = jp

			ctx := api.WithUser(context.Background(), testUser)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/commands", strings.NewReader(tc.requestBody))
			req = req.WithContext(ctx)

			// when
			w := httptest.NewRecorder()
			al.router.ServeHTTP(w, req)

			// then
			assert.Equal(t, http.StatusOK, w.Code)

			// wait until async task executeMultiClientJob finishes
			<-al.testDone
			// success case
			assert.Contains(t, w.Body.String(), `{"data":{"jid":`)

			gotResp := api.NewSuccessPayload(newJobResponse{})
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &gotResp))
			gotPropMap, ok := gotResp.Data.(map[string]interface{})
			require.True(t, ok)
			jidObj, found := gotPropMap["jid"]
			require.True(t, found)
			gotJID, ok := jidObj.(string)
			require.True(t, ok)
			require.NotEmpty(t, gotJID)

			gotMultiJob, err := jp.GetMultiJob(ctx, gotJID)
			require.NoError(t, err)
			require.NotNil(t, gotMultiJob)

			if !tc.client.IsPaused() {
				require.Len(t, gotMultiJob.Jobs, 1)
			}

			if tc.jobFailedErr != nil {
				assert.Equal(t, models.JobStatusFailed, gotMultiJob.Jobs[0].Status)
				assert.Contains(t, gotMultiJob.Jobs[0].Error, tc.jobFailedErr.Error())
			} else {
				assert.Equal(t, models.JobStatusRunning, gotMultiJob.Jobs[0].Status)
			}

		})
	}
}

func TestHandlePostMultiClientCommandWithGroupIDs(t *testing.T) {
	testUser := "test-user"
	defaultTimeout := 60

	testCases := []struct {
		name string

		requestBody string

		wantStatusCode int
		wantJobCount   int
		wantErrCode    string
		wantErrTitle   string
		wantErrDetail  string
	}{
		{
			name: "valid when group id with 2 clients",
			requestBody: `{
				"command": "/bin/date;foo;whoami",
				"timeout_sec": 30,
				"group_ids": ["group-1"],
				"abort_on_error": false,
				"execute_concurrently": false
			}`,
			wantStatusCode: http.StatusOK,
			wantJobCount:   2,
		},
		{
			name: "invalid when empty group ids",
			requestBody: `{
				"command": "/bin/date;foo;whoami",
				"timeout_sec": 30,
				"group_ids": [],
				"abort_on_error": false,
				"execute_concurrently": false
			}`,
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "at least 1 client should be specified",
		},
		{
			name: "valid when group id with 1 client",
			requestBody: `{
				"command": "/bin/date;foo;whoami",
				"timeout_sec": 30,
				"group_ids": ["group-2"],
				"abort_on_error": false,
				"execute_concurrently": false
			}`,
			wantStatusCode: http.StatusOK,
			wantJobCount:   1,
		},
		{
			name: "valid when group id and client id",
			requestBody: `{
				"command": "/bin/date;foo;whoami",
				"timeout_sec": 30,
				"client_ids": ["client-1"],
				"group_ids": ["group-2"],
				"abort_on_error": false,
				"execute_concurrently": false
			}`,
			wantStatusCode: http.StatusOK,
			wantJobCount:   2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			curUser := makeTestUser(testUser)

			connMock1 := makeConnMock(t, 1, time.Date(2020, 10, 10, 10, 10, 1, 0, time.UTC))
			connMock2 := makeConnMock(t, 2, time.Date(2020, 10, 10, 10, 10, 2, 0, time.UTC))
			connMock4 := makeConnMock(t, 4, time.Date(2020, 10, 10, 10, 10, 4, 0, time.UTC))

			c1 := clients.New(t).ID("client-1").Connection(connMock1).Logger(testLog).Build()
			c2 := clients.New(t).ID("client-2").Connection(connMock2).Logger(testLog).Build()
			c3 := clients.New(t).ID("client-3").DisconnectedDuration(5 * time.Minute).Logger(testLog).Build()
			c4 := clients.New(t).ID("client-4").Connection(connMock4).Logger(testLog).Build()

			g1 := makeClientGroup("group-1", &cgroups.ClientParams{
				ClientID: &cgroups.ParamValues{"client-1", "client-2"},
				OS:       &cgroups.ParamValues{"Linux*"},
				Version:  &cgroups.ParamValues{"0.1.1*"},
			})

			g2 := makeClientGroup("group-2", &cgroups.ClientParams{
				ClientID: &cgroups.ParamValues{"client-4"},
				OS:       &cgroups.ParamValues{"Linux*"},
				Version:  &cgroups.ParamValues{"0.1.1*"},
			})

			c1.SetAllowedUserGroups([]string{"group-1"})
			c2.SetAllowedUserGroups([]string{"group-1"})
			c4.SetAllowedUserGroups([]string{"group-2"})

			al := makeAPIListener(curUser,
				clients.NewClientRepository([]*clientdata.Client{c1, c2, c3, c4}, &hour, testLog),
				defaultTimeout,
				nil,
				testLog)

			var done chan bool
			if tc.wantStatusCode == http.StatusOK {
				done = make(chan bool)
				al.testDone = done
			}

			jp := makeJobsProvider(t, DataSourceOptions, testLog)
			defer jp.Close()

			gp := makeGroupsProvider(t, DataSourceOptions)
			defer gp.Close()

			al.initRouter()

			al.jobProvider = jp
			al.clientGroupProvider = gp

			ctx := api.WithUser(context.Background(), testUser)

			err := gp.Create(ctx, g1)
			assert.NoError(t, err)
			err = gp.Create(ctx, g2)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/commands", strings.NewReader(tc.requestBody))
			req = req.WithContext(ctx)

			// when
			w := httptest.NewRecorder()
			al.router.ServeHTTP(w, req)

			// then
			assert.Equal(t, tc.wantStatusCode, w.Code)
			if tc.wantStatusCode == http.StatusOK {
				// wait until async task executeMultiClientJob finishes
				<-al.testDone // success case
				assert.Contains(t, w.Body.String(), `{"data":{"jid":`)
				gotResp := api.NewSuccessPayload(newJobResponse{})
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &gotResp))
				gotPropMap, ok := gotResp.Data.(map[string]interface{})
				require.True(t, ok)
				jidObj, found := gotPropMap["jid"]
				require.True(t, found)
				gotJID, ok := jidObj.(string)
				require.True(t, ok)
				require.NotEmpty(t, gotJID)

				gotMultiJob, err := jp.GetMultiJob(ctx, gotJID)
				require.NoError(t, err)
				require.NotNil(t, gotMultiJob)
				require.Len(t, gotMultiJob.Jobs, tc.wantJobCount)
			} else {
				// failure case
				wantResp := api.NewErrAPIPayloadFromMessage(tc.wantErrCode, tc.wantErrTitle, tc.wantErrDetail)
				wantRespBytes, err := json.Marshal(wantResp)
				require.NoError(t, err)
				require.Equal(t, string(wantRespBytes), w.Body.String())
			}
		})
	}
}

func TestHandlePostMultiClientCommandWithTags(t *testing.T) {
	testUser := "test-user"
	defaultTimeout := 60

	testCases := []struct {
		name string

		requestBody string

		wantStatusCode int
		wantJobCount   int
		wantErrCode    string
		wantErrTitle   string
		wantErrDetail  string
	}{
		{
			name: "valid when only tags included",
			requestBody: `{
				"command": "/bin/date;foo;whoami",
				"timeout_sec": 30,
				"tags": {
					"tags": [
						"linux"
					],
					"operator": "OR"
				},
				"abort_on_error": false,
				"execute_concurrently": false
			}`,
			wantStatusCode: http.StatusOK,
			wantJobCount:   2,
		},
		{
			name: "valid when only tags included and missing operator",
			requestBody: `{
				"command": "/bin/date;foo;whoami",
				"timeout_sec": 30,
				"tags": {
					"tags": [
						"linux"
					]
				},
				"abort_on_error": false,
				"execute_concurrently": false
			}`,
			wantStatusCode: http.StatusOK,
			wantJobCount:   2,
		},
		{
			name: "error when client ids and tags included",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30,
			"client_ids": ["client-1", "client-2"],
			"tags": {
				"tags": [
					"linux",
					"windows"
				],
				"operator": "OR"
			}
		}`,
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Multiple targeting parameters.",
			wantErrDetail:  ErrRequestIncludesMultipleTargetingParams.Error(),
		},
		{
			name: "error when empty tags",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30,
			"tags": {
				"tags": [],
				"operator": "OR"
			}
		}`,
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "No tags specified.",
			wantErrDetail:  ErrMissingTagsInMultiJobRequest.Error(),
		},
		{
			name: "error when no clients for tag",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30,
			"tags": {
				"tags": ["random"],
				"operator": "OR"
			}
		}`,
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "at least 1 client should be specified",
		},
		{
			name: "error when group ids and tags included",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30,
			"group_ids": ["group-1"],
			"tags": {
				"tags": [
					"linux",
					"windows"
				],
				"operator": "OR"
			}
		}`,
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Multiple targeting parameters.",
			wantErrDetail:  ErrRequestIncludesMultipleTargetingParams.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			curUser := makeTestUser(testUser)

			connMock1 := makeConnMock(t, 1, time.Date(2020, 10, 10, 10, 10, 1, 0, time.UTC))
			connMock2 := makeConnMock(t, 2, time.Date(2020, 10, 10, 10, 10, 2, 0, time.UTC))
			connMock4 := makeConnMock(t, 4, time.Date(2020, 10, 10, 10, 10, 4, 0, time.UTC))

			c1 := clients.New(t).ID("client-1").Connection(connMock1).Logger(testLog).Build()
			c2 := clients.New(t).ID("client-2").Connection(connMock2).Logger(testLog).Build()
			c3 := clients.New(t).ID("client-3").DisconnectedDuration(5 * time.Minute).Logger(testLog).Build()
			c4 := clients.New(t).ID("client-4").Connection(connMock4).Logger(testLog).Build()

			c1.SetTags([]string{"linux"})
			c2.SetTags([]string{"windows"})
			c3.SetTags([]string{"mac"})
			c4.SetTags([]string{"linux", "windows"})

			g1 := makeClientGroup("group-1", &cgroups.ClientParams{
				ClientID: &cgroups.ParamValues{"client-1", "client-2"},
				OS:       &cgroups.ParamValues{"Linux*"},
				Version:  &cgroups.ParamValues{"0.1.1*"},
			})

			g2 := makeClientGroup("group-2", &cgroups.ClientParams{
				ClientID: &cgroups.ParamValues{"client-4"},
				OS:       &cgroups.ParamValues{"Linux*"},
				Version:  &cgroups.ParamValues{"0.1.1*"},
			})

			c1.SetAllowedUserGroups([]string{"group-1"})
			c2.SetAllowedUserGroups([]string{"group-1"})
			c4.SetAllowedUserGroups([]string{"group-2"})

			clientList := []*clientdata.Client{c1, c2, c4}

			p := clients.NewFakeClientProvider(t, nil, nil)

			al := makeAPIListener(curUser,
				clients.NewClientRepositoryWithDB(nil, &hour, p, testLog),
				defaultTimeout,
				nil,
				testLog)

			// make sure the repo has the test clients
			for _, cl := range clientList {
				err := al.clientService.GetRepo().Save(cl)
				assert.NoError(t, err)
			}

			var done chan bool
			if tc.wantStatusCode == http.StatusOK {
				done = make(chan bool)
				al.testDone = done
			}

			jp := makeJobsProvider(t, DataSourceOptions, testLog)
			defer jp.Close()

			gp := makeGroupsProvider(t, DataSourceOptions)
			defer gp.Close()

			al.initRouter()

			al.jobProvider = jp
			al.clientGroupProvider = gp

			ctx := api.WithUser(context.Background(), testUser)

			err := gp.Create(ctx, g1)
			assert.NoError(t, err)
			err = gp.Create(ctx, g2)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/commands", strings.NewReader(tc.requestBody))
			req = req.WithContext(ctx)

			// when
			w := httptest.NewRecorder()
			al.router.ServeHTTP(w, req)

			// then
			assert.Equal(t, tc.wantStatusCode, w.Code)
			if tc.wantStatusCode == http.StatusOK {
				// wait until async task executeMultiClientJob finishes
				<-al.testDone
				// success case
				assert.Contains(t, w.Body.String(), `{"data":{"jid":`)
				gotResp := api.NewSuccessPayload(newJobResponse{})
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &gotResp))
				gotPropMap, ok := gotResp.Data.(map[string]interface{})
				require.True(t, ok)
				jidObj, found := gotPropMap["jid"]
				require.True(t, found)
				gotJID, ok := jidObj.(string)
				require.True(t, ok)
				require.NotEmpty(t, gotJID)

				gotMultiJob, err := jp.GetMultiJob(ctx, gotJID)
				require.NoError(t, err)
				require.NotNil(t, gotMultiJob)
				require.Len(t, gotMultiJob.Jobs, tc.wantJobCount)
			} else {
				// failure case
				wantResp := api.NewErrAPIPayloadFromMessage(tc.wantErrCode, tc.wantErrTitle, tc.wantErrDetail)
				wantRespBytes, err := json.Marshal(wantResp)
				require.NoError(t, err)
				require.Equal(t, string(wantRespBytes), w.Body.String())
			}
		})
	}
}

func TestHandlePostMultiClientWSCommandWithTags(t *testing.T) {
	testUser := "user1"
	testLongLivedPwd := "theprefi_mynicefi-xedl-enth-long-livedpasswor"
	mockTokenManager := authorization.NewManager(
		CommonAPITokenTestDb(t, "user1", "theprefi", "the name", authorization.APITokenReadWrite, "mynicefi-xedl-enth-long-livedpasswor")) // APIToken database

	defaultTimeout := 60

	testCases := []struct {
		name string

		requestBody string

		shouldSucceed bool
		wantJobCount  int
		wantErrCode   string
		wantErrTitle  string
		wantErrDetail string
	}{
		{
			name: "valid with client ids",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30,
			"client_ids": ["client-1", "client-2"],
			"abort_on_error": false,
			"execute_concurrently": false
		}`,
			shouldSucceed: true,
			wantJobCount:  2,
		},
		{
			name: "no targeting params provided",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30
		}`,
			shouldSucceed: false,
			wantErrDetail: ErrRequestMissingTargetingParams.Error(),
		},
		{
			name: "valid when only tags included",
			requestBody: `{
				"command": "/bin/date;foo;whoami",
				"timeout_sec": 30,
				"tags": {
					"tags": [
						"linux"
					],
					"operator": "OR"
				},
				"abort_on_error": false,
				"execute_concurrently": false
			}`,
			shouldSucceed: true,
			wantJobCount:  2,
		},
		{
			name: "error when empty tags",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30,
			"tags": {
				"tags": [],
				"operator": "OR"
			}
		}`,
			shouldSucceed: false,
			wantErrDetail: ErrMissingTagsInMultiJobRequest.Error(),
		},
		{
			name: "error when no clients for tag",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30,
			"tags": {
				"tags": ["random"],
				"operator": "OR"
			}
		}`,
			shouldSucceed: false,
			wantErrDetail: "at least 1 client should be specified",
		},
		{
			name: "error when client ids and tags included",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30,
			"client_ids": ["client-1", "client-2"],
			"tags": {
				"tags": [
					"linux",
					"windows"
				],
				"operator": "OR"
			}
		}`,
			shouldSucceed: false,
			wantErrDetail: ErrRequestIncludesMultipleTargetingParams.Error(),
		},
		{
			name: "error when group ids and tags included",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30,
			"group_ids": ["group-1"],
			"tags": {
				"tags": [
					"linux",
					"windows"
				],
				"operator": "OR"
			}
		}`,
			shouldSucceed: false,
			wantErrDetail: ErrRequestIncludesMultipleTargetingParams.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			curUser := makeTestUser(testUser)

			connMock1 := makeConnMock(t, 1, time.Date(2020, 10, 10, 10, 10, 1, 0, time.UTC))
			connMock2 := makeConnMock(t, 2, time.Date(2020, 10, 10, 10, 10, 2, 0, time.UTC))
			connMock4 := makeConnMock(t, 4, time.Date(2020, 10, 10, 10, 10, 4, 0, time.UTC))

			c1 := clients.New(t).ID("client-1").Connection(connMock1).Logger(testLog).Build()
			c2 := clients.New(t).ID("client-2").Connection(connMock2).Logger(testLog).Build()
			c3 := clients.New(t).ID("client-3").DisconnectedDuration(5 * time.Minute).Logger(testLog).Build()
			c4 := clients.New(t).ID("client-4").Connection(connMock4).Logger(testLog).Build()

			c1.SetTags([]string{"linux"})
			c2.SetTags([]string{"windows"})
			c3.SetTags([]string{"mac"})
			c4.SetTags([]string{"linux", "windows"})

			g1 := makeClientGroup("group-1", &cgroups.ClientParams{
				ClientID: &cgroups.ParamValues{"client-1", "client-2"},
				OS:       &cgroups.ParamValues{"Linux*"},
				Version:  &cgroups.ParamValues{"0.1.1*"},
			})

			g2 := makeClientGroup("group-2", &cgroups.ClientParams{
				ClientID: &cgroups.ParamValues{"client-4"},
				OS:       &cgroups.ParamValues{"Linux*"},
				Version:  &cgroups.ParamValues{"0.1.1*"},
			})

			c1.SetAllowedUserGroups([]string{"group-1"})
			c2.SetAllowedUserGroups([]string{"group-1"})
			c4.SetAllowedUserGroups([]string{"group-2"})

			clientList := []*clientdata.Client{c1, c2, c4}

			p := clients.NewFakeClientProvider(t, nil, nil)

			al := makeAPIListener(curUser,
				clients.NewClientRepositoryWithDB(nil, &hour, p, testLog),
				defaultTimeout,
				mockTokenManager,
				testLog)

			// make sure the repo has the test clients
			for _, cl := range clientList {
				err := al.clientService.GetRepo().Save(cl)
				assert.NoError(t, err)
			}

			var done chan bool
			if tc.shouldSucceed {
				done = make(chan bool)
				al.testDone = done
			}

			jp := makeJobsProvider(t, DataSourceOptions, testLog)
			defer jp.Close()

			gp := makeGroupsProvider(t, DataSourceOptions)
			defer gp.Close()

			al.initRouter()

			al.jobProvider = jp
			al.clientGroupProvider = gp

			ctx := api.WithUser(context.Background(), testUser)

			err := gp.Create(ctx, g1)
			assert.NoError(t, err)
			err = gp.Create(ctx, g2)
			assert.NoError(t, err)

			// setup a web socket server running the handler under test
			s := httptest.NewServer(al.wsAuth(http.HandlerFunc(al.handleCommandsWS)))
			defer s.Close()

			// prep the test user auth
			reqHeader := makeAuthHeader(testUser, testLongLivedPwd)

			// dial the test websocket server running the handler under test
			wsURL := httpToWS(t, s.URL)
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, reqHeader)
			assert.NoError(t, err)
			defer ws.Close()

			// send the request to the handler under test
			err = ws.WriteMessage(websocket.TextMessage, []byte(tc.requestBody))
			assert.NoError(t, err)

			if tc.shouldSucceed {
				<-al.testDone

				// gotta find the job id somehow. not available in the WS as that is updated via
				// the client listener which isn't running as part of the test.
				multiJobIDs := al.jobsDoneChannel.GetAllKeys()
				multiJobID := multiJobIDs[0]
				multiJob, err := jp.GetMultiJob(ctx, multiJobID)
				assert.NoError(t, err)

				assert.Equal(t, tc.wantJobCount, len(multiJob.Jobs))

				// ask the server to close the web socket
				err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				assert.NoError(t, err)
			} else {
				_, res, err := ws.ReadMessage()
				assert.NoError(t, err)

				result := strings.TrimSpace(string(res))
				wantResp := api.NewErrAPIPayloadFromMessage(tc.wantErrCode, tc.wantErrTitle, tc.wantErrDetail)
				wantRespBytes, err := json.Marshal(wantResp)
				require.NoError(t, err)
				require.Equal(t, string(wantRespBytes), result)
			}
		})
	}
}

func TestHandlePostMultiClientScriptWithTags(t *testing.T) {
	defaultTimeout := 60

	testCases := []struct {
		name string

		requestBody string

		wantStatusCode int
		wantJobCount   int
		wantErrCode    string
		wantErrTitle   string
		wantErrDetail  string
	}{
		{
			name: "valid when only tags included",
			requestBody: `{
				"script": "dGVzdC5zaA==",
				"timeout_sec": 30,
				"tags": {
					"tags": [
						"linux"
					],
					"operator": "OR"
				},
				"abort_on_error": false,
				"execute_concurrently": false
			}`,
			wantStatusCode: http.StatusOK,
			wantJobCount:   2,
		},
		{
			name: "no targeting params provided",
			requestBody: `
		{
			"script": "dGVzdC5zaA==",
			"timeout_sec": 30
		}`,
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Missing targeting parameters.",
			wantErrDetail:  ErrRequestMissingTargetingParams.Error(),
		},
		{
			name: "error when empty tags",
			requestBody: `
		{
			"script": "dGVzdC5zaA==",
			"timeout_sec": 30,
			"tags": {
				"tags": [],
				"operator": "OR"
			}
		}`,
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "No tags specified.",
			wantErrDetail:  ErrMissingTagsInMultiJobRequest.Error(),
		},
		{
			name: "error when no clients for tag",
			requestBody: `
		{
			"script": "dGVzdC5zaA==",
			"timeout_sec": 30,
			"tags": {
				"tags": ["random"],
				"operator": "OR"
			}
		}`,
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "at least 1 client should be specified",
		},
		{
			name: "error when client ids and tags included",
			requestBody: `
		{
			"script": "dGVzdC5zaA==",
			"timeout_sec": 30,
			"client_ids": ["client-1", "client-2"],
			"tags": {
				"tags": [
					"linux",
					"windows"
				],
				"operator": "OR"
			}
		}`,
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Multiple targeting parameters.",
			wantErrDetail:  ErrRequestIncludesMultipleTargetingParams.Error(),
		},
		{
			name: "error when group ids and tags included",
			requestBody: `
		{
			"script": "dGVzdC5zaA==",
			"timeout_sec": 30,
			"group_ids": ["group-1"],
			"tags": {
				"tags": [
					"linux",
					"windows"
				],
				"operator": "OR"
			}
		}`,
			wantStatusCode: http.StatusBadRequest,
			wantErrTitle:   "Multiple targeting parameters.",
			wantErrDetail:  ErrRequestIncludesMultipleTargetingParams.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given

			testUser := "test-user"
			curUser := makeTestUser(testUser)

			connMock1 := makeConnMock(t, 1, time.Date(2020, 10, 10, 10, 10, 1, 0, time.UTC))
			connMock2 := makeConnMock(t, 2, time.Date(2020, 10, 10, 10, 10, 2, 0, time.UTC))
			connMock4 := makeConnMock(t, 4, time.Date(2020, 10, 10, 10, 10, 4, 0, time.UTC))

			c1 := clients.New(t).ID("client-1").Connection(connMock1).Logger(testLog).Build()
			c2 := clients.New(t).ID("client-2").Connection(connMock2).Logger(testLog).Build()
			c3 := clients.New(t).ID("client-3").DisconnectedDuration(5 * time.Minute).Logger(testLog).Build()
			c4 := clients.New(t).ID("client-4").Connection(connMock4).Logger(testLog).Build()

			c1.SetTags([]string{"linux"})
			c2.SetTags([]string{"windows"})
			c3.SetTags([]string{"mac"})
			c4.SetTags([]string{"linux", "windows"})

			g1 := makeClientGroup("group-1", &cgroups.ClientParams{
				ClientID: &cgroups.ParamValues{"client-1", "client-2"},
				OS:       &cgroups.ParamValues{"Linux*"},
				Version:  &cgroups.ParamValues{"0.1.1*"},
			})

			g2 := makeClientGroup("group-2", &cgroups.ClientParams{
				ClientID: &cgroups.ParamValues{"client-4"},
				OS:       &cgroups.ParamValues{"Linux*"},
				Version:  &cgroups.ParamValues{"0.1.1*"},
			})

			c1.SetAllowedUserGroups([]string{"group-1"})
			c2.SetAllowedUserGroups([]string{"group-1"})
			c4.SetAllowedUserGroups([]string{"group-2"})

			clientList := []*clientdata.Client{c1, c2, c4}

			p := clients.NewFakeClientProvider(t, nil, nil)

			al := makeAPIListener(curUser,
				clients.NewClientRepositoryWithDB(nil, &hour, p, testLog),
				defaultTimeout,
				nil,
				testLog)

			// make sure the repo has the test clients
			for _, cl := range clientList {
				err := al.clientService.GetRepo().Save(cl)
				assert.NoError(t, err)
			}

			var done chan bool
			if tc.wantStatusCode == http.StatusOK {
				done = make(chan bool)
				al.testDone = done
			}

			jp := makeJobsProvider(t, DataSourceOptions, testLog)
			defer jp.Close()

			gp := makeGroupsProvider(t, DataSourceOptions)
			defer gp.Close()

			al.initRouter()

			al.jobProvider = jp
			al.clientGroupProvider = gp

			ctx := api.WithUser(context.Background(), testUser)

			err := gp.Create(ctx, g1)
			assert.NoError(t, err)
			err = gp.Create(ctx, g2)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/scripts", strings.NewReader(tc.requestBody))
			req = req.WithContext(ctx)

			// when
			w := httptest.NewRecorder()
			al.router.ServeHTTP(w, req)

			// then
			assert.Equal(t, tc.wantStatusCode, w.Code)
			if tc.wantStatusCode == http.StatusOK {
				// wait until async task executeMultiClientJob finishes
				<-al.testDone
				// success case
				assert.Contains(t, w.Body.String(), `{"data":{"jid":`)
				gotResp := api.NewSuccessPayload(newJobResponse{})
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &gotResp))
				gotPropMap, ok := gotResp.Data.(map[string]interface{})
				require.True(t, ok)
				jidObj, found := gotPropMap["jid"]
				require.True(t, found)
				gotJID, ok := jidObj.(string)
				require.True(t, ok)
				require.NotEmpty(t, gotJID)

				gotMultiJob, err := jp.GetMultiJob(ctx, gotJID)
				require.NoError(t, err)
				require.NotNil(t, gotMultiJob)
				require.Len(t, gotMultiJob.Jobs, tc.wantJobCount)
			} else {
				// failure case
				wantResp := api.NewErrAPIPayloadFromMessage(tc.wantErrCode, tc.wantErrTitle, tc.wantErrDetail)
				wantRespBytes, err := json.Marshal(wantResp)
				require.NoError(t, err)
				require.Equal(t, string(wantRespBytes), w.Body.String())
			}
		})
	}
}

func TestHandlePostMultiClientWSScriptWithTags(t *testing.T) {
	defaultTimeout := 60
	mockTokenManager := authorization.NewManager(
		CommonAPITokenTestDb(t, "user1", "theprefi", "the name", authorization.APITokenReadWrite, "mynicefi-xedl-enth-long-livedpasswor")) // APIToken database

	testCases := []struct {
		name string

		requestBody string

		shouldSucceed bool
		wantJobCount  int
		wantErrCode   string
		wantErrTitle  string
		wantErrDetail string
	}{
		{
			name: "valid with client ids",
			requestBody: `
		{
			"script": "dGVzdC5zaA==",
			"timeout_sec": 30,
			"client_ids": ["client-1", "client-2"],
			"abort_on_error": false,
			"execute_concurrently": false
		}`,
			shouldSucceed: true,
			wantJobCount:  2,
		},
		{
			name: "no targeting params provided",
			requestBody: `
		{
			"script": "dGVzdC5zaA==",
			"timeout_sec": 30
		}`,
			shouldSucceed: false,
			wantErrDetail: ErrRequestMissingTargetingParams.Error(),
		},
		{
			name: "valid when only tags included",
			requestBody: `{
				"script": "dGVzdC5zaA==",
				"timeout_sec": 30,
				"tags": {
					"tags": [
						"linux",
						"windows"
					],
					"operator": "AND"
				},
				"abort_on_error": false,
				"execute_concurrently": false
			}`,
			shouldSucceed: true,
			wantJobCount:  1,
		},
		{
			name: "error when empty tags",
			requestBody: `
		{
			"script": "dGVzdC5zaA==",
			"timeout_sec": 30,
			"tags": {
				"tags": [],
				"operator": "OR"
			}
		}`,
			shouldSucceed: false,
			wantErrDetail: ErrMissingTagsInMultiJobRequest.Error(),
		},
		{
			name: "error when no clients for tag",
			requestBody: `
		{
			"script": "dGVzdC5zaA==",
			"timeout_sec": 30,
			"tags": {
				"tags": ["random"],
				"operator": "OR"
			}
		}`,
			shouldSucceed: false,
			wantErrDetail: "at least 1 client should be specified",
		},
		{
			name: "error when client ids and tags included",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30,
			"client_ids": ["client-1", "client-2"],
			"tags": {
				"tags": [
					"linux",
					"windows"
				],
				"operator": "OR"
			}
		}`,
			shouldSucceed: false,
			wantErrDetail: ErrRequestIncludesMultipleTargetingParams.Error(),
		},
		{
			name: "error when group ids and tags included",
			requestBody: `
		{
			"command": "/bin/date;foo;whoami",
			"timeout_sec": 30,
			"group_ids": ["group-1"],
			"tags": {
				"tags": [
					"linux",
					"windows"
				],
				"operator": "OR"
			}
		}`,
			shouldSucceed: false,
			wantErrDetail: ErrRequestIncludesMultipleTargetingParams.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// given

			testUser := "user1"
			testLongLivedPwd := "theprefi_mynicefi-xedl-enth-long-livedpasswor"
			curUser := makeTestUser(testUser)

			connMock1 := makeConnMock(t, 1, time.Date(2020, 10, 10, 10, 10, 1, 0, time.UTC))
			connMock2 := makeConnMock(t, 2, time.Date(2020, 10, 10, 10, 10, 2, 0, time.UTC))
			connMock4 := makeConnMock(t, 4, time.Date(2020, 10, 10, 10, 10, 4, 0, time.UTC))

			c1 := clients.New(t).ID("client-1").Connection(connMock1).Logger(testLog).Build()
			c2 := clients.New(t).ID("client-2").Connection(connMock2).Logger(testLog).Build()
			c3 := clients.New(t).ID("client-3").DisconnectedDuration(5 * time.Minute).Logger(testLog).Build()
			c4 := clients.New(t).ID("client-4").Connection(connMock4).Logger(testLog).Build()

			c1.SetTags([]string{"linux"})
			c2.SetTags([]string{"windows"})
			c3.SetTags([]string{"mac"})
			c4.SetTags([]string{"linux", "windows"})

			g1 := makeClientGroup("group-1", &cgroups.ClientParams{
				ClientID: &cgroups.ParamValues{"client-1", "client-2"},
				OS:       &cgroups.ParamValues{"Linux*"},
				Version:  &cgroups.ParamValues{"0.1.1*"},
			})

			g2 := makeClientGroup("group-2", &cgroups.ClientParams{
				ClientID: &cgroups.ParamValues{"client-4"},
				OS:       &cgroups.ParamValues{"Linux*"},
				Version:  &cgroups.ParamValues{"0.1.1*"},
			})

			c1.SetAllowedUserGroups([]string{"group-1"})
			c2.SetAllowedUserGroups([]string{"group-1"})
			c4.SetAllowedUserGroups([]string{"group-2"})

			clientList := []*clientdata.Client{c1, c2, c4}

			p := clients.NewFakeClientProvider(t, nil, nil)

			al := makeAPIListener(curUser,
				clients.NewClientRepositoryWithDB(nil, &hour, p, testLog),
				defaultTimeout,
				mockTokenManager,
				testLog)

			// make sure the repo has the test clients
			for _, cl := range clientList {
				err := al.clientService.GetRepo().Save(cl)
				assert.NoError(t, err)
			}

			var done chan bool
			if tc.shouldSucceed {
				done = make(chan bool)
				al.testDone = done
			}

			jp := makeJobsProvider(t, DataSourceOptions, testLog)
			defer jp.Close()

			gp := makeGroupsProvider(t, DataSourceOptions)
			defer gp.Close()

			al.initRouter()

			al.jobProvider = jp
			al.clientGroupProvider = gp

			ctx := api.WithUser(context.Background(), testUser)

			err := gp.Create(ctx, g1)
			assert.NoError(t, err)
			err = gp.Create(ctx, g2)
			assert.NoError(t, err)

			// setup a web socket server running the handler under test
			s := httptest.NewServer(al.wsAuth(http.HandlerFunc(al.handleScriptsWS)))
			defer s.Close()

			// prep the test user auth
			reqHeader := makeAuthHeader(testUser, testLongLivedPwd)

			// dial the test websocket server running the handler under test
			wsURL := httpToWS(t, s.URL)
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, reqHeader)
			assert.NoError(t, err)
			defer ws.Close()

			// send the request to the handler under test
			err = ws.WriteMessage(websocket.TextMessage, []byte(tc.requestBody))
			assert.NoError(t, err)

			if tc.shouldSucceed {
				<-al.testDone

				// gotta find the job id somehow. not available in the WS as that is updated via
				// the client listener which isn't running as part of the test.
				multiJobIDs := al.jobsDoneChannel.GetAllKeys()
				multiJobID := multiJobIDs[0]
				multiJob, err := jp.GetMultiJob(ctx, multiJobID)
				assert.NoError(t, err)

				assert.Equal(t, tc.wantJobCount, len(multiJob.Jobs))

				// ask the server to close the web socket
				err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				assert.NoError(t, err)
			} else {
				_, res, err := ws.ReadMessage()
				assert.NoError(t, err)

				result := strings.TrimSpace(string(res))
				wantResp := api.NewErrAPIPayloadFromMessage(tc.wantErrCode, tc.wantErrTitle, tc.wantErrDetail)
				wantRespBytes, err := json.Marshal(wantResp)
				require.NoError(t, err)
				require.Equal(t, string(wantRespBytes), result)
			}
		})
	}
}

func makeTestUser(testUser string) (curUser *users.User) {
	curUser = &users.User{
		Username: testUser,
		Groups:   []string{users.Administrators},
	}
	return curUser
}

func makeAuthHeader(testUser string, testToken string) (reqHeader http.Header) {
	auth := testUser + ":" + testToken
	authContent := base64.StdEncoding.EncodeToString([]byte(auth))
	reqHeader = http.Header{}
	reqHeader.Add("Authorization", "Basic "+authContent)
	return reqHeader
}

func makeClientGroup(groupID string, params *cgroups.ClientParams) (gp *cgroups.ClientGroup) {
	gp = &cgroups.ClientGroup{
		ID:     groupID,
		Params: params,
	}
	return gp
}

func httpToWS(t *testing.T, u string) string {
	t.Helper()

	wsURL, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}

	switch wsURL.Scheme {
	case "http":
		wsURL.Scheme = "ws"
	case "https":
		wsURL.Scheme = "wss"
	}

	return wsURL.String()
}

func makeConnMock(t *testing.T, pid int, startedAt time.Time) (connMock *test.ConnMock) {
	t.Helper()
	connMock = test.NewConnMock()
	connMock.ReturnOk = true
	sshSuccessResp := comm.RunCmdResponse{Pid: pid, StartedAt: startedAt}
	sshRespBytes, err := json.Marshal(sshSuccessResp)
	require.NoError(t, err)
	connMock.ReturnResponsePayload = sshRespBytes
	return connMock
}

func makeAPIListener(
	curUser *users.User,
	clientRepo *clients.ClientRepository,
	defaultTimeout int,
	tokenManager *authorization.Manager,
	testLog *logger.Logger) (al *APIListener) {
	clientService := clients.NewClientService(nil, nil, clientRepo, testLog, nil)
	al = &APIListener{
		insecureForTests: true,
		Server: &Server{
			clientService: clientService,
			config: &chconfig.Config{
				Server: chconfig.ServerConfig{
					RunRemoteCmdTimeoutSec: defaultTimeout,
				},
				API: chconfig.APIConfig{
					MaxRequestBytes: 1024 * 1024,
				},
			},
			uiJobWebSockets: ws.NewWebSocketCache(),
			jobsDoneChannel: jobResultChanMap{
				m: make(map[string]chan *models.Job),
			},
		},
		bannedUsers:  security.NewBanList(time.Duration(60) * time.Second),
		tokenManager: tokenManager,
		userService:  users.NewAPIService(users.NewStaticProvider([]*users.User{curUser}), false, 0, -1),
		Logger:       testLog,
	}

	return al
}

func makeJobsProvider(t *testing.T, dataSourceOptions sqlite.DataSourceOptions, testLog *logger.Logger) (jp *jobs.SqliteProvider) {
	t.Helper()
	jobsDB, err := sqlite.New(
		":memory:",
		jobsmigration.AssetNames(),
		jobsmigration.Asset,
		dataSourceOptions,
	)
	require.NoError(t, err)
	jp = jobs.NewSqliteProvider(jobsDB, testLog)
	return jp
}

func makeGroupsProvider(t *testing.T, dataSourceOptions sqlite.DataSourceOptions) (gp *cgroups.SqliteProvider) {
	t.Helper()
	groupsDB, err := sqlite.New(
		":memory:",
		client_groups.AssetNames(),
		client_groups.Asset,
		dataSourceOptions,
	)
	require.NoError(t, err)

	gp, err = cgroups.NewSqliteProvider(groupsDB)
	assert.NoError(t, err)
	return gp
}

func makeScheduleManager(t *testing.T, jp *jobs.SqliteProvider, jobRunner schedule.JobRunner, testLog *logger.Logger) (scheduleManager *schedule.Manager) {
	t.Helper()
	scheduleManager = schedule.NewManager(jobRunner, jp.GetDB(), testLog, 30)

	return scheduleManager
}
