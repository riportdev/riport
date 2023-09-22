package auditlog

import (
	"context"
	"database/sql"
	"os"
	"path"
	"sync"
	"time"

	"github.com/riportdev/riport/db/sqlite"

	"github.com/riportdev/riport/share/logger"
	"github.com/riportdev/riport/share/query"
)

const (
	sqliteFilename  = "auditlog.db"
	rotatedFilename = "auditlog.2006-01-02.db"
)

type RotationProvider struct {
	logger            *logger.Logger
	period            time.Duration
	ticker            *time.Ticker
	dataDir           string
	dataSourceOptions sqlite.DataSourceOptions

	mtx    sync.RWMutex
	sqlite *SQLiteProvider
}

func newRotationProvider(l *logger.Logger, period time.Duration, dataDir string, dataSourceOptions sqlite.DataSourceOptions) (*RotationProvider, error) {
	sqlite, err := newSQLiteProvider(dataDir, dataSourceOptions)
	if err != nil {
		return nil, err
	}

	r := &RotationProvider{
		logger:  l,
		period:  period,
		dataDir: dataDir,
		sqlite:  sqlite,
		ticker:  time.NewTicker(period),
	}
	err = r.rotateIfNeeded()
	if err != nil {
		return nil, err
	}

	go r.rotationLoop()

	return r, nil
}

func (r *RotationProvider) rotationLoop() {
	for range r.ticker.C {
		err := r.rotate()
		if err != nil {
			r.logger.Errorf("Could not rotate auditlog: %v", err)
		}
	}
}

func (r *RotationProvider) rotate() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	err := r.sqlite.Close()
	if err != nil {
		return err
	}

	sqliteFn := path.Join(r.dataDir, sqliteFilename)
	rotatedFn := path.Join(r.dataDir, time.Now().Format(rotatedFilename))
	err = os.Rename(sqliteFn, rotatedFn)
	if err != nil {
		return err
	}

	r.sqlite, err = newSQLiteProvider(r.dataDir, r.dataSourceOptions)
	if err != nil {
		return err
	}

	return nil
}

func (r *RotationProvider) rotateIfNeeded() error {
	oldest, err := r.sqlite.OldestTimestamp(context.Background())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	if time.Since(oldest) > r.period {
		return r.rotate()
	}

	return nil
}

func (r *RotationProvider) Save(e *Entry) error {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	return r.sqlite.Save(e)
}
func (r *RotationProvider) List(ctx context.Context, l *query.ListOptions) ([]*Entry, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	return r.sqlite.List(ctx, l)
}
func (r *RotationProvider) Count(ctx context.Context, l *query.ListOptions) (int, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	return r.sqlite.Count(ctx, l)
}
func (r *RotationProvider) Close() error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.ticker.Stop()
	return r.sqlite.Close()
}
