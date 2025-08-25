package store

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"collector/internal/config"
	"collector/internal/core/domain"
	"collector/pkg/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBStorage struct {
	logger   *slog.Logger
	poolConn *pgxpool.Pool
	metrics  *domain.Metrics
	mx       *sync.RWMutex
}

func NewDBStorage(
	ctx context.Context,
	logger *slog.Logger,
	conf *config.ServerConfig,
	metrics *domain.Metrics,
) (*DBStorage, error) {
	if conf.GetDSN() == "" {
		return nil, fmt.Errorf("dsn is empty")
	}

	poolConn, pollConnErr := db.NewPoolConn(ctx, conf.GetDSN())
	if pollConnErr != nil {
		return nil, fmt.Errorf("(db) get new poll connection error: %w", pollConnErr)
	}

	return &DBStorage{
		logger:   logger,
		poolConn: poolConn,
		metrics:  metrics,
		mx:       new(sync.RWMutex),
	}, nil
}

func (d *DBStorage) GetStoreType() Type {
	return DBStoreType
}

func (d *DBStorage) Save(ctx context.Context) error {
	tx, err := d.poolConn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin (db) transaction error: %w", err)
	}

	defer func() {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			d.logger.WarnContext(ctx, "(db) transaction commit error", slog.Any("error", commitErr))

			if rErr := tx.Rollback(ctx); rErr != nil {
				d.logger.WarnContext(
					ctx,
					"(db) transaction rollback error",
					slog.Any("error", commitErr),
				)
			}
		}
	}()

	_, truncErr := tx.Exec(ctx, "TRUNCATE TABLE counters; TRUNCATE TABLE gauges")
	if truncErr != nil {
		return fmt.Errorf("(db) truncate table error: %w", truncErr)
	}

	now := time.Now()
	for k, v := range d.GetMetrics().GetGauges() {
		_, txErr := tx.Exec(
			ctx,
			"INSERT INTO gauges (name, value, created_at) VALUES ($1, $2, $3)",
			k,
			v,
			now,
		)
		if txErr != nil {
			return fmt.Errorf("(db) transaction insert gauge error: %w", txErr)
		}
	}

	for k, v := range d.GetMetrics().GetCounters() {
		_, txErr := tx.Exec(
			ctx,
			"INSERT INTO counters (name, value, created_at) VALUES ($1, $2, $3)",
			k,
			v,
			now,
		)
		if txErr != nil {
			return fmt.Errorf("(db) transaction insert counter error: %w", txErr)
		}
	}

	return nil
}

func (d *DBStorage) Restore(ctx context.Context) error {
	metrics := domain.NewMetrics()

	queryG := "SELECT name, value FROM gauges"
	gauges, gQueryErr := d.poolConn.Query(ctx, queryG)

	if gQueryErr != nil {
		return fmt.Errorf("(db) select gauges error: %w", gQueryErr)
	}

	defer gauges.Close()

	for gauges.Next() {
		var gauge domain.Gauge

		gTxErr := gauges.Scan(&gauge.Name, &gauge.Value)
		if gTxErr != nil {
			return fmt.Errorf("(db) scan gauge error: %w", gQueryErr)
		}

		metrics.Gauges[gauge.Name] = gauge.Value
	}

	if gReadErr := gauges.Err(); gReadErr != nil {
		return fmt.Errorf("(db) read gauges error: %w", gReadErr)
	}

	queryC := "SELECT name, value FROM counters"

	counters, cQueryErr := d.poolConn.Query(ctx, queryC)
	if cQueryErr != nil {
		return fmt.Errorf("(db) select counters error: %w", cQueryErr)
	}

	defer counters.Close()

	for counters.Next() {
		var c domain.Counter

		cTxErr := counters.Scan(&c.Name, &c.Value)
		if cTxErr != nil {
			return fmt.Errorf("(db) scan counter error: %w", cTxErr)
		}

		metrics.Counters[c.Name] = c.Value
	}

	if cReadErr := counters.Err(); cReadErr != nil {
		return fmt.Errorf("(db) read counters error: %w", cReadErr)
	}

	d.logger.DebugContext(ctx, "restored state", slog.Any("state", &metrics))

	d.SetMetrics(metrics)

	return nil
}

func (d *DBStorage) GetMetrics() *domain.Metrics {
	d.mx.RLock()
	defer d.mx.RUnlock()

	return d.metrics
}

func (d *DBStorage) SetMetrics(metrics *domain.Metrics) {
	d.mx.Lock()
	defer d.mx.Unlock()

	d.metrics = metrics
}

func (d *DBStorage) Close() error {
	if d.poolConn != nil {
		d.poolConn.Close()
	}

	return nil
}
