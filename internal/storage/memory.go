package storage

var Storage *MemStorage

type MemStorage struct {
	Metrics Metrics
}

type Metrics struct {
	Counters map[string]int64
	Gauges   map[string]float64
}

func (s MemStorage) AddCounterValue(metricName string, value int64) {
	s.setCounterValue(metricName, s.GetCounterValue(metricName)+value)
}

func (s MemStorage) GetCounterValue(metricName string) int64 {
	return s.Metrics.Counters[metricName]
}

func (s MemStorage) setCounterValue(metricName string, value int64) {
	s.Metrics.Counters[metricName] = value
}

func (s MemStorage) GetGaugeValue(metricName string) float64 {
	return s.Metrics.Gauges[metricName]
}

func (s MemStorage) SetGaugeValue(metricName string, value float64) {
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
