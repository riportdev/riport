package auditlog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/riportdev/riport/server/api"
	"github.com/riportdev/riport/server/clients/clientdata"
	chshare "github.com/riportdev/riport/share"
)

type Entry struct {
	Timestamp      time.Time `db:"timestamp" json:"timestamp"`
	Username       string    `db:"username" json:"username"`
	RemoteIP       string    `db:"remote_ip" json:"remote_ip"`
	Application    string    `db:"application" json:"application"`
	Action         string    `db:"action" json:"action"`
	ID             string    `db:"affected_id" json:"affected_id"`
	ClientID       string    `db:"client_id" json:"client_id"`
	ClientHostName string    `db:"client_hostname" json:"client_hostname"`
	Request        string    `db:"request" json:"request"`
	Response       string    `db:"response" json:"response"`

	al *AuditLog
}

func (e *Entry) WithID(id interface{}) *Entry {
	if e == nil {
		return e
	}

	e.ID = fmt.Sprint(id)
	return e
}

func (e *Entry) WithHTTPRequest(req *http.Request) *Entry {
	if e == nil {
		return e
	}

	e.Username = api.GetUser(req.Context(), e.al.logger)
	e.RemoteIP = chshare.RemoteIP(req)

	return e
}

func (e *Entry) WithRequest(request interface{}) *Entry {
	if e == nil {
		return e
	}

	reqJSON, err := json.Marshal(request)
	if err != nil {
		e.al.logger.Errorf("Could not marshal auditlog request: %v", err)
		return e
	}
	e.Request = string(reqJSON)

	return e
}

func (e *Entry) WithResponse(response interface{}) *Entry {
	if e == nil {
		return e
	}

	respJSON, err := json.Marshal(response)
	if err != nil {
		e.al.logger.Errorf("Could not marshal auditlog response: %v", err)
		return e
	}
	e.Response = string(respJSON)

	return e
}

func (e *Entry) WithClient(c *clientdata.Client) *Entry {
	if e == nil {
		return e
	}

	e.ClientID = c.GetID()
	e.ClientHostName = c.GetHostname()
	return e
}

func (e *Entry) WithClientID(cid string) *Entry {
	if e == nil {
		return e
	}

	e.ClientID = cid

	client, err := e.al.clientGetter.GetByID(cid)
	if err != nil {
		e.al.logger.Errorf("Could not get client for auditlog: %v", err)
		return e
	}
	if client != nil {
		e.ClientHostName = client.GetHostname()
	}

	return e
}

func (e *Entry) Save() {
	if e == nil {
		return
	}

	err := e.al.savePreparedEntry(e)
	if err != nil {
		e.al.logger.Errorf("Could not save auditlog entry: %v", err)
	}
}

func (e *Entry) SaveForMultipleClients(clients []*clientdata.Client) {
	if e == nil {
		return
	}

	for _, c := range clients {
		e.WithClient(c).Save()
	}
}
