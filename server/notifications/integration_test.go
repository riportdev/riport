package notifications_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/suite"

	"github.com/riportdev/riport/db/sqlite"
	"github.com/riportdev/riport/server/notifications"
	"github.com/riportdev/riport/server/notifications/channels/rmailer"
	"github.com/riportdev/riport/server/notifications/channels/scriptRunner"
	notificationsrepo "github.com/riportdev/riport/server/notifications/repository/sqlite"
	"github.com/riportdev/riport/share/logger"
	"github.com/riportdev/riport/share/simpleops"
)

var testLog = logger.NewLogger("client", logger.LogOutput{File: os.Stdout}, logger.LogLevelDebug)

type NotificationsIntegrationTestSuite struct {
	suite.Suite
	dispatcher     notifications.Dispatcher
	store          notificationsrepo.Repository
	server         *smtpmock.Server
	runner         notifications.Processor
	mailConsumer   notifications.Consumer
	scriptConsumer notifications.Consumer
}

func (suite *NotificationsIntegrationTestSuite) SetupTest() {
	db, err := sqlite.New(":memory:", notificationsrepo.AssetNames(), notificationsrepo.Asset, sqlite.DataSourceOptions{})
	suite.NoError(err)
	suite.store = notificationsrepo.NewRepository(db, testLog)
	suite.dispatcher = notifications.NewDispatcher(suite.store)
	suite.server = smtpmock.New(smtpmock.ConfigurationAttr{
		//LogToStdout:              true, // for debugging (especially connection)
		//LogServerActivity:        true, // for debugging (especially connection)
		MultipleMessageReceiving: true,
		// PortNumber:               33334, // randomly generated
	})

	if err := suite.server.Start(); err != nil {
		fmt.Println(err)
	}

	suite.mailConsumer = rmailer.NewConsumer(rmailer.NewRMailer(rmailer.Config{
		Host:     "localhost",
		Port:     suite.server.PortNumber(),
		Domain:   "example.com",
		From:     "test@example.com",
		TLS:      false,
		AuthType: rmailer.AuthTypeNone,
		NoNoop:   true,
	}, testLog), testLog)

	dir, err := os.Getwd()
	suite.NoError(err)

	suite.scriptConsumer = scriptRunner.NewConsumer(testLog, dir)

	suite.runner = notifications.NewProcessor(logger.NewLogger("notifications", logger.NewLogOutput("out.log"), logger.LogLevelInfo), suite.store, suite.mailConsumer, suite.scriptConsumer)

}

type ScriptIO struct {
	Recipients []string `json:"recipients"`
	Data       string   `json:"data"`
}

func (suite *NotificationsIntegrationTestSuite) TestDispatcherCreatesNotification() {
	notification := notifications.NotificationData{
		Target:      "smtp",
		Recipients:  []string{"stefan.tester@example.com"},
		Subject:     "test-subject",
		Content:     "test-content-mail",
		ContentType: notifications.ContentTypeTextHTML,
	}
	_, err := suite.dispatcher.Dispatch(context.Background(), problemIdentifiable, notification)
	suite.NoError(err)

	notification = notifications.NotificationData{
		Target:      "./test.sh",
		Recipients:  []string{"r1@example.com", "somethin323-55@test.co"},
		Subject:     "test-subject",
		Content:     "test-content",
		ContentType: notifications.ContentTypeTextPlain,
	}
	d, err := suite.dispatcher.Dispatch(context.Background(), problemIdentifiable, notification)
	suite.NoError(err)
	time.Sleep(time.Millisecond * 100)
	suite.T().Log(suite.store.Details(context.Background(), d.ID()))

	suite.ExpectedMessages(1)
	// suite.ExpectMessage(notification.Recipients, notification.Subject, string(notification.ContentType), notification.Content)

	in := ScriptIO{
		Recipients: []string{"r1@example.com", "somethin323-55@test.co"},
		Data:       "test-content",
	}

	out, err := simpleops.ReadJSONFileIntoStruct[ScriptIO]("out.json")
	suite.NoError(err)
	suite.Equal(in, out)
}

func (suite *NotificationsIntegrationTestSuite) ExpectedMessages(count int) bool {
	return suite.Len(suite.server.Messages(), count)
}

func (suite *NotificationsIntegrationTestSuite) ExpectMessage(to []string, subject string, contentType string, content string) {
	if !suite.ExpectedMessages(1) {
		return
	}
	receivedMail := rmailer.ReceivedMail{Message: suite.server.Messages()[0]}

	suite.Equal(to, receivedMail.GetTo())

	suite.Equal(subject, receivedMail.GetSubject())

	suite.Equal(contentType, receivedMail.GetContentType())

	suite.Equal(content, receivedMail.GetContent())

}

func TestNotificationsIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationsIntegrationTestSuite))
}
