package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/OleG2e/collector/internal/config"
	"github.com/OleG2e/collector/pkg/logging"
	"go.uber.org/zap"
)

type MemStorage struct {
	Metrics  Metrics
	DBFile   *os.File
	conf     *config.ServerConfig
	l        *logging.ZapLogger
	ctx      context.Context
	mx       *sync.RWMutex
	PoolConn *pgxpool.Pool
}

type Metrics struct {
	Counters map[string]int64   `json:"counters"`
	Gauges   map[string]float64 `json:"gauges"`
}

func (ms *MemStorage) AddCounterValue(metricName string, value int64) {
	curVal, hasValue := ms.GetCounterValue(metricName)
	if !hasValue {
		ms.setCounterValue(metricName, value)
		return
	}

	ms.setCounterValue(metricName, curVal+value)
}

func (ms *MemStorage) GetCounterValue(metricName string) (int64, bool) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()

	v, hasValue := ms.Metrics.Counters[metricName]
	return v, hasValue
}

func (ms *MemStorage) setCounterValue(metricName string, value int64) {
	ms.mx.Lock()
	defer ms.mx.Unlock()

	ms.Metrics.Counters[metricName] = value
}

func (ms *MemStorage) GetGaugeValue(metricName string) (float64, bool) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()

	v, hasValue := ms.Metrics.Gauges[metricName]
	return v, hasValue
}

func (ms *MemStorage) SetGaugeValue(metricName string, value float64) {
	ms.mx.Lock()
	defer ms.mx.Unlock()

	ms.Metrics.Gauges[metricName] = value
}

func NewMemStorage(ctx context.Context, l *logging.ZapLogger, conf *config.ServerConfig, poolConn *pgxpool.Pool) *MemStorage {
	ms := &MemStorage{
		Metrics:  newMetrics(),
		l:        l,
		ctx:      ctx,
		conf:     conf,
		PoolConn: poolConn,
		mx:       &sync.RWMutex{},
	}
	ms.openDBFile()

	return ms
}

func newMetrics() Metrics {
	return Metrics{
		Counters: make(map[string]int64),
		Gauges:   make(map[string]float64),
	}
}

func (ms *MemStorage) InitFlushStorageTicker(storeInterval time.Duration) {
	ticker := time.NewTicker(storeInterval)
	go func() {
		for range ticker.C {
			if err := ms.FlushStorage(); err != nil {
				ms.l.ErrorCtx(ms.ctx, "flush storage error", zap.Error(err))
			}
		}
	}()
}

func (ms *MemStorage) RestoreStorage() {
	ms.mx.Lock()
	defer ms.mx.Unlock()

	reader := bufio.NewReader(ms.DBFile)
	dec := json.NewDecoder(reader)

	lastState := newMetrics()

	if err := dec.Decode(&lastState); err != nil && err != io.EOF {
		ms.l.FatalCtx(ms.ctx, "restore storage error", zap.Error(err))
	}

	ms.Metrics = lastState
}

func (ms *MemStorage) FlushStorage() error {
	ms.mx.Lock()
	defer ms.mx.Unlock()

	data, err := json.Marshal(&ms.Metrics)

	if err != nil {
		return err
	}

	if ms.DBFile == nil {
		ms.openDBFile()

		defer func(file *os.File) {
			fileCloseErr := file.Close()
			if fileCloseErr != nil && !errors.Is(fileCloseErr, os.ErrClosed) {
				ms.l.ErrorCtx(ms.ctx, "file close error", zap.Error(fileCloseErr))
			}
			ms.DBFile = nil
		}(ms.DBFile)
	}

	_, err = ms.DBFile.WriteAt(data, 0)

	ms.l.InfoCtx(ms.ctx, "flush storage")

	return err
}

func (ms *MemStorage) openDBFile() {
	file, fileErr := os.OpenFile(ms.conf.FileStoragePath, os.O_RDWR|os.O_CREATE, 0o666)

	if fileErr != nil {
		ms.l.FatalCtx(ms.ctx, "open DB file error", zap.Error(fileErr))
	}

	ms.DBFile = file
}
