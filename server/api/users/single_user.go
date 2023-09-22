package users

import (
	"net/http"

	errors2 "github.com/riportdev/riport/server/api/errors"
	"github.com/riportdev/riport/share/enums"
)

type StaticProvider struct {
	*UserCache
}

func NewStaticProvider(users []*User) *StaticProvider {
	return &StaticProvider{
		UserCache: NewUserCache(users),
	}
}

func (p StaticProvider) Type() enums.ProviderSource {
	return enums.ProviderSourceStatic
}

func (p StaticProvider) SupportsGroupPermissions() bool {
	return false
}

func (p *StaticProvider) ListGroups() ([]Group, error) {
	return nil, errors2.APIError{
		Message:    "The single user authentication doesn't support this feature.",
		HTTPStatus: http.StatusBadRequest,
	}
}

func (p *StaticProvider) GetGroup(string) (Group, error) {
	return Group{}, errors2.APIError{
		Message:    "The single user authentication doesn't support this feature.",
		HTTPStatus: http.StatusBadRequest,
	}
}

func (p *StaticProvider) UpdateGroup(string, Group) error {
	return errors2.APIError{
		Message:    "The single user authentication doesn't support this feature.",
		HTTPStatus: http.StatusBadRequest,
	}
}

func (p *StaticProvider) DeleteGroup(string) error {
	return errors2.APIError{
		Message:    "The single user authentication doesn't support this feature.",
		HTTPStatus: http.StatusBadRequest,
	}
}

func (p *StaticProvider) Add(usr *User) error {
	return errors2.APIError{
		Message:    "The single user authentication doesn't support this operation.",
		HTTPStatus: http.StatusBadRequest,
	}
}
func (p *StaticProvider) Update(usr *User, username string) error {
	return errors2.APIError{
		Message:    "The single user authentication doesn't support this operation.",
		HTTPStatus: http.StatusBadRequest,
	}
}
func (p *StaticProvider) Delete(username string) error {
	return errors2.APIError{
		Message:    "The single user authentication doesn't support this operation.",
		HTTPStatus: http.StatusBadRequest,
	}
}
