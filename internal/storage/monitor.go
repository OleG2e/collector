package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/OleG2e/collector/pkg/hashing"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/OleG2e/collector/pkg/retry"

	"github.com/OleG2e/collector/internal/config"
	"github.com/OleG2e/collector/pkg/logging"
	"go.uber.org/zap"

	"github.com/OleG2e/collector/internal/network"
)

type MonitorStorage struct {
	Stats        map[string]any
	pollCount    int64
	RuntimeStats *runtime.MemStats
	l            *logging.ZapLogger
	agentConfig  *config.AgentConfig
	mx           *sync.RWMutex
	httpClient   *http.Client
}

func NewMonitor(l *logging.ZapLogger, agentConfig *config.AgentConfig) *MonitorStorage {
	g := make(map[string]any)
	ms := &runtime.MemStats{}
	mx := &sync.RWMutex{}
	httpClient := http.DefaultClient

	return &MonitorStorage{
		Stats:        g,
		RuntimeStats: ms,
		mx:           mx,
		l:            l,
		httpClient:   httpClient,
		agentConfig:  agentConfig,
	}
}

func (s *MonitorStorage) refreshStats() {
	s.incrementPollCount()
	runtime.ReadMemStats(s.RuntimeStats)
	s.seedGauge()
}

func (s *MonitorStorage) resetPollCount() {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.pollCount = 0
}

func (s *MonitorStorage) incrementPollCount() {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.pollCount++
}

func (s *MonitorStorage) initSendTicker(ctx context.Context) {
	ticker := time.NewTicker(s.agentConfig.GetReportIntervalDuration())
	go func() {
		for range ticker.C {
			sendDataErr := s.sendData(ctx)
			if sendDataErr != nil {
				s.l.ErrorCtx(ctx, "send error", zap.Error(sendDataErr))
			}

			if sendDataErr == nil {
				s.resetPollCount()
			}
		}
	}()
}

func RunMonitor(ctx context.Context, l *logging.ZapLogger, agentConfig *config.AgentConfig) {
	mon := NewMonitor(l, agentConfig)
	mon.initSendTicker(ctx)
	for {
		mon.refreshStats()
		time.Sleep(mon.agentConfig.GetPollIntervalDuration())
	}
}

func (s *MonitorStorage) seedGauge() {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.Stats["Alloc"] = s.RuntimeStats.Alloc
	s.Stats["BuckHashSys"] = s.RuntimeStats.BuckHashSys
	s.Stats["Frees"] = s.RuntimeStats.Frees
	s.Stats["GCCPUFraction"] = s.RuntimeStats.GCCPUFraction
	s.Stats["GCSys"] = s.RuntimeStats.GCSys
	s.Stats["HeapAlloc"] = s.RuntimeStats.HeapAlloc
	s.Stats["HeapIdle"] = s.RuntimeStats.HeapIdle
	s.Stats["HeapInuse"] = s.RuntimeStats.HeapInuse
	s.Stats["HeapObjects"] = s.RuntimeStats.HeapObjects
	s.Stats["HeapReleased"] = s.RuntimeStats.HeapReleased
	s.Stats["HeapSys"] = s.RuntimeStats.HeapSys
	s.Stats["LastGC"] = s.RuntimeStats.LastGC
	s.Stats["Lookups"] = s.RuntimeStats.Lookups
	s.Stats["MCacheInuse"] = s.RuntimeStats.MCacheInuse
	s.Stats["MCacheSys"] = s.RuntimeStats.MCacheSys
	s.Stats["MSpanInuse"] = s.RuntimeStats.MSpanInuse
	s.Stats["MSpanSys"] = s.RuntimeStats.MSpanSys
	s.Stats["Mallocs"] = s.RuntimeStats.Mallocs
	s.Stats["NextGC"] = s.RuntimeStats.NextGC
	s.Stats["NumForcedGC"] = s.RuntimeStats.NumForcedGC
	s.Stats["NumGC"] = s.RuntimeStats.NumGC
	s.Stats["OtherSys"] = s.RuntimeStats.OtherSys
	s.Stats["PauseTotalNs"] = s.RuntimeStats.PauseTotalNs
	s.Stats["StackInuse"] = s.RuntimeStats.StackInuse
	s.Stats["StackSys"] = s.RuntimeStats.StackSys
	s.Stats["Sys"] = s.RuntimeStats.Sys
	s.Stats["TotalAlloc"] = s.RuntimeStats.TotalAlloc
	s.Stats["RandomValue"] = rand.Int63()
}

func (s *MonitorStorage) getStats() map[string]any {
	s.mx.RLock()
	defer s.mx.RUnlock()

	mapCopy := make(map[string]any, len(s.Stats))
	for key, val := range s.Stats {
		mapCopy[key] = val
	}

	return mapCopy
}

func (s *MonitorStorage) getStatForms() ([]network.MetricForm, error) {
	stats := s.getStats()
	forms := make([]network.MetricForm, len(stats))
	i := 0
	for key, val := range stats {
		valConverted, convertErr := strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
		if convertErr != nil {
			return nil, convertErr
		}
		form := network.MetricForm{ID: key, MType: network.MetricTypeGauge, Value: &valConverted}
		forms[i] = form
		i++
	}

	return forms, nil
}

func (s *MonitorStorage) getPollCountForm() network.MetricForm {
	delta := s.getPollCount()
	return network.MetricForm{ID: "PollCount", MType: network.MetricTypeCounter, Delta: &delta}
}

func (s *MonitorStorage) getPollCount() int64 {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.pollCount
}

func (s *MonitorStorage) sendData(ctx context.Context) error {
	address := s.agentConfig.GetAddress()
	url := "http://" + address + "/updates/"

	stats, statErr := s.getStatForms()

	if statErr != nil {
		return statErr
	}

	data := make([]network.MetricForm, len(stats)+1)
	data = append(data, stats...)
	data = append(data, s.getPollCountForm())

	dataMarshalled, marshErr := json.Marshal(data)
	if marshErr != nil {
		return marshErr
	}

	hash := ""
	if s.agentConfig.HasHashKey() {
		hash = hashing.HashByKey(string(dataMarshalled), s.agentConfig.GetHashKey())
	}

	tryErr := retry.Try(func() error {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(dataMarshalled))

		if reqErr != nil {
			return reqErr
		}

		req.Header.Add("Content-Type", "application/json")

		if s.agentConfig.HasHashKey() {
			req.Header.Add(network.HashHeader, hash)
		}

		reqCloseErr := req.Body.Close()
		if reqCloseErr != nil {
			return reqCloseErr
		}

		resp, clientErr := s.httpClient.Do(req)

		if clientErr != nil {
			return clientErr
		}

		_, bodyErr := io.ReadAll(resp.Body)

		if bodyErr != nil {
			return bodyErr
		}

		respCloseErr := resp.Body.Close()
		if respCloseErr != nil {
			return respCloseErr
		}

		return nil
	})

	return tryErr
}
