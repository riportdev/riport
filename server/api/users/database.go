package users

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	errors2 "github.com/riportdev/riport/server/api/errors"
	"github.com/riportdev/riport/share/enums"
	"github.com/riportdev/riport/share/logger"

	"github.com/jmoiron/sqlx"
)

type UserDatabase struct {
	db *sqlx.DB

	usersTableName        string
	groupsTableName       string
	groupDetailsTableName string

	twoFAOn     bool
	totPOn      bool
	plusEnabled bool
	logger      *logger.Logger
}

func NewUserDatabase(
	DB *sqlx.DB,
	usersTableName, groupsTableName, groupDetailsTableName string,
	twoFAOn, totPOn bool,
	plusEnabled bool, logger *logger.Logger,
) (*UserDatabase, error) {
	d := &UserDatabase{
		db: DB,

		usersTableName:        usersTableName,
		groupsTableName:       groupsTableName,
		groupDetailsTableName: groupDetailsTableName,

		twoFAOn:     twoFAOn,
		totPOn:      totPOn,
		plusEnabled: plusEnabled,
		logger:      logger,
	}
	if err := d.checkDatabaseTables(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *UserDatabase) getSelectClause() string {
	s := "username, password, password_expired"
	if d.twoFAOn {
		s += ", two_fa_send_to"
	}
	if d.totPOn {
		s += ", totp_secret"
	}
	return s
}

// checkDatabaseTables @todo use context for all db operations
func (d *UserDatabase) checkDatabaseTables() error {
	_, err := d.db.Exec(fmt.Sprintf("SELECT %s FROM `%s` LIMIT 0", d.getSelectClause(), d.usersTableName))
	if err != nil {
		err = fmt.Errorf("%v, if you have 2fa enabled please check additional column requirements at https://oss.riport.io/docs/no02-api-auth.html#database", err)
		return err
	}
	_, err = d.db.Exec(fmt.Sprintf("SELECT username, `group` FROM `%s` LIMIT 0", d.groupsTableName))
	if err != nil {
		return err
	}
	if d.groupDetailsTableName != "" {
		extPermSelect := ""
		if d.plusEnabled {
			// when plus is enabled, we need to select the other fields as well
			extPermSelect = ", tunnels_restricted, commands_restricted"
		}
		_, err = d.db.Exec(fmt.Sprintf("SELECT name, permissions %s FROM `%s` LIMIT 0", extPermSelect, d.groupDetailsTableName))
		if err != nil {
			return err
		}
	}

	return nil
}

// GetByUsername @todo use context for all db operations
func (d *UserDatabase) GetByUsername(username string) (*User, error) {
	user := &User{}
	err := d.db.Get(user, fmt.Sprintf("SELECT %s FROM `%s` WHERE username = ? LIMIT 1", d.getSelectClause(), d.usersTableName), username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	err = d.db.Select(&user.Groups, fmt.Sprintf("SELECT DISTINCT(`group`) FROM `%s` WHERE username = ?", d.groupsTableName), username)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return user, nil
}

// GetAll @todo use context for all db operations
func (d *UserDatabase) GetAll() ([]*User, error) {
	var usrs []*User
	err := d.db.Select(&usrs, fmt.Sprintf("SELECT %s FROM `%s` ORDER BY username", d.getSelectClause(), d.usersTableName))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var groups []struct {
		Username string `db:"username"`
		Group    string `db:"group"`
	}
	err = d.db.Select(&groups, fmt.Sprintf("SELECT `username`, `group` FROM `%s` ORDER BY `group`", d.groupsTableName))
	if err != nil {
		if err == sql.ErrNoRows {
			return usrs, nil
		}
		return nil, err
	}
	for i := range groups {
		for y := range usrs {
			if usrs[y].Username == groups[i].Username {
				usrs[y].Groups = append(usrs[y].Groups, groups[i].Group)
			}
		}
	}

	return usrs, nil
}

func (d *UserDatabase) ListGroups() ([]Group, error) {
	var groups []Group

	if d.groupDetailsTableName != "" {
		err := d.db.Select(&groups, fmt.Sprintf("SELECT * FROM `%s` ORDER BY `name`", d.groupDetailsTableName))
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
	}

	var userGroups []string
	err := d.db.Select(&userGroups, fmt.Sprintf("SELECT DISTINCT `group` FROM `%s` ORDER BY `group`", d.groupsTableName))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for _, ug := range userGroups {
		found := false
		for _, g := range groups {
			if ug == g.Name {
				found = true
				break
			}
		}
		if !found {
			groups = append(groups, NewGroup(ug, nil, nil))
		}
	}

	return groups, nil
}

func (d *UserDatabase) GetGroup(name string) (Group, error) {
	if d.groupDetailsTableName == "" {
		return NewGroup(name, nil, nil), nil
	}

	group := Group{}
	err := d.db.Get(&group, fmt.Sprintf("SELECT * FROM `%s` WHERE name = ? LIMIT 1", d.groupDetailsTableName), name)
	if err == sql.ErrNoRows {
		return NewGroup(name, nil, nil), nil
	} else if err != nil {
		return Group{}, err
	}

	return group, nil
}

func (d *UserDatabase) UpdateGroup(name string, group Group) error {
	if d.groupDetailsTableName == "" {
		return errors2.APIError{
			Message:    "User group details table must be configured for this operation.",
			HTTPStatus: http.StatusBadRequest,
		}
	}

	// We rely on a unique index. Let the database decide, if INSERT or UPDATE is needed.
	var err error
	group.Name = name

	qt1 := ""
	qt2 := ""
	if group.TunnelsRestricted != nil {
		qt1 = ", tunnels_restricted"
		qt2 = ", :tunnels_restricted"
	}

	qc1 := ""
	qc2 := ""
	if group.CommandsRestricted != nil {
		qc1 = ", commands_restricted"
		qc2 = ", :commands_restricted"
	}
	// compose the query (assume the extended fields are present)
	// We rely on a unique index. Let the database decide, if INSERT or UPDATE is needed.
	qb := fmt.Sprintf("REPLACE INTO `%s` (name, permissions%s%s) VALUES (:name, :permissions%s%s)", d.groupDetailsTableName, qt1, qc1, qt2, qc2)

	_, err = d.db.NamedExec(qb, group)

	if err != nil {
		if d.plusEnabled {
			return err
		}
		_, err = d.db.NamedExec(qb+"(name, permissions) VALUES (:name, :permissions)", group) // ignore the extended fields
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *UserDatabase) DeleteGroup(name string) error {
	if d.groupDetailsTableName == "" {
		return errors2.APIError{
			Message:    "User group details table must be configured for this operation.",
			HTTPStatus: http.StatusBadRequest,
		}
	}

	tx, err := d.db.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("DELETE FROM `%s` WHERE `group` = ?", d.groupsTableName), name)
	if err != nil {
		d.handleRollback(tx)
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("DELETE FROM `%s` WHERE `name` = ?", d.groupDetailsTableName), name)
	if err != nil {
		d.handleRollback(tx)
		return err
	}

	return tx.Commit()
}

func (d *UserDatabase) handleRollback(tx *sqlx.Tx) {
	err := tx.Rollback()
	if err != nil {
		d.logger.Errorf("Failed to rollback transaction: %v", err)
	}
}

// Add todo use context for all db operations
func (d *UserDatabase) Add(usr *User) error {
	tx, err := d.db.Beginx()
	if err != nil {
		return err
	}

	columns := []string{
		"`username`",
		"`password`",
	}
	params := []interface{}{
		usr.Username,
		usr.Password,
	}

	if d.twoFAOn {
		columns = append(columns, "`two_fa_send_to`")
		params = append(params, usr.TwoFASendTo)
	}

	if d.totPOn {
		columns = append(columns, "`totp_secret`")
		params = append(params, usr.TotP)
	}

	_, err = tx.Exec(
		fmt.Sprintf(
			"INSERT INTO `%s` (%s) VALUES (%s)",
			d.usersTableName,
			strings.Join(columns, ", "),
			strings.TrimRight(strings.Repeat("?,", len(params)), ","),
		),
		params...,
	)

	if err != nil {
		d.handleRollback(tx)
		return err
	}

	for i := range usr.Groups {
		_, err := tx.Exec(
			fmt.Sprintf("INSERT INTO `%s` (`username`, `group`) VALUES (?, ?)", d.groupsTableName),
			usr.Username,
			usr.Groups[i],
		)
		if err != nil {
			d.handleRollback(tx)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Update @todo use context for all db operations
func (d *UserDatabase) Update(usr *User, usernameToUpdate string) error {
	if usernameToUpdate == "" {
		return errors.New("cannot update user with empty username")
	}

	params := []interface{}{}
	statements := []string{}
	if usr.Password != "" {
		statements = append(statements, "`password` = ?")
		params = append(params, usr.Password)
	}

	if usr.PasswordExpired != nil {
		statements = append(statements, "`password_expired` = ?")
		params = append(params, usr.PasswordExpired)
	}

	if usr.TwoFASendTo != "" {
		statements = append(statements, "`two_fa_send_to` = ?")
		params = append(params, usr.TwoFASendTo)
	}

	if usr.TotP != "" {
		statements = append(statements, "`totp_secret` = ?")
		params = append(params, usr.TotP)
	}

	if usr.Username != "" && usr.Username != usernameToUpdate {
		statements = append(statements, "`username` = ?")
		params = append(params, usr.Username)
	}

	tx, err := d.db.Beginx()
	if err != nil {
		return err
	}

	if len(params) > 0 {
		q := fmt.Sprintf(
			"UPDATE `%s` SET %s WHERE username = ?",
			d.usersTableName,
			strings.Join(statements, ", "),
		)
		params = append(params, usernameToUpdate)
		_, err := tx.Exec(q, params...)
		if err != nil {
			d.handleRollback(tx)
			return err
		}
	}

	if usr.Username != "" && usernameToUpdate != usr.Username {
		_, err := tx.Exec(
			fmt.Sprintf("UPDATE `%s` SET `username` = ? WHERE `username` = ?", d.groupsTableName),
			usr.Username,
			usernameToUpdate,
		)
		if err != nil {
			d.handleRollback(tx)
			return err
		}
	}

	groupUserName := usernameToUpdate
	if usr.Username != "" {
		groupUserName = usr.Username
	}

	if usr.Groups != nil {
		_, err := tx.Exec(fmt.Sprintf("DELETE FROM `%s` WHERE `username` = ?", d.groupsTableName), groupUserName)
		if err != nil {
			d.handleRollback(tx)
			return err
		}
	}

	for i := range usr.Groups {
		group := usr.Groups[i]
		_, err := tx.Exec(
			fmt.Sprintf("INSERT INTO `%s` (`username`, `group`) VALUES (?, ?)", d.groupsTableName),
			groupUserName,
			group,
		)
		if err != nil {
			d.handleRollback(tx)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Delete @todo use context for all db operations
func (d *UserDatabase) Delete(usernameToDelete string) error {
	tx, err := d.db.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("DELETE FROM `%s` WHERE `username` = ?", d.usersTableName), usernameToDelete)
	if err != nil {
		d.handleRollback(tx)
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("DELETE FROM `%s` WHERE `username` = ?", d.groupsTableName), usernameToDelete)
	if err != nil {
		d.handleRollback(tx)
		return err
	}

	return tx.Commit()
}

func (d UserDatabase) Type() enums.ProviderSource {
	return enums.ProviderSourceDB
}
func (d UserDatabase) SupportsGroupPermissions() bool {
	return d.groupDetailsTableName != ""
}
