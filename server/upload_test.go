package chserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/riportdev/riport/server/api"
	"github.com/riportdev/riport/server/api/users"
	"github.com/riportdev/riport/server/chconfig"
	"github.com/riportdev/riport/server/clients"
	"github.com/riportdev/riport/server/clients/clientdata"
	"github.com/riportdev/riport/share/files"
	"github.com/riportdev/riport/share/models"
	"github.com/riportdev/riport/share/test"
)

func MockUserService(user string, group string) *users.APIService {
	curUser := &users.User{
		Username: user,
		Groups:   []string{group},
	}
	return users.NewAPIService(users.NewStaticProvider([]*users.User{curUser}), false, 0, -1)
}

func FsCallback(fs *test.FileAPIMock, t *testing.T) {
	fs.On("CreateDirIfNotExists", "/data/"+files.DefaultUploadTempFolder, files.DefaultMode).Return(true, nil)

	fileExpectation := func(f io.Reader) bool {
		actualFileContent, err := ioutil.ReadAll(f)

		require.NoError(t, err)

		return string(actualFileContent) == "some content"
	}
	fs.On("CreateFile", "/data/filepush/id-123_rport_filepush", mock.MatchedBy(fileExpectation)).Return(int64(10), nil)

	fileMock := &test.ReadWriteCloserMock{}
	fileMock.Reader = strings.NewReader("some content")
	fileMock.On("Close").Return(nil)

	fs.On("Open", "/data/filepush/id-123_rport_filepush").Return(fileMock, nil)
	fs.On("Remove", "/data/filepush/id-123_rport_filepush").Return(nil)
}

func TestHandleFileUploads(t *testing.T) {
	testCases := []struct {
		name                string
		wantStatus          int
		useFsCallback       bool
		wantResp            *models.UploadResponseShort
		wantClientInputFile *models.UploadedFile
		fileName            string
		fileContent         string
		formParts           map[string][]string
		cl                  *clientdata.Client
		clientTags          []string
		user                string
		group               string
		wantErrCode         string
		wantErrTitle        string
		wantErrDetail       string
	}{
		{
			name:       "send file success",
			wantStatus: http.StatusOK,
			user:       "admin",
			group:      users.Administrators,
			wantResp: &models.UploadResponseShort{
				ID:        "id-123",
				Filepath:  "/destination/myfile.txt",
				SizeBytes: 10,
			},
			useFsCallback: true,
			fileName:      "file.txt",
			fileContent:   "some content",
			cl:            clients.New(t).ID("22114341234").Logger(testLog).Build(),
			formParts: map[string][]string{
				"client_id": {
					"22114341234",
				},
				"dest": {
					"/destination/myfile.txt",
				},
				"id": {
					"id-123",
				},
				"user": {
					"admin",
				},
				"group": {
					"group",
				},
				"mode": {
					"0744",
				},
				"force": {
					"1",
				},
				"sync": {
					"1",
				},
			},
			wantClientInputFile: &models.UploadedFile{
				ID:                   "id-123",
				SourceFilePath:       "/data/filepush/id-123_rport_filepush",
				DestinationPath:      "/destination/myfile.txt",
				DestinationFileMode:  0744,
				DestinationFileOwner: "admin",
				DestinationFileGroup: "group",
				ForceWrite:           true,
				Sync:                 true,
				Md5Checksum:          test.Md5Hash("some content"),
			},
		},
		{
			name:       "send file success targeting tags",
			wantStatus: http.StatusOK,
			user:       "admin",
			group:      users.Administrators,
			wantResp: &models.UploadResponseShort{
				ID:        "id-123",
				Filepath:  "/destination/myfile.txt",
				SizeBytes: 10,
			},
			useFsCallback: true,
			fileName:      "file.txt",
			fileContent:   "some content",
			cl:            clients.New(t).ID("22114341234").Logger(testLog).Build(),
			clientTags:    []string{"linux"},
			formParts: map[string][]string{
				"tags": {
					`{
						"tags": ["linux"],
						"operator": "OR"
					}`,
				},
				"dest": {
					"/destination/myfile.txt",
				},
				"id": {
					"id-123",
				},
				"user": {
					"admin",
				},
				"group": {
					"group",
				},
				"mode": {
					"0744",
				},
				"force": {
					"1",
				},
				"sync": {
					"1",
				},
			},
			wantClientInputFile: &models.UploadedFile{
				ID:                   "id-123",
				SourceFilePath:       "/data/filepush/id-123_rport_filepush",
				DestinationPath:      "/destination/myfile.txt",
				DestinationFileMode:  0744,
				DestinationFileOwner: "admin",
				DestinationFileGroup: "group",
				ForceWrite:           true,
				Sync:                 true,
				Md5Checksum:          test.Md5Hash("some content"),
			},
		},
		{
			name:          "send file failed, multiple targeting params",
			wantStatus:    http.StatusBadRequest,
			user:          "admin",
			group:         "",
			wantResp:      &models.UploadResponseShort{},
			useFsCallback: true,
			fileName:      "file.txt",
			fileContent:   "some content",
			cl:            clients.New(t).ID("22114341234").Logger(testLog).Build(),
			formParts: map[string][]string{
				"client_id": {
					"22114341234",
				},
				"tags": {
					`{
						"tags": ["linux"],
						"operator": "OR"
					}`,
				},
				"dest": {
					"/destination/myfile.txt",
				},
				"id": {
					"id-123",
				},
			},
			wantClientInputFile: &models.UploadedFile{},
			wantErrTitle:        "Multiple targeting parameters.",
			wantErrDetail:       "multiple targeting options are not supported. Please specify only one",
		},
		{
			name:          "send file failed, missing tags element",
			wantStatus:    http.StatusBadRequest,
			user:          "admin",
			group:         "",
			wantResp:      &models.UploadResponseShort{},
			useFsCallback: true,
			fileName:      "file.txt",
			fileContent:   "some content",
			cl:            clients.New(t).ID("22114341234").Logger(testLog).Build(),
			formParts: map[string][]string{
				"dest": {
					"/destination/myfile.txt",
				},
				"id": {
					"id-123",
				},
			},
			wantClientInputFile: &models.UploadedFile{},
			wantErrTitle:        "Missing targeting parameters.",
			wantErrDetail:       "please specify targeting options, such as client ids, groups ids or tags",
		},
		{
			name:          "send file failed, empty tags",
			wantStatus:    http.StatusBadRequest,
			user:          "admin",
			group:         "",
			wantResp:      &models.UploadResponseShort{},
			useFsCallback: true,
			fileName:      "file.txt",
			fileContent:   "some content",
			cl:            clients.New(t).ID("22114341234").Logger(testLog).Build(),
			formParts: map[string][]string{
				"tags": {
					`{
						"tags": [],
						"operator": "OR"
					}`,
				},
				"dest": {
					"/destination/myfile.txt",
				},
				"id": {
					"id-123",
				},
			},
			wantClientInputFile: &models.UploadedFile{},
			wantErrTitle:        "No tags specified.",
			wantErrDetail:       "please specify tags in the tags list",
		},
		{
			name:          "send file denied, bad user rights",
			wantStatus:    http.StatusForbidden,
			user:          "loser",
			group:         "",
			wantResp:      &models.UploadResponseShort{},
			useFsCallback: true,
			fileName:      "file.txt",
			fileContent:   "some content",
			cl:            clients.New(t).ID("22114341234").Logger(testLog).Build(),
			formParts: map[string][]string{
				"client_id": {
					"22114341234",
				},
				"dest": {
					"/destination/myfile.txt",
				},
				"id": {
					"id-123",
				},
			},
			wantClientInputFile: &models.UploadedFile{},
			wantErrCode:         "ACCESS_CONTROL_VIOLATION",
			wantErrTitle:        "upload forbidden",
			wantErrDetail:       "Access denied to client(s) with ID(s): 22114341234",
		},
		{
			name:          "send file denied, bad destination",
			wantStatus:    http.StatusBadRequest,
			user:          "loser",
			group:         "",
			wantResp:      &models.UploadResponseShort{},
			useFsCallback: true,
			fileName:      "file.txt",
			fileContent:   "some content",
			cl:            clients.New(t).ID("22114341234").Logger(testLog).Build(),
			formParts: map[string][]string{
				"client_id": {
					"22114341234",
				},
				"dest": {
					"/proc/myfile.txt",
				},
				"id": {
					"id-123",
				},
			},
			wantClientInputFile: &models.UploadedFile{},
			wantErrCode:         "BAD_DESTINATION",
			wantErrTitle:        "upload denied",
			wantErrDetail:       "uploads to /proc/ are forbidden",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cl := tc.cl

			if tc.clientTags != nil {
				cl.SetTags(tc.clientTags)
			}

			connMock := test.NewConnMock()

			connMock.ReturnOk = true

			done := make(chan bool)
			connMock.DoneChannel = done

			cl.SetConnection(connMock)

			fileAPIMock := test.NewFileAPIMock()
			if tc.useFsCallback {
				FsCallback(fileAPIMock, t)
			}

			al := APIListener{
				insecureForTests: true,
				Server: &Server{
					clientService: clients.NewClientService(
						nil,
						nil,
						clients.NewClientRepository([]*clientdata.Client{cl}, &hour, testLog),
						testLog,
						nil,
					),
					clientGroupProvider: mockClientGroupProvider{},
					config: &chconfig.Config{
						Server: chconfig.ServerConfig{
							DataDir: "/data",
						},
						API: chconfig.APIConfig{
							MaxFilePushSize: int64(10 << 20),
						},
					},
					filesAPI: fileAPIMock,
				},
				Logger:      testLog,
				userService: MockUserService(tc.user, tc.group),
			}

			al.initRouter()

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("upload", tc.fileName)
			require.NoError(t, err)

			_, err = io.Copy(part, strings.NewReader(tc.fileContent))
			require.NoError(t, err)

			for key, vals := range tc.formParts {
				for _, val := range vals {
					err = writer.WriteField(key, val)
					require.NoError(t, err)
				}
			}

			err = writer.Close()
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/files", body)
			req.Header.Add("Content-Type", writer.FormDataContentType())
			ctx := api.WithUser(context.Background(), tc.user)
			req = req.WithContext(ctx)

			rec := httptest.NewRecorder()
			al.router.ServeHTTP(rec, req)

			t.Logf("Got response %s", rec.Body)
			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantErrTitle != "" {
				wantResp := api.NewErrAPIPayloadFromMessage(tc.wantErrCode, tc.wantErrTitle, tc.wantErrDetail)
				wantRespBytes, err := json.Marshal(wantResp)
				require.NoError(t, err)
				require.Equal(t, string(wantRespBytes), rec.Body.String())
				return
			}

			var successResp struct {
				Data *models.UploadResponseShort `json:"data"`
			}

			dec := json.NewDecoder(rec.Body)
			dec.DisallowUnknownFields()
			err = dec.Decode(&successResp)
			require.NoError(t, err)

			assert.Equal(t, tc.wantResp, successResp.Data)

			select {
			case <-done:
				assertClientPayload(t, connMock, tc.wantClientInputFile)
			case <-time.After(time.Second * 2):
				assertClientPayload(t, connMock, tc.wantClientInputFile)
			}
		})
	}
}

func assertClientPayload(t *testing.T, connMock *test.ConnMock, wantClientInputFile *models.UploadedFile) {
	name, wantReply, payload := connMock.InputSendRequest()

	actualInputFile := &models.UploadedFile{}
	err := actualInputFile.FromBytes(payload)
	require.NoError(t, err)

	assert.Equal(t, "upload", name)
	assert.Equal(t, wantClientInputFile, actualInputFile)
	assert.True(t, wantReply)
}
