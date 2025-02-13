package storage

import (
	"context"
	"sync"
	"time"

	"github.com/OleG2e/collector/internal/config"
	"github.com/OleG2e/collector/pkg/logging"
	"go.uber.org/zap"
)

type StoreType string

const (
	FileStoreType = StoreType("file")
	DBStoreType   = StoreType("db")
)

type StoreAlgo interface {
	store(m *Metrics) error
	restore() (*Metrics, error)
	CloseStorage() error
	GetStoreType() StoreType
}

type MemStorage struct {
	Metrics   *Metrics
	conf      *config.ServerConfig
	l         *logging.ZapLogger
	ctx       context.Context
	mx        *sync.RWMutex
	storeAlgo StoreAlgo
}

type Metrics struct {
	Counters map[string]int64   `json:"counters"`
	Gauges   map[string]float64 `json:"gauges"`
}

func (ms *MemStorage) store() error {
	return ms.storeAlgo.store(ms.Metrics)
}

func (ms *MemStorage) setStoreAlgo(sa StoreAlgo) {
	ms.storeAlgo = sa
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

func GetStoreAlgo(ctx context.Context, l *logging.ZapLogger, conf *config.ServerConfig) StoreAlgo {
	dbStorage, dbErr := NewDBStorage(ctx, l, conf)
	if dbErr != nil {
		fileStorage, fileErr := NewFileStorage(ctx, l, conf)
		if fileErr != nil {
			l.PanicCtx(ctx, "failed to create storage", zap.Error(fileErr))
		}
		return fileStorage
	}
	return dbStorage
}

func NewMemStorage(
	ctx context.Context,
	l *logging.ZapLogger,
	conf *config.ServerConfig,
	storeAlgo StoreAlgo,
) *MemStorage {
	ms := &MemStorage{
		Metrics:   newMetrics(),
		l:         l,
		ctx:       ctx,
		conf:      conf,
		storeAlgo: storeAlgo,
		mx:        &sync.RWMutex{},
	}

	return ms
}

func newMetrics() *Metrics {
	return &Metrics{
		Counters: make(map[string]int64),
		Gauges:   make(map[string]float64),
	}
}

func (ms *MemStorage) GetStoreAlgo() StoreAlgo {
	return ms.storeAlgo
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

func (ms *MemStorage) RestoreStorage() error {
	ms.mx.Lock()
	defer ms.mx.Unlock()

	metrics, err := ms.storeAlgo.restore()
	if err != nil {
		return err
	}

	if metrics != nil {
		ms.Metrics = metrics
	}

	return nil
}

func (ms *MemStorage) FlushStorage() error {
	ms.mx.Lock()
	defer ms.mx.Unlock()

	ms.l.InfoCtx(ms.ctx, "flush storage")

	return ms.store()
}

func (ms *MemStorage) CloseStorage() error {
	return ms.storeAlgo.CloseStorage()
}
