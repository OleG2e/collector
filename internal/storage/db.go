package storage

import (
	"context"
	"errors"
	"time"

	"github.com/OleG2e/collector/internal/config"
	"github.com/OleG2e/collector/pkg/db"
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
	conf *config.ServerConfig,
) (*DBStorage, error) {
	if conf.GetDSN() == "" {
		return nil, errors.New("DSN is not provided")
	}

	poolConn, dbErr := db.NewPoolConn(ctx, conf)
	if dbErr != nil {
		return nil, dbErr
	}

	return &DBStorage{
		ctx:      ctx,
		l:        l,
		poolConn: poolConn,
	}, nil
}

func (d *DBStorage) GetStoreType() StoreType {
	return DBStoreType
}

func (d *DBStorage) store(m *Metrics) error {
	ctx := context.Background()
	tx, err := d.poolConn.Begin(ctx)

	if err != nil {
		return err
	}

	defer func() {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			d.l.WarnCtx(ctx, "transaction commit error", zap.Error(commitErr))
			if rErr := tx.Rollback(ctx); rErr != nil {
				d.l.WarnCtx(ctx, "transaction rollback error", zap.Error(commitErr))
			}
		}
	}()

	_, truncErr := tx.Exec(ctx, "TRUNCATE TABLE counters; TRUNCATE TABLE gauges")
	if truncErr != nil {
		return truncErr
	}

	now := time.Now()
	for k, v := range m.Gauges {
		_, txErr := tx.Exec(ctx, "INSERT INTO gauges (name, value, created_at) VALUES ($1, $2, $3)", k, v, now)
		if txErr != nil {
			d.l.WarnCtx(ctx, "transaction error", zap.Error(txErr))
			return txErr
		}
	}

	for k, v := range m.Counters {
		_, txErr := tx.Exec(ctx, "INSERT INTO counters (name, value, created_at) VALUES ($1, $2, $3)", k, v, now)
		if txErr != nil {
			d.l.WarnCtx(ctx, "transaction error", zap.Error(txErr))
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

	queryG := "SELECT name, value FROM gauges"
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

	queryC := "SELECT name, value FROM counters"
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

func (d *DBStorage) CloseStorage() error {
	if d.poolConn != nil {
		d.poolConn.Close()
	}
	return nil
}
