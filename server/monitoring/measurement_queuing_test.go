package monitoring_test

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/riportdev/riport/server/monitoring"
	"github.com/riportdev/riport/share/logger"
	"github.com/riportdev/riport/share/models"
)

var testLog = logger.NewLogger("measurement-queue", logger.LogOutput{File: os.Stdout}, logger.LogLevelDebug)

type MockSaver struct {
	ms         []*models.Measurement
	count      atomic.Int64
	slow       atomic.Bool
	saveCalled chan struct{}
}

func (m *MockSaver) SaveMeasurement(ctx context.Context, measurement *models.Measurement) error {
	if m.slow.Load() {
		time.Sleep(time.Millisecond * 10)
	}
	m.ms = append(m.ms, measurement)
	m.count.Add(1)
	close(m.saveCalled)
	return nil
}

type QueuingTestSuite struct {
	suite.Suite
	q     monitoring.MeasurementSaver
	saver *MockSaver
}

func (suite *QueuingTestSuite) SetupTest() {
	suite.saver = &MockSaver{
		ms:         make([]*models.Measurement, 0),
		saveCalled: make(chan struct{}),
	}
	suite.q = monitoring.NewMeasurementQueuing(testLog, suite.saver, 0)
}

func (suite *QueuingTestSuite) TestEnqueue() {
	suite.q.Notify(models.Measurement{})
	<-suite.saver.saveCalled
	suite.Equal(suite.saver.count.Load(), int64(1))
}

func (suite *QueuingTestSuite) TestSlowEnqueue() {
	suite.saver.slow.Store(true)
	stopper := time.Now()
	suite.q.Notify(models.Measurement{})

	suite.Less(time.Since(stopper), time.Millisecond)
}

func (suite *QueuingTestSuite) TestCleanClose() {
	suite.saver.slow.Store(true)
	suite.q.Notify(models.Measurement{})
	_ = suite.q.Close()
	suite.q.Notify(models.Measurement{})
	suite.Equal(suite.saver.count.Load(), int64(1))
}

func TestMeasurementQueuingTestSuite(t *testing.T) {
	suite.Run(t, new(QueuingTestSuite))
}
