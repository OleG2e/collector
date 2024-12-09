package storage

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/OleG2e/collector/internal/config"
)

var httpClient = http.DefaultClient
var monitor *monitorStorage
var pollCount = 0

type monitorStorage struct {
	Stats        map[string]any
	RuntimeStats *runtime.MemStats
	mx           sync.RWMutex
}

func (s *monitorStorage) refreshStats() {
	pollCount++
	runtime.ReadMemStats(s.RuntimeStats)
	s.seedGauge()
}

func (s *monitorStorage) resetPollCount() {
	pollCount = 0
}

func (s *monitorStorage) initSendTicker() {
	reportInterval := time.Duration(config.GetConfig().GetReportInterval()) * time.Second
	ticker := time.NewTicker(reportInterval)
	go func() {
		for range ticker.C {
			sendGaugeDataErr := s.sendGaugeData()
			if sendGaugeDataErr != nil {
				log.Println(sendGaugeDataErr)
			}
			sendCounterDataErr := s.sendCounterData()
			if sendCounterDataErr != nil {
				log.Println(sendCounterDataErr)
			}

			if sendGaugeDataErr == nil && sendCounterDataErr == nil {
				s.resetPollCount()
			}
		}
	}()
}

func RunMonitor() {
	initMonitor()
	pollInterval := time.Duration(config.GetConfig().GetPollInterval()) * time.Second
	for {
		monitor.refreshStats()
		time.Sleep(pollInterval)
	}
}

func (s *monitorStorage) seedGauge() {
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

func initMonitor() {
	g := make(map[string]any)
	ms := runtime.MemStats{}

	monitor = &monitorStorage{Stats: g, RuntimeStats: &ms}
	monitor.initSendTicker()
}

func (s *monitorStorage) sendGaugeData() error {
	for k, v := range s.Stats {
		hp := config.GetConfig().GetServerHostPort()
		req, reqErr := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/update/gauge/%s/%v", hp, k, v), http.NoBody)

		if req != nil {
			req.Header.Add("Content-Type", "text/plain")

			reqCloseErr := req.Body.Close()
			if reqCloseErr != nil {
				return errors.New(reqCloseErr.Error())
			}
		}

		if reqErr != nil {
			return errors.New(reqErr.Error())
		}

		resp, clientErr := httpClient.Do(req)

		if resp != nil {
			_, bodyErr := io.ReadAll(resp.Body)

			if bodyErr != nil {
				return errors.New(bodyErr.Error())
			}

			respCloseErr := resp.Body.Close()
			if respCloseErr != nil {
				return errors.New(respCloseErr.Error())
			}
		}

		if clientErr != nil {
			return errors.New(clientErr.Error())
		}
	}
	return nil
}

func (s *monitorStorage) sendCounterData() error {
	hp := config.GetConfig().GetServerHostPort()
	url := fmt.Sprintf("http://%s/update/counter/PollCount/%d", hp, pollCount)
	req, reqErr := http.NewRequest(http.MethodPost, url, http.NoBody)

	if req != nil {
		req.Header.Add("Content-Type", "text/plain")

		reqCloseErr := req.Body.Close()
		if reqCloseErr != nil {
			return errors.New(reqCloseErr.Error())
		}
	}

	if reqErr != nil {
		return errors.New(reqErr.Error())
	}

	resp, clientErr := httpClient.Do(req)

	if resp != nil {
		_, bodyErr := io.ReadAll(resp.Body)

		if bodyErr != nil {
			return errors.New(bodyErr.Error())
		}

		respCloseErr := resp.Body.Close()
		if respCloseErr != nil {
			return errors.New(respCloseErr.Error())
		}
	}

	if clientErr != nil {
		return errors.New(clientErr.Error())
	}

	return nil
}
