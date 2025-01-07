package storage

import (
	"context"
	"time"

	"github.com/OleG2e/collector/pkg/logging"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DBStorage struct {
	ctx      context.Context
	l        *logging.ZapLogger
	poolConn *pgxpool.Pool
}

type Counter struct {
	Name  string
	Value int64
}

type Gauge struct {
	Name  string
	Value float64
}

func NewDBStorage(
	ctx context.Context,
	l *logging.ZapLogger,
	poolConn *pgxpool.Pool,
) *DBStorage {

	return &DBStorage{
		ctx:      ctx,
		l:        l,
		poolConn: poolConn,
	}
}

func (d *DBStorage) GetStoreType() StoreType {
	return dbStore
}

func (d *DBStorage) store(m *Metrics) error {
	tx, err := d.poolConn.Begin(d.ctx)

	defer func() {
		if commitErr := tx.Commit(d.ctx); commitErr != nil {
			d.l.WarnCtx(d.ctx, "transaction commit error", zap.Error(commitErr))
			if rErr := tx.Rollback(d.ctx); rErr != nil {
				d.l.WarnCtx(d.ctx, "transaction rollback error", zap.Error(commitErr))
			}
		}
	}()

	if err != nil {
		return err
	}

	now := time.Now()
	for k, v := range m.Gauges {
		_, txErr := tx.Exec(d.ctx, "INSERT INTO gauges (name, value, created_at, updated_at) VALUES ($1, $2, $3, $4)", k, v, now, now)
		if txErr != nil {
			d.l.WarnCtx(d.ctx, "transaction error", zap.Error(txErr))
			return txErr
		}
	}

	for k, v := range m.Counters {
		_, txErr := tx.Exec(d.ctx, "INSERT INTO counters (name, value, created_at, updated_at) VALUES ($1, $2, $3, $4)", k, v, now, now)
		if txErr != nil {
			d.l.WarnCtx(d.ctx, "transaction error", zap.Error(txErr))
			return txErr
		}
	}

	return nil
}

func (d *DBStorage) restore() (*Metrics, error) {
	m := Metrics{
		Counters: make(map[string]int64),
		Gauges:   make(map[string]float64),
	}

	queryG := "SELECT DISTINCT ON (name) name, value FROM (SELECT name, value FROM gauges ORDER BY created_at DESC)"
	gauges, gQueryErr := d.poolConn.Query(d.ctx, queryG)
	if gQueryErr != nil {
		return nil, gQueryErr
	}

	defer gauges.Close()

	for gauges.Next() {
		var g Gauge
		gTxErr := gauges.Scan(&g.Name, &g.Value)
		if gTxErr != nil {
			d.l.WarnCtx(d.ctx, "scan gauge error", zap.Error(gTxErr))
			return nil, gTxErr
		}
		m.Gauges[g.Name] = g.Value
	}

	if gReadErr := gauges.Err(); gReadErr != nil {
		d.l.WarnCtx(d.ctx, "scan gauges error", zap.Error(gReadErr))
		return nil, gReadErr
	}

	queryC := "SELECT DISTINCT ON (name) name, value FROM (SELECT name, value FROM counters ORDER BY created_at DESC)"
	counters, cQueryErr := d.poolConn.Query(d.ctx, queryC)
	if cQueryErr != nil {
		return nil, cQueryErr
	}

	defer counters.Close()

	for counters.Next() {
		var c Counter
		cTxErr := counters.Scan(&c.Name, &c.Value)
		if cTxErr != nil {
			d.l.WarnCtx(d.ctx, "scan counter error", zap.Error(cTxErr))
			return nil, cTxErr
		}
		m.Counters[c.Name] = c.Value
	}

	if cReadErr := counters.Err(); cReadErr != nil {
		d.l.WarnCtx(d.ctx, "scan counters error", zap.Error(cReadErr))
		return nil, cReadErr
	}

	d.l.DebugCtx(d.ctx, "restored state", zap.Any("state", &m))

	return &m, nil
}
