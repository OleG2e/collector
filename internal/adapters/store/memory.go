package store

import (
	"context"
	"sync"

	"collector/internal/core/domain"
)

type MemStorage struct {
	metrics *domain.Metrics
	mx      *sync.RWMutex
}

func NewMemoryStorage(metrics *domain.Metrics) *MemStorage {
	return &MemStorage{
		metrics: metrics,
		mx:      new(sync.RWMutex),
	}
}

func (f *MemStorage) GetStoreType() Type {
	return MemoryStoreType
}

func (f *MemStorage) Save(_ context.Context) error {
	return nil
}

func (f *MemStorage) Restore(_ context.Context) error {
	return nil
}

func (f *MemStorage) GetMetrics() *domain.Metrics {
	f.mx.RLock()
	defer f.mx.RUnlock()

	return f.metrics
}

func (f *MemStorage) SetMetrics(metrics *domain.Metrics) {
	f.mx.Lock()
	defer f.mx.Unlock()

	f.metrics = metrics
}

func (f *MemStorage) Close() error {
	return nil
}
