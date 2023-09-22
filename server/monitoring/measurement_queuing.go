package monitoring

import (
	"context"
	"sync"
	"time"

	"github.com/riportdev/riport/share/logger"
	"github.com/riportdev/riport/share/models"
)

type saver interface {
	SaveMeasurement(ctx context.Context, measurement *models.Measurement) error
}

type MeasurementSaver interface {
	Notify(models.Measurement) bool
	Close() error
}

type queue struct {
	saver    saver
	queue    chan models.Measurement
	ctx      context.Context
	cancelFn context.CancelFunc
	wg       sync.WaitGroup
	logger   *logger.Logger
}

func (q *queue) Close() error {
	q.cancelFn() // Signal the context first.
	q.wg.Wait()  // Wait for all the goroutines to finish.
	return nil
}

func (q *queue) Notify(measurement models.Measurement) bool {
	select {
	case <-q.ctx.Done():
		return false
	case q.queue <- measurement:
		return true
	}
}

func (q *queue) process() {
	defer q.wg.Done()

	// Process items until shutdown signal is received
	for {
		select {
		case <-q.ctx.Done():
			q.saveAllEnqueuedMeasurements() // drain chan
			return                          // quit processing
		case m := <-q.queue:
			if err := q.saver.SaveMeasurement(q.ctx, &m); err != nil {
				q.logger.Errorf("Failed to save measurement for client %s: %s", m.ClientID, err)
			}
		}
	}

}

func (q *queue) saveAllEnqueuedMeasurements() {
	for {
		select {
		case m := <-q.queue:
			if err := q.saver.SaveMeasurement(q.ctx, &m); err != nil {
				q.logger.Errorf("Failed to save measurement for client %s: %s", m.ClientID, err)
			}
		default:
			return
		}
	}
}

func (q *queue) logQueueLength() {
	defer q.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-q.ctx.Done():
			return
		case <-ticker.C:
			length := len(q.queue)
			if length > 0 {
				q.logger.Debugf("Enqueued measurements: %v", length)
			}
		}
	}
}

func NewMeasurementQueuing(logger *logger.Logger, saver saver, queueSize int) MeasurementSaver {
	ctx, cfn := context.WithCancel(context.Background())

	q := &queue{
		saver:    saver,
		queue:    make(chan models.Measurement, queueSize),
		ctx:      ctx,
		cancelFn: cfn,
		logger:   logger,
	}

	q.wg.Add(2)
	go q.process()
	go q.logQueueLength()

	return q
}
