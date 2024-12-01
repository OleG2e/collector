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
	curVal, hasValue := s.GetCounterValue(metricName)
	if !hasValue {
		s.setCounterValue(metricName, value)
		return
	}

	s.setCounterValue(metricName, curVal+value)
}

func (s MemStorage) GetCounterValue(metricName string) (int64, bool) {
	v, hasValue := s.Metrics.Counters[metricName]
	return v, hasValue
}

func (s MemStorage) setCounterValue(metricName string, value int64) {
	s.Metrics.Counters[metricName] = value
}

func (s MemStorage) GetGaugeValue(metricName string) (float64, bool) {
	v, hasValue := s.Metrics.Gauges[metricName]
	return v, hasValue
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
