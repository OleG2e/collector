package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/OleG2e/collector/internal/container"
)

const storageFilename = "storage.db"

var (
	storage *MemStorage
	once    sync.Once
)

type MemStorage struct {
	Metrics Metrics
	DBFile  *os.File
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

func NewMemStorage() *MemStorage {
	ms := &MemStorage{
		Metrics: Metrics{Counters: make(map[string]int64), Gauges: make(map[string]float64)},
		mx:      &sync.RWMutex{},
	}
	ms.openDbFile()

	return ms
}

func GetStorage() *MemStorage {
	return storage
}

func InitStorage() {
	once.Do(func() {
		storage = NewMemStorage()

		storeInterval := container.GetServerConfig().GetStoreIntervalDuration()
		container.GetLogger().Debug("storeInterval", storeInterval)

		if storeInterval > 0 {
			storage.initFlushStorageTicker(storeInterval)
		}

		if container.GetServerConfig().Restore {
			storage.restoreStorage()
		}

		container.GetLogger().Debug("init storage success")
	})
}

func (ms *MemStorage) initFlushStorageTicker(storeInterval time.Duration) {
	ticker := time.NewTicker(storeInterval)
	go func() {
		for range ticker.C {
			if err := ms.FlushStorage(); err != nil {
				container.GetLogger().Errorf("flush storage error: %v", err)
			}
		}
	}()
}

func (ms *MemStorage) restoreStorage() {
	ms.mx.Lock()
	defer ms.mx.Unlock()

	reader := bufio.NewReader(ms.DBFile)
	dec := json.NewDecoder(reader)

	logger := container.GetLogger()
	var states []Metrics
	for dec.More() {
		var decodedMetric Metrics

		err := dec.Decode(&decodedMetric)

		if err != nil {
			logger.Fatal(err)
		}

		states = append(states, decodedMetric)
	}

	if len(states) > 0 {
		lastState := states[len(states)-1]

		ms.Metrics = lastState

		return
	}
	logger.Infoln("no restore metrics found")
}

func (ms *MemStorage) FlushStorage() error {
	ms.mx.Lock()
	defer ms.mx.Unlock()

	data, err := json.Marshal(&ms.Metrics)

	if err != nil {
		return err
	}

	if ms.DBFile == nil {
		ms.openDbFile()

		defer func(file *os.File) {
			fileCloseErr := file.Close()
			if fileCloseErr != nil && !errors.Is(fileCloseErr, os.ErrClosed) {
				container.GetLogger().Error(fileCloseErr)
			}
			ms.DBFile = nil
		}(ms.DBFile)
	}

	_, err = ms.DBFile.Write(data)

	container.GetLogger().Infoln("flush storage")

	if err != nil {
		return err
	}

	_, err = ms.DBFile.WriteString("\n")
	if err != nil {
		return err
	}

	return nil
}

func (ms *MemStorage) openDbFile() {
	path := container.GetServerConfig().FileStoragePath
	file, fileErr := os.OpenFile(path+"/"+storageFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)

	if fileErr != nil {
		container.GetLogger().Fatal(fileErr)
	}

	ms.DBFile = file
}
