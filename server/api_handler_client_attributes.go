package chserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/riportdev/riport/server/api"
	"github.com/riportdev/riport/server/clients/clientdata"
	"github.com/riportdev/riport/server/routes"
	"github.com/riportdev/riport/share/comm"
	"github.com/riportdev/riport/share/models"
)

type clientCtxKeyType int

const clientCtxKey clientCtxKeyType = iota

func (al *APIListener) getClientFromContext(ctx context.Context) (*clientdata.Client, error) {
	maybeClient := ctx.Value(clientCtxKey)
	if maybeClient == nil {
		return nil, fmt.Errorf("client not present in the request")
	}
	client, ok := maybeClient.(*clientdata.Client)
	if !ok {
		return nil, fmt.Errorf("client is not of the client type")
	}
	return client, nil
}

func (al *APIListener) handleGetClientAttributes(w http.ResponseWriter, req *http.Request) {

	ctx := req.Context()

	client, err := al.getClientFromContext(ctx)
	if err != nil {
		al.jsonErrorResponseWithTitle(w, http.StatusInternalServerError, "client not present in the request")
	}

	al.writeJSONResponse(w, http.StatusOK, api.NewSuccessPayload(client.GetAttributes()))
}

type Resp struct {
	OK string `json:"ok"`
}

func (al *APIListener) handleUpdateClientAttributes(w http.ResponseWriter, req *http.Request) {

	ctx := req.Context()

	client, err := al.getClientFromContext(ctx)
	if err != nil {
		al.jsonErrorResponseWithTitle(w, http.StatusInternalServerError, "client not present in the request")
	}

	attributesRaw, err := io.ReadAll(req.Body)
	if err != nil {
		al.jsonErrorResponseWithTitle(w, http.StatusBadRequest, fmt.Sprintf("failed reading request: %v", err))
		return
	}

	attributes := models.Attributes{}
	err = json.Unmarshal(attributesRaw, &attributes)
	if err != nil {
		al.jsonErrorResponseWithTitle(w, http.StatusBadRequest, fmt.Sprintf("failed parsing attributes: %v", err))
		return
	}

	sshResp := &Resp{}
	err = comm.SendRequestAndGetResponse(client.GetConnection(), comm.RequestTypeUpdateClientAttributes, attributes, sshResp, al.Log())
	if err != nil {
		if _, ok := err.(*comm.ClientError); ok {
			al.jsonErrorResponseWithTitle(w, http.StatusConflict, err.Error())
		} else {
			al.jsonErrorResponseWithError(w, http.StatusInternalServerError, "Failed to execute remote command.", err)
		}
		return
	}

	client.SetAttributes(attributes)

	err = al.clientService.GetRepo().Save(client)
	if err != nil {
		al.writeJSONResponse(w, http.StatusOK, api.NewSuccessPayload("client attributes updated, error saving changes to local db, changes will be visible after next client connection"))
	}

	al.writeJSONResponse(w, http.StatusOK, api.NewSuccessPayload("ok"))
}

func (al *APIListener) withActiveClient(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		vars := mux.Vars(request)
		cid := vars[routes.ParamClientID]
		if cid == "" {
			al.jsonErrorResponseWithTitle(writer, http.StatusBadRequest, fmt.Sprintf("Missing %q route param.", routes.ParamClientID))
			return
		}

		client, err := al.clientService.GetActiveByID(cid)
		if err != nil {
			al.jsonErrorResponseWithError(writer, http.StatusInternalServerError, fmt.Sprintf("Failed to find an active client with id=%q.", cid), err)
			return
		}
		if client == nil {
			al.jsonErrorResponseWithTitle(writer, http.StatusNotFound, fmt.Sprintf("Active client with id=%q not found.", cid))
			return
		}

		if client.IsPaused() {
			al.jsonErrorResponseWithTitle(writer, http.StatusNotFound, fmt.Sprintf("failed to execute command/script for client with id %s due to client being paused (reason = %s)", client.GetID(), client.GetPausedReason()))
			return
		}

		next.ServeHTTP(writer, request.WithContext(context.WithValue(request.Context(), clientCtxKey, client)))
	})
}
