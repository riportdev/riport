package script

import (
	"net/http"

	errors2 "github.com/riportdev/riport/server/api/errors"
)

func Validate(iv *InputScript) error {
	errs := errors2.APIErrors{}

	if iv.Name == "" {
		errs = append(errs, errors2.APIError{
			Message:    "name is required",
			HTTPStatus: http.StatusBadRequest,
		})
	}
	if iv.Script == "" {
		errs = append(errs, errors2.APIError{
			Message:    "script is required",
			HTTPStatus: http.StatusBadRequest,
		})
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}
