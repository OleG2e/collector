package storage

import (
	"fmt"
	"github.com/OleG2e/collector/internal/config"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

var httpClient = http.DefaultClient
var monitor *monitorStorage
var pollCount = 0

type monitorStorage struct {
	Stats        map[string]any
	RuntimeStats *runtime.MemStats
}

func (s *monitorStorage) refreshStats() {
	pollCount++
	runtime.ReadMemStats(s.RuntimeStats)
	s.seedGauge()
}

func (s *monitorStorage) initSendTicker() {
	reportInterval := time.Duration(config.GetConfig().GetReportInterval()) * time.Second
	ticker := time.NewTicker(reportInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.sendGaugeData()
				s.sendCounterData()
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

func (s *monitorStorage) sendGaugeData() {
	for k, v := range s.Stats {
		hp := config.GetConfig().GetServerHostPort()
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/update/gauge/%s/%v", hp, k, v), http.NoBody)

		if err != nil {
			log.Println(err)
			continue
		}

		req.Header.Add("Content-Type", "text/plain")

		resp, clientErr := httpClient.Do(req)

		if clientErr != nil {
			log.Println(clientErr)
			continue
		}

		if resp != nil {
			defer func(Body io.ReadCloser) {
				respErr := Body.Close()
				if respErr != nil {
					log.Println(clientErr)
				}
			}(resp.Body)
		}
	}
}

func (s *monitorStorage) sendCounterData() {
	hp := config.GetConfig().GetServerHostPort()
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/update/counter/PollCount/%d", hp, pollCount), http.NoBody)

	if err != nil {
		log.Println(err)
	}

	req.Header.Add("Content-Type", "text/plain")

	resp, clientErr := httpClient.Do(req)

	if clientErr != nil {
		log.Println(clientErr)
	}

	if resp != nil {
		defer func(Body io.ReadCloser) {
			respErr := Body.Close()
			if respErr != nil {
				log.Println(clientErr)
			}
		}(resp.Body)
	}
}
