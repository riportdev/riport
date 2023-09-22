package chserver

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	errors2 "github.com/riportdev/riport/server/api/errors"
	errors3 "github.com/riportdev/riport/share/errors"

	"github.com/riportdev/riport/server/api"
	"github.com/riportdev/riport/server/auditlog"
	"github.com/riportdev/riport/server/clients"
	"github.com/riportdev/riport/server/clients/clientdata"
	"github.com/riportdev/riport/share/comm"
	"github.com/riportdev/riport/share/files"
	"github.com/riportdev/riport/share/models"
	"github.com/riportdev/riport/share/random"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const uploadBufSize = 1000000 // 1Mb

type UploadRequest struct {
	File                 multipart.File
	FileHeader           *multipart.FileHeader
	ClientIDs            []string
	GroupIDs             []string
	ClientTags           *models.JobClientTags
	clientsInGroupsCount int
	Clients              []*clientdata.Client
	*models.UploadedFile
}

func (ur UploadRequest) GetClientIDs() (ids []string) {
	return ur.ClientIDs
}

func (ur UploadRequest) GetGroupIDs() (ids []string) {
	return ur.GroupIDs
}

func (ur UploadRequest) GetClientTags() (clientTags *models.JobClientTags) {
	return ur.ClientTags
}

func (al *APIListener) handleFileUploads(w http.ResponseWriter, req *http.Request) {
	uploadRequest, err := al.uploadRequestFromRequest(req)
	if err != nil {
		al.jsonError(w, err)
		return
	}

	curUser, err := al.getUserModelForAuth(req.Context())
	if err != nil {
		al.jsonErrorResponseWithTitle(w, http.StatusBadRequest, err.Error())
		return
	}
	defer uploadRequest.File.Close()

	wasCreated, err := al.filesAPI.CreateDirIfNotExists(al.config.GetUploadDir(), files.DefaultMode)
	if err != nil {
		al.jsonError(w, err)
		return
	}
	if wasCreated {
		al.Infof("created directory %s", al.config.GetUploadDir())
	}

	uploadRequest.SourceFilePath = al.genFilePath(uploadRequest.ID)

	err = uploadRequest.Validate()
	if err != nil {
		al.jsonErrorResponseWithTitle(w, http.StatusBadRequest, err.Error())
		return
	}

	if err = validateRemoteDestination(uploadRequest); err != nil {
		al.jsonErrorResponseWithDetail(w, http.StatusBadRequest, "BAD_DESTINATION", "upload denied", err.Error())
		return
	}

	cr := clients.ClientServiceProvider{}
	clientGroups, err := al.clientGroupProvider.GetAll(req.Context())
	if err != nil {
		al.jsonError(w, err)
	}
	if err := cr.CheckClientsAccess(uploadRequest.Clients, curUser, clientGroups); err != nil {
		al.jsonErrorResponseWithDetail(w, http.StatusForbidden, "ACCESS_CONTROL_VIOLATION", "upload forbidden", err.Error())
		return
	}

	copiedBytes, err := al.filesAPI.CreateFile(uploadRequest.SourceFilePath, uploadRequest.File)
	if err != nil {
		al.jsonError(w, err)
		return
	}

	file, err := al.filesAPI.Open(uploadRequest.SourceFilePath)
	if err != nil {
		al.jsonError(w, err)
		return
	}

	md5Checksum, err := files.Md5HashFromReader(file)
	if err != nil {
		al.jsonError(w, err)
		return
	}

	uploadRequest.Md5Checksum = md5Checksum

	al.Debugf(
		"stored file %s on server, size %d, Content-Type %s, temp location: %s, md5 checksum: %x",
		uploadRequest.FileHeader.Filename,
		uploadRequest.FileHeader.Size,
		uploadRequest.FileHeader.Header.Get("Content-Type"),
		uploadRequest.SourceFilePath,
		md5Checksum,
	)

	uploadRep := &models.UploadResponseShort{
		ID:        uploadRequest.ID,
		Filepath:  uploadRequest.DestinationPath,
		SizeBytes: copiedBytes,
	}
	al.auditLog.Entry(auditlog.ApplicationUploads, auditlog.ActionCreate).
		WithHTTPRequest(req).
		WithRequest(uploadRequest.UploadedFile).
		WithResponse(uploadRep).
		WithID(uploadRequest.UploadedFile.ID).
		SaveForMultipleClients(uploadRequest.Clients)

	go al.sendFileToClients(uploadRequest)

	response := api.NewSuccessPayload(uploadRep)

	al.writeJSONResponse(w, http.StatusOK, response)
}

func (al *APIListener) handleUploadsWS(w http.ResponseWriter, req *http.Request) {
	uiConn, err := apiUpgrader.Upgrade(w, req, nil)
	if err != nil {
		al.Errorf("Failed to establish WS connection: %v", err)
		return
	}

	connID, err := uuid.NewUUID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	al.Server.uploadWebSockets.Store(connID, uiConn)

	defer al.Server.uploadWebSockets.Delete(connID)
	defer uiConn.Close()

	for {
		_, _, err := uiConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				al.Infof("closed ws connection: %v", err)
			}
			break
		}
	}
}

func (al *APIListener) genFilePath(uuid string) string {
	uniqueFilename := fmt.Sprintf("%s_rport_filepush", uuid)

	return filepath.Join(al.config.GetUploadDir(), uniqueFilename)
}

type uploadResult struct {
	resp   *models.UploadResponse
	err    error
	client *clientdata.Client
}

type UploadOutput struct {
	ClientID string `json:"client_id"`
	*models.UploadResponse
}

func (al *APIListener) sendFileToClients(uploadRequest *UploadRequest) {
	wg := &sync.WaitGroup{}
	wg.Add(len(uploadRequest.Clients))

	resChan := make(chan *uploadResult, len(uploadRequest.Clients))

	for _, cl := range uploadRequest.Clients {
		go al.sendFileToClient(wg, uploadRequest.UploadedFile, cl, resChan)
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	al.consumeUploadResults(resChan, uploadRequest)

	err := al.filesAPI.Remove(uploadRequest.SourceFilePath)
	if err != nil {
		al.Errorf("failed to delete temp file path %s: %v", uploadRequest.SourceFilePath, err)
	}
}

func (al *APIListener) consumeUploadResults(resChan chan *uploadResult, uploadRequest *UploadRequest) {
	for res := range resChan {
		clientID := res.client.GetID()
		output := &UploadOutput{
			ClientID:       clientID,
			UploadResponse: res.resp,
		}
		if res.err != nil {
			errTxt := res.err.Error()
			if errTxt == "client error: unknown request" {
				errTxt = "client doens't support uploads, please upgrade client to the latest version to make it work"
			}
			output.UploadResponse = &models.UploadResponse{
				Message: errTxt,
				Status:  "error",
				UploadResponseShort: models.UploadResponseShort{
					ID: uploadRequest.ID,
				},
			}
			al.Errorf(
				"upload failure: %s, file id: %s, file path: %s, client %s",
				errTxt,
				uploadRequest.ID,
				uploadRequest.DestinationPath,
				clientID,
			)
			al.auditLog.Entry(auditlog.ApplicationUploads, auditlog.ActionFailed).
				WithRequest(uploadRequest.UploadedFile).
				WithResponse(output).
				WithID(uploadRequest.UploadedFile.ID).
				WithClient(res.client).
				Save()
		} else {
			al.Infof(
				"upload success, file id: %s, file path: %s, client %s",
				uploadRequest.ID,
				uploadRequest.DestinationPath,
				clientID,
			)
			al.auditLog.Entry(auditlog.ApplicationUploads, auditlog.ActionSuccess).
				WithRequest(uploadRequest.UploadedFile).
				WithResponse(output).
				WithID(uploadRequest.UploadedFile.ID).
				WithClient(res.client).
				Save()
		}

		al.notifyUploadEventListeners(output)
	}
}

func (al *APIListener) sendFileToClient(wg *sync.WaitGroup, file *models.UploadedFile, cl *clientdata.Client, resChan chan *uploadResult) {
	defer wg.Done()

	fileReceptionConfig := cl.GetFileReceptionConfig()
	if fileReceptionConfig != nil && !fileReceptionConfig.Enabled {
		resChan <- &uploadResult{
			err:    errors3.ErrUploadsDisabled,
			client: cl,
			resp:   nil,
		}
		return
	}
	resp := &models.UploadResponse{}
	err := comm.SendRequestAndGetResponse(cl.GetConnection(), comm.RequestTypeUpload, file, resp, al.Log())

	resChan <- &uploadResult{
		err:    err,
		client: cl,
		resp:   resp,
	}
}

func (al *APIListener) notifyUploadEventListeners(msg interface{}) {
	al.uploadWebSockets.Range(func(key, value interface{}) bool {
		if wsConn, ok := value.(*websocket.Conn); ok {
			err := wsConn.WriteJSON(msg)
			if err != nil {
				al.Errorf("failed to send notification to websocket client %s: %v", key, err)
			}
		}
		return true
	})
}

func (al *APIListener) uploadRequestFromRequest(req *http.Request) (ur *UploadRequest, err error) {
	ur = &UploadRequest{
		UploadedFile: &models.UploadedFile{},
	}

	err = req.ParseMultipartForm(uploadBufSize)
	if err != nil {
		return nil, &errors2.APIError{
			Err:        err,
			HTTPStatus: http.StatusBadRequest,
		}
	}

	ur.ClientIDs = req.MultipartForm.Value["client_id"]
	ur.GroupIDs = req.MultipartForm.Value["group_id"]

	clientTags, err := getClientTagsFromReqForm(req)
	if err != nil {
		return nil, &errors2.APIError{
			Err:        err,
			HTTPStatus: http.StatusBadRequest,
		}
	}

	ur.ClientTags = clientTags

	orderedClients, clientsInGroupsCount, err := al.getOrderedClientsWithValidation(req.Context(), ur)
	if err != nil {
		return nil, err
	}

	ur.Clients = orderedClients
	ur.clientsInGroupsCount = clientsInGroupsCount

	err = ur.FromMultipartRequest(req)
	if err != nil {
		return nil, &errors2.APIError{
			Err:        err,
			HTTPStatus: http.StatusBadRequest,
		}
	}

	ur.File, ur.FileHeader, err = req.FormFile("upload")
	if err != nil {
		return nil, &errors2.APIError{
			Err:        err,
			HTTPStatus: http.StatusBadRequest,
		}
	}

	if ur.UploadedFile.ID == "" {
		id, e := random.UUID4()
		if e != nil {
			al.Errorf("failed to generate uuid, will fallback to timestamp uuid, error: %v", e)
			id = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		ur.UploadedFile.ID = id
	}

	return ur, nil
}

func getClientTagsFromReqForm(req *http.Request) (clientTags *models.JobClientTags, err error) {
	jsonTags := req.MultipartForm.Value["tags"]

	if len(jsonTags) == 0 {
		return nil, nil
	}

	if len(jsonTags) > 1 {
		return nil, errors.New("tags form value must only contain a single element")
	}

	clientTags = &models.JobClientTags{}
	err = json.Unmarshal([]byte(jsonTags[0]), clientTags)
	if err != nil {
		return nil, err
	}

	return clientTags, nil
}

func validateRemoteDestination(ur *UploadRequest) error {
	// deny uploads to the below unix folders because there is no reason why a user should do that.
	denied := []string{"/proc/", "/sys/", "/dev/", "/run/"}
	for _, v := range denied {
		if strings.HasPrefix(ur.DestinationPath, v) {
			return fmt.Errorf("uploads to %s are forbidden", v)
		}
	}
	return nil
}
