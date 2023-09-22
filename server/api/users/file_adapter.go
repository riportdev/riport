package users

import (
	"fmt"
	"net/http"
	"sync"

	errors2 "github.com/riportdev/riport/server/api/errors"
	"github.com/riportdev/riport/share/enums"
	"github.com/riportdev/riport/share/logger"
)

type FileProvider interface {
	ReadUsersFromFile() ([]*User, error)
	SaveUsersToFile(users []*User) error
}

type FileAdapter struct {
	*UserCache
	*logger.Logger

	mtx          sync.Mutex
	FileProvider FileProvider
}

func NewFileAdapter(logger *logger.Logger, fileProvider FileProvider) (*FileAdapter, error) {
	fa := &FileAdapter{
		UserCache:    NewUserCache(nil),
		Logger:       logger,
		FileProvider: fileProvider,
	}
	if err := fa.load(); err != nil {
		return nil, err
	}
	go fa.reload()
	return fa, nil
}

// load reads users from FileProvider and saves them in cache. It's called from New, on every write and when reload signal is received.
func (fa *FileAdapter) load() error {
	users, err := fa.FileProvider.ReadUsersFromFile()
	if err != nil {
		return err
	}
	fa.Infof("Loaded %v users from file.", len(users))
	fa.UserCache.Load(users)
	return nil
}

func (fa *FileAdapter) ListGroups() ([]Group, error) {
	return nil, errors2.APIError{
		Message:    "The json file authentication backend doesn't support this feature.",
		HTTPStatus: http.StatusBadRequest,
	}
}

func (fa *FileAdapter) GetGroup(string) (Group, error) {
	return Group{}, errors2.APIError{
		Message:    "The json file authentication backend doesn't support this feature.",
		HTTPStatus: http.StatusBadRequest,
	}
}

func (fa *FileAdapter) UpdateGroup(string, Group) error {
	return errors2.APIError{
		Message:    "The json file authentication backend doesn't support this feature.",
		HTTPStatus: http.StatusBadRequest,
	}
}

func (fa *FileAdapter) DeleteGroup(string) error {
	return errors2.APIError{
		Message:    "The json file authentication backend doesn't support this feature.",
		HTTPStatus: http.StatusBadRequest,
	}
}

func (fa *FileAdapter) Delete(usernameToDelete string) error {
	fa.mtx.Lock()
	defer fa.mtx.Unlock()

	usersFromFile, err := fa.FileProvider.ReadUsersFromFile()
	if err != nil {
		return err
	}
	foundIndex := -1
	for i := range usersFromFile {
		if usersFromFile[i].Username == usernameToDelete {
			foundIndex = i
			break
		}
	}

	if foundIndex < 0 {
		return errors2.APIError{
			Message:    fmt.Sprintf("cannot find user by username '%s'", usernameToDelete),
			HTTPStatus: http.StatusNotFound,
		}
	}

	usersToWriteToFile := append(usersFromFile[:foundIndex], usersFromFile[foundIndex+1:]...)
	err = fa.FileProvider.SaveUsersToFile(usersToWriteToFile)
	if err != nil {
		return err
	}

	return fa.load()
}

func (fa *FileAdapter) Update(dataToChange *User, usernameToFind string) error {
	fa.mtx.Lock()
	defer fa.mtx.Unlock()

	users, err := fa.FileProvider.ReadUsersFromFile()
	if err != nil {
		return err
	}

	userFound := -1
	for i := range users {
		if users[i].Username == usernameToFind {
			userFound = i
		}
		if dataToChange.Username != "" && users[i].Username == dataToChange.Username && dataToChange.Username != usernameToFind {
			return errors2.APIError{
				Message:    "Another user with this username already exists",
				HTTPStatus: http.StatusBadRequest,
			}
		}
	}

	if userFound < 0 {
		return errors2.APIError{
			Message:    fmt.Sprintf("cannot find user by username '%s'", usernameToFind),
			HTTPStatus: http.StatusNotFound,
		}
	}

	if dataToChange.Password != "" {
		users[userFound].Password = dataToChange.Password
		users[userFound].PasswordExpired = PasswordExpired(false)
	}
	if dataToChange.PasswordExpired != nil {
		users[userFound].PasswordExpired = dataToChange.PasswordExpired
	}
	if dataToChange.Groups != nil {
		users[userFound].Groups = dataToChange.Groups
	}
	if dataToChange.Username != "" {
		users[userFound].Username = dataToChange.Username
	}
	if dataToChange.TotP != "" {
		users[userFound].TotP = dataToChange.TotP
	}

	err = fa.FileProvider.SaveUsersToFile(users)
	if err != nil {
		return err
	}

	return fa.load()
}

func (fa *FileAdapter) Add(dataToChange *User) error {
	fa.mtx.Lock()
	defer fa.mtx.Unlock()

	users, err := fa.FileProvider.ReadUsersFromFile()
	if err != nil {
		return err
	}

	for i := range users {
		if users[i].Username == dataToChange.Username {
			return errors2.APIError{
				Message:    "Another user with this username already exists",
				HTTPStatus: http.StatusBadRequest,
			}
		}
	}

	users = append(users, dataToChange)
	err = fa.FileProvider.SaveUsersToFile(users)
	if err != nil {
		return err
	}

	return fa.load()
}

func (fa *FileAdapter) Type() enums.ProviderSource {
	return enums.ProviderSourceFile
}

func (fa *FileAdapter) SupportsGroupPermissions() bool {
	return false
}
