package chserver

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/riportdev/riport/server/api"
	"github.com/riportdev/riport/server/routes"
	"github.com/riportdev/riport/share/query"
)

var (
	supportedFilters = map[string]bool{
		"state":            true,
		"reference_id":     true,
		"transport":        true,
		"subject":          true,
		"timestamp[gt]":    true,
		"timestamp[lt]":    true,
		"timestamp[since]": true,
		"timestamp[until]": true,
	}
	supportedSorts = map[string]bool{
		"timestamp": true,
		"state":     true,
	}
)

func (al *APIListener) notificationsList(ctx context.Context, options *query.ListOptions) (*api.SuccessPayload, error) {

	err := query.ValidateListOptions(options, supportedSorts, supportedFilters, nil, &query.PaginationConfig{
		DefaultLimit: 10,
		MaxLimit:     100,
	})
	if err != nil {
		return nil, err
	}

	entries, err := al.notificationsStorage.List(ctx, options)
	if err != nil {
		return nil, err
	}

	count, err := al.notificationsStorage.Count(ctx, options)
	if err != nil {
		return nil, err
	}

	return &api.SuccessPayload{
		Data: entries,
		Meta: api.NewMeta(count),
	}, nil
}

func (al *APIListener) handleGetNotifications(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	options := query.GetListOptions(request)
	result, err := al.notificationsList(ctx, options)
	if err != nil {
		al.jsonError(writer, err)
		return
	}

	al.writeJSONResponse(writer, http.StatusOK, result)
}

func (al *APIListener) handleGetNotificationDetails(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	vars := mux.Vars(request)
	nid := vars[routes.ParamNotificationID]

	notification, found, err := al.notificationsStorage.Details(ctx, nid)
	if err != nil {
		al.jsonError(writer, err)
		return
	}

	if !found {
		al.writeJSONResponse(writer, http.StatusNotFound, nil)
		return
	}

	al.writeJSONResponse(writer, http.StatusOK, notification)
}
