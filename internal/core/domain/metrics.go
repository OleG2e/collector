package domain

import "sync"

type Counter struct {
	Name  string
	Value int64
}

type Gauge struct {
	Name  string
	Value float64
}

type Metrics struct {
	Counters map[string]int64   `json:"counters"`
	Gauges   map[string]float64 `json:"gauges"`
	mx       *sync.RWMutex
}

func NewMetrics() *Metrics {
	return &Metrics{
		Counters: make(map[string]int64),
		Gauges:   make(map[string]float64),
		mx:       new(sync.RWMutex),
	}
}

func (m *Metrics) AddCounterValue(metricName string, value int64) {
	curVal, hasValue := m.GetCounterValue(metricName)
	if !hasValue {
		m.setCounterValue(metricName, value)

		return
	}

	m.setCounterValue(metricName, curVal+value)
}

func (m *Metrics) GetCounterValue(metricName string) (int64, bool) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	v, hasValue := m.Counters[metricName]

	return v, hasValue
}

func (m *Metrics) GetGaugeValue(metricName string) (float64, bool) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	v, hasValue := m.Gauges[metricName]

	return v, hasValue
}

func (m *Metrics) SetGaugeValue(metricName string, value float64) {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.Gauges[metricName] = value
}

func (m *Metrics) GetGauges() map[string]float64 {
	m.mx.RLock()
	defer m.mx.RUnlock()

	mapCopy := make(map[string]float64, len(m.Gauges))
	for key, val := range m.Gauges {
		mapCopy[key] = val
	}

	return mapCopy
}

func (m *Metrics) GetCounters() map[string]int64 {
	m.mx.RLock()
	defer m.mx.RUnlock()

	mapCopy := make(map[string]int64, len(m.Counters))
	for key, val := range m.Counters {
		mapCopy[key] = val
	}

	return mapCopy
}

func (m *Metrics) setCounterValue(metricName string, value int64) {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.Counters[metricName] = value
}
