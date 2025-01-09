package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

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
	ctx          context.Context
}

func NewMonitor(ctx context.Context, l *logging.ZapLogger, agentConfig *config.AgentConfig) *MonitorStorage {
	g := make(map[string]any)
	ms := &runtime.MemStats{}
	mx := &sync.RWMutex{}
	httpClient := http.DefaultClient

	return &MonitorStorage{
		Stats:        g,
		RuntimeStats: ms,
		mx:           mx,
		l:            l,
		ctx:          ctx,
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

func (s *MonitorStorage) initSendTicker() {
	ticker := time.NewTicker(s.agentConfig.GetReportIntervalDuration())
	go func() {
		for range ticker.C {
			sendGaugeDataErr := s.sendGaugeData()
			if sendGaugeDataErr != nil {
				s.l.ErrorCtx(s.ctx, "send gauge error", zap.Error(sendGaugeDataErr))
			}
			sendCounterDataErr := s.sendCounterData()
			if sendCounterDataErr != nil {
				s.l.ErrorCtx(s.ctx, "send counter error", zap.Error(sendCounterDataErr))
			}

			if sendGaugeDataErr == nil && sendCounterDataErr == nil {
				s.resetPollCount()
			}
		}
	}()
}

func RunMonitor(ctx context.Context, l *logging.ZapLogger, agentConfig *config.AgentConfig) {
	mon := NewMonitor(ctx, l, agentConfig)
	mon.initSendTicker()
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

func (s *MonitorStorage) getStatForms() []network.MetricForm {
	stats := s.getStats()
	forms := make([]network.MetricForm, len(stats))
	i := 0
	for key, val := range stats {
		valConverted, convertErr := strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
		if convertErr != nil {
			s.l.ErrorCtx(s.ctx, "convert stat forms error", zap.Error(convertErr))
		}
		form := network.MetricForm{ID: key, MType: "gauge", Value: &valConverted}
		forms[i] = form
		i++
	}

	return forms
}

func (s *MonitorStorage) sendGaugeData() error {
	address := s.agentConfig.GetAddress()
	url := "http://" + address + "/update/"
	for _, form := range s.getStatForms() {
		formMarshalled, marshErr := json.Marshal(form)
		if marshErr != nil {
			return marshErr
		}

		reader := bytes.NewReader(formMarshalled)
		req, reqErr := http.NewRequestWithContext(s.ctx, http.MethodPost, url, reader)

		if reqErr != nil {
			return reqErr
		}

		req.Header.Add("Content-Type", "application/json")

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
	}

	return nil
}

func (s *MonitorStorage) getPollCount() *int64 {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return &s.pollCount
}

func (s *MonitorStorage) sendCounterData() error {
	address := s.agentConfig.GetAddress()
	url := fmt.Sprintf("http://%s/update/", address)

	form := network.MetricForm{
		ID:    "PollCount",
		MType: "counter",
		Delta: s.getPollCount(),
	}

	formMarshalled, marshErr := json.Marshal(form)
	if marshErr != nil {
		return marshErr
	}

	req, reqErr := http.NewRequestWithContext(s.ctx, http.MethodPost, url, bytes.NewReader(formMarshalled))

	if reqErr != nil {
		return reqErr
	}

	req.Header.Add("Content-Type", "application/json")

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
}
