package storage

import "sync"

var Storage *MemStorage

type MemStorage struct {
	Metrics Metrics
}

type Metrics struct {
	Counters map[string]int64
	Gauges   map[string]float64
	mx       sync.RWMutex
}

func (s *MemStorage) AddCounterValue(metricName string, value int64) {
	curVal, hasValue := s.GetCounterValue(metricName)
	if !hasValue {
		s.setCounterValue(metricName, value)
		return
	}

	s.setCounterValue(metricName, curVal+value)
}

func (s *MemStorage) GetCounterValue(metricName string) (int64, bool) {
	s.Metrics.mx.RLock()
	defer s.Metrics.mx.RUnlock()

	mapCopy := make(map[string]int64, len(s.Metrics.Counters))
	for key, val := range s.Metrics.Counters {
		mapCopy[key] = val
	}

	v, hasValue := mapCopy[metricName]
	return v, hasValue
}

func (s *MemStorage) setCounterValue(metricName string, value int64) {
	s.Metrics.mx.Lock()
	defer s.Metrics.mx.Unlock()

	mapCopy := make(map[string]int64, len(s.Metrics.Counters))
	for key, val := range s.Metrics.Counters {
		mapCopy[key] = val
	}

	mapCopy[metricName] = value
	s.Metrics.Counters = mapCopy
}

func (s *MemStorage) GetGaugeValue(metricName string) (float64, bool) {
	s.Metrics.mx.RLock()
	defer s.Metrics.mx.RUnlock()

	v, hasValue := s.Metrics.Gauges[metricName]
	return v, hasValue
}

func (s *MemStorage) SetGaugeValue(metricName string, value float64) {
	s.Metrics.mx.Lock()
	defer s.Metrics.mx.Unlock()

	s.Metrics.Gauges[metricName] = value
}

func NewMemStorage() MemStorage {
	return MemStorage{Metrics: Metrics{Counters: make(map[string]int64), Gauges: make(map[string]float64)}}
}

func GetStorage() *MemStorage {
	if Storage == nil {
		ms := NewMemStorage()
		Storage = &ms
	}
	return Storage
}
