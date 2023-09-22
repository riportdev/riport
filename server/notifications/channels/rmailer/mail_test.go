package rmailer_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"github.com/stretchr/testify/suite"

	"github.com/riportdev/riport/server/chconfig"
	"github.com/riportdev/riport/server/notifications/channels/rmailer"
	"github.com/riportdev/riport/share/logger"
)

var testLog = logger.NewLogger("client", logger.LogOutput{File: os.Stdout}, logger.LogLevelDebug)

type MailTestSuite struct {
	suite.Suite
	server *smtpmock.Server
	mailer rmailer.Mailer
}

func (ts *MailTestSuite) SetupSuite() {
	ts.server = smtpmock.New(smtpmock.ConfigurationAttr{
		//LogToStdout:              true, // for debugging (especially connection)
		//LogServerActivity:        true, // for debugging (especially connection)
		MultipleMessageReceiving: true,
		//		PortNumber:        33334, // randomly generated
	})
	ts.NoError(ts.server.Start())

	ts.mailer = rmailer.NewRMailer(rmailer.Config{
		Host:     "localhost",
		Port:     ts.server.PortNumber(),
		Domain:   "example.com",
		From:     "test@example.com",
		TLS:      false,
		AuthType: rmailer.AuthTypeNone,
		NoNoop:   true,
	}, testLog)

	if err := ts.server.Start(); err != nil {
		fmt.Println(err)
	}
}

func (ts *MailTestSuite) TestMailTimeout() {

	port := 11111
	ts.neverRespondingSMTPServer(port)

	mailer := ts.mailerFromPort(port)

	start := time.Now()

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Millisecond*2)
	defer cancelFn()

	err := mailer.Send(ctx, []string{"tina.recipient@example.com", "just+fff@some.mail.com"}, "test subject!", rmailer.ContentTypeTextHTML, "test\r\n\r\n<b>content</b>")
	ts.T().Log(err)
	ts.Error(err)

	ts.WithinRange(time.Now(), start.Add(time.Microsecond*500), start.Add(time.Millisecond*100))
}

func (ts *MailTestSuite) TestMailErrorOnTooManyHangingConnections() {

	port := 11112
	ts.neverRespondingSMTPServer(port)

	mailer := ts.mailerFromPort(port)

	mailCount := rmailer.MaxHangingMailSends * 10

	for i := 0; i < mailCount; i++ {
		go func() {
			//!!! mailer should not be run asynchronously, it's only run here like this to ensure error queue to fill up
			_ = mailer.Send(context.Background(), []string{"tina.recipient@example.com", "just+fff@some.mail.com"}, "test subject!", rmailer.ContentTypeTextHTML, "test\r\n\r\n<b>content</b>")
		}()
	}
	time.Sleep(time.Millisecond * 100) // wait not for all messages to be sent (could use wait group for that) but to enqueue messages into error queue

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancelFn()
	ts.ErrorContains(mailer.Send(ctx, []string{"tina.recipient@example.com", "just+fff@some.mail.com"}, "test subject!", rmailer.ContentTypeTextHTML, "test\r\n\r\n<b>content</b>"), "server non-responsive")

}

func (ts *MailTestSuite) mailerFromPort(port int) rmailer.Mailer {
	return rmailer.NewRMailer(rmailer.Config{
		Host:     "localhost",
		Port:     port,
		Domain:   "example.com",
		From:     "test@example.com",
		TLS:      false,
		AuthType: rmailer.AuthTypeNone,
		NoNoop:   true,
	}, testLog)
}

func (ts *MailTestSuite) neverRespondingSMTPServer(port int) {
	logError := ts.NoError
	go func() {
		listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
		logError(err)
		for {
			_, _ = listener.Accept()
		}
	}()
}

func (ts *MailTestSuite) TestMailSent() {
	ts.NoError(ts.mailer.Send(context.Background(), []string{"tina.recipient@example.com", "just+fff@some.mail.com"}, "test subject!", rmailer.ContentTypeTextHTML, "test\r\n\r\n<b>content</b>"))
	ts.ExpectMessage([]string{"tina.recipient@example.com", "just+fff@some.mail.com"}, "test subject!", "text/html; charset=UTF-8", "test\r\n\r\n<b>content</b>")
}

func (ts *MailTestSuite) TestMailSMTPConfigCompatibility() {

	config, err := rmailer.ConfigFromSMTPConfig(chconfig.SMTPConfig{
		Server:       "testsmtp.somedomain.com",
		AuthUsername: "test-user",
		AuthPassword: "test-password",
		SenderEmail:  "test@somedomain.com",
		Secure:       true,
	})

	ts.NoError(err)

	ts.Equal(rmailer.Config{
		Host:     "testsmtp.somedomain.com",
		Port:     -1,
		Domain:   "somedomain.com",
		From:     "test@somedomain.com",
		TLS:      true,
		AuthType: rmailer.AuthTypeUserPass,
		AuthUserPass: rmailer.AuthUserPass{
			User: "test-user",
			Pass: "test-password",
		},
		NoNoop: false,
	}, config)
}

func (ts *MailTestSuite) TestMailContentTypeValidation() {
	ts.Error(ts.mailer.Send(context.Background(), []string{"tina.recipient@example.com", "just+fff@some.mail.com"}, "test subject!", rmailer.ContentType("test"), "test\r\n\r\n<b>content</b>"))
}

func (ts *MailTestSuite) TestContentTypeValidation() {
	ts.NoError(rmailer.ContentType("text/html").Valid())
	ts.NoError(rmailer.ContentType("text/plain").Valid())
	ts.Error(rmailer.ContentType("should-fail").Valid())
}

func (ts *MailTestSuite) ExpectedMessages(count int) bool {
	return ts.Len(ts.server.Messages(), count)
}

func (ts *MailTestSuite) ExpectMessage(to []string, subject string, contentType string, content string) {
	if !ts.ExpectedMessages(1) {
		return
	}
	receivedMail := rmailer.ReceivedMail{Message: ts.server.Messages()[0]}

	ts.Equal(to, receivedMail.GetTo())

	ts.Equal(subject, receivedMail.GetSubject())

	ts.Equal(contentType, receivedMail.GetContentType())

	ts.Equal(content, receivedMail.GetContent())

}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMailTestSuite(t *testing.T) {
	suite.Run(t, new(MailTestSuite))
}
