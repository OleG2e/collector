package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/OleG2e/collector/internal/config"
	"github.com/OleG2e/collector/pkg/logging"
	"go.uber.org/zap"
)

type MemStorage struct {
	Metrics Metrics
	DBFile  *os.File
	conf    *config.ServerConfig
	l       *logging.ZapLogger
	ctx     context.Context
	mx      *sync.RWMutex
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

func NewMemStorage(ctx context.Context, l *logging.ZapLogger, conf *config.ServerConfig) *MemStorage {
	ms := &MemStorage{
		Metrics: newMetrics(),
		l:       l,
		ctx:     ctx,
		conf:    conf,
		mx:      &sync.RWMutex{},
	}

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

	file, fileErr := os.Open(ms.conf.FileStoragePath)

	defer func(file *os.File) {
		fileCloseErr := file.Close()
		if fileCloseErr != nil {
			ms.l.WarnCtx(ms.ctx, "file close error", zap.Error(fileCloseErr))
		}
	}(file)

	if file == nil {
		return
	}

	if fileErr != nil {
		ms.l.WarnCtx(ms.ctx, "open DB file error", zap.Error(fileErr))
		return
	}

	dec := json.NewDecoder(bufio.NewReader(file))

	lastState := newMetrics()

	if err := dec.Decode(&lastState); err != nil && err != io.EOF {
		ms.l.FatalCtx(ms.ctx, "restore storage error", zap.Error(err))
	}

	ms.Metrics = lastState
}

func (ms *MemStorage) FlushStorage() error {
	ms.mx.Lock()
	defer ms.mx.Unlock()

	dir := path.Dir(ms.conf.FileStoragePath)
	tmpFile, tmpFileErr := os.CreateTemp(dir, "collector-*.bak")
	if tmpFileErr != nil {
		return tmpFileErr
	}

	defer func(tmpFile *os.File) {
		fileCloseErr := tmpFile.Close()
		if fileCloseErr != nil {
			ms.l.WarnCtx(ms.ctx, "tmp file close error", zap.Error(fileCloseErr))
		}
	}(tmpFile)

	data, err := json.Marshal(&ms.Metrics)

	if err != nil {
		return err
	}

	_, err = tmpFile.Write(data)

	if err != nil {
		return err
	}

	err = os.Rename(tmpFile.Name(), ms.conf.FileStoragePath)

	ms.l.InfoCtx(ms.ctx, "flush storage")

	return err
}
