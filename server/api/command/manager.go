package command

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/riportdev/riport/share/query"
	"github.com/riportdev/riport/share/types"

	errors2 "github.com/riportdev/riport/server/api/errors"
)

var (
	supportedSortAndFilters = map[string]bool{
		"id":         true,
		"name":       true,
		"created_by": true,
		"created_at": true,
		"updated_by": true,
		"updated_at": true,
		"cmd":        true,
		"tags":       true,
	}
	supportedFields = map[string]map[string]bool{
		"commands": {
			"id":         true,
			"name":       true,
			"created_by": true,
			"created_at": true,
			"updated_by": true,
			"updated_at": true,
			"cmd":        true,
			"tags":       true,
		},
	}
	manualFiltersConfig = map[string]bool{
		"tags": true,
	}
)

type DbProvider interface {
	GetByID(ctx context.Context, id string, ro *query.RetrieveOptions) (val *Command, found bool, err error)
	List(ctx context.Context, lo *query.ListOptions) ([]Command, error)
	Save(ctx context.Context, s *Command) (string, error)
	Delete(ctx context.Context, id string) error
	io.Closer
}

type Manager struct {
	db DbProvider
}

func NewManager(db DbProvider) *Manager {
	return &Manager{
		db: db,
	}
}

func (m *Manager) List(ctx context.Context, re *http.Request) ([]Command, int, error) {
	listOptions := query.GetListOptions(re)

	err := query.ValidateListOptions(listOptions, supportedSortAndFilters, supportedSortAndFilters, supportedFields, &query.PaginationConfig{
		MaxLimit:     100,
		DefaultLimit: 20,
	})
	if err != nil {
		return nil, 0, err
	}

	manualFilters, dbFilters := query.SplitFilters(listOptions.Filters, manualFiltersConfig)
	pagination := listOptions.Pagination

	listOptions.Filters = dbFilters
	listOptions.Pagination = nil

	entries, err := m.db.List(ctx, listOptions)
	if err != nil {
		return nil, 0, err
	}

	filtered := make([]Command, 0, len(entries))
	for _, entry := range entries {
		matches, err := query.MatchesFilters(entry, manualFilters)
		if err != nil {
			return nil, 0, err
		}
		if matches {
			filtered = append(filtered, entry)
		}
	}

	totalCount := len(filtered)
	start, end := pagination.GetStartEnd(totalCount)
	limited := filtered[start:end]

	return limited, totalCount, nil
}

func (m *Manager) GetOne(ctx context.Context, re *http.Request, id string) (*Command, bool, error) {
	retrieveOptions := query.GetRetrieveOptions(re)

	err := query.ValidateRetrieveOptions(retrieveOptions, supportedFields)
	if err != nil {
		return nil, false, err
	}

	val, found, err := m.db.GetByID(ctx, id, retrieveOptions)
	if err != nil {
		return nil, false, err
	}

	if !found {
		return nil, false, nil
	}

	return val, true, nil
}

func (m *Manager) Create(ctx context.Context, valueToStore *InputCommand, username string) (*Command, error) {
	err := Validate(valueToStore)
	if err != nil {
		return nil, err
	}

	existingCommand, err := m.db.List(ctx, &query.ListOptions{
		Filters: []query.FilterOption{
			{
				Column: []string{"name"},
				Values: []string{valueToStore.Name},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(existingCommand) > 0 {
		return nil, errors2.APIError{
			Message:    fmt.Sprintf("another command with the same name '%s' exists", valueToStore.Name),
			HTTPStatus: http.StatusConflict,
		}
	}
	if valueToStore.TimoutSec == 0 {
		valueToStore.TimoutSec = DefaultTimeoutSec
	}

	now := time.Now()
	commandToSave := &Command{
		Name:      valueToStore.Name,
		CreatedBy: username,
		CreatedAt: &now,
		UpdatedBy: username,
		UpdatedAt: &now,
		Cmd:       valueToStore.Cmd,
		Tags:      (*types.StringSlice)(&valueToStore.Tags),
		TimoutSec: &valueToStore.TimoutSec,
	}
	commandToSave.ID, err = m.db.Save(ctx, commandToSave)
	if err != nil {
		return nil, err
	}

	return commandToSave, nil
}

func (m *Manager) Update(ctx context.Context, existingID string, valueToStore *InputCommand, username string) (*Command, error) {
	err := Validate(valueToStore)
	if err != nil {
		return nil, err
	}

	existing, foundByID, err := m.db.GetByID(ctx, existingID, &query.RetrieveOptions{})
	if err != nil {
		return nil, err
	}

	if !foundByID {
		return nil, errors2.APIError{
			Message:    "cannot find entry by the provided ID",
			HTTPStatus: http.StatusNotFound,
		}
	}

	commandsWithSameName, err := m.db.List(ctx, &query.ListOptions{
		Filters: []query.FilterOption{
			{
				Column: []string{"name"},
				Values: []string{valueToStore.Name},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(commandsWithSameName) > 0 && commandsWithSameName[0].ID != existingID {
		return nil, errors2.APIError{
			Message:    fmt.Sprintf("another command with the same name '%s' exists", valueToStore.Name),
			HTTPStatus: http.StatusConflict,
		}
	}
	if valueToStore.TimoutSec == 0 {
		valueToStore.TimoutSec = DefaultTimeoutSec
	}

	now := time.Now()
	commandToSave := &Command{
		ID:        existingID,
		Name:      valueToStore.Name,
		CreatedBy: existing.CreatedBy,
		CreatedAt: existing.CreatedAt,
		UpdatedBy: username,
		UpdatedAt: &now,
		Cmd:       valueToStore.Cmd,
		Tags:      (*types.StringSlice)(&valueToStore.Tags),
		TimoutSec: &valueToStore.TimoutSec,
	}
	_, err = m.db.Save(ctx, commandToSave)
	if err != nil {
		return nil, err
	}

	return commandToSave, nil
}

func (m *Manager) Delete(ctx context.Context, id string) error {
	_, found, err := m.db.GetByID(ctx, id, &query.RetrieveOptions{})
	if err != nil {
		return errors2.APIError{
			Err:        err,
			HTTPStatus: http.StatusInternalServerError,
		}
	}

	if !found {
		return errors2.APIError{
			Message:    "cannot find this entry by the provided id",
			HTTPStatus: http.StatusNotFound,
		}
	}

	err = m.db.Delete(ctx, id)
	if err != nil {
		return errors2.APIError{
			Err:        err,
			HTTPStatus: http.StatusInternalServerError,
		}
	}

	return nil
}

func (m *Manager) Close() error {
	return m.db.Close()
}
