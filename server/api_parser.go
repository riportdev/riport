package chserver

import (
	"encoding/json"
	"io"
	"net/http"

	errors2 "github.com/riportdev/riport/server/api/errors"
)

func parseRequestBody(reqBody io.ReadCloser, dest interface{}) error {
	dec := json.NewDecoder(reqBody)
	dec.DisallowUnknownFields()
	err := dec.Decode(dest)
	if err == io.EOF { // is handled separately to return an informative error message
		return errors2.APIError{
			Message:    "Missing body with json data.",
			HTTPStatus: http.StatusBadRequest,
		}
	}

	if err != nil {
		return errors2.APIError{
			Message:    "Invalid JSON data.",
			Err:        err,
			HTTPStatus: http.StatusBadRequest,
		}
	}

	return nil
}
