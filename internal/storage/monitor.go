package storage

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

var httpClient = http.DefaultClient
var monitor *monitorStorage
var pollCount = 0
var pollInterval = time.Duration(2) * time.Second
var reportInterval = time.Duration(10) * time.Second

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
	ticker := time.NewTicker(reportInterval)
	go func() {
		for range ticker.C {
			s.sendGaugeData()
			s.sendCounterData()
		}
	}()
}

func RunMonitor() {
	initMonitor()
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
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/update/gauge/%s/%v", k, v), http.NoBody)

		if err != nil {
			log.Println(err)
			continue
		}

		req.Header.Add("Content-Type", "text/plain")

		resp, clientErr := httpClient.Do(req)

		closeErr := resp.Body.Close()
		if closeErr != nil {
			log.Println(clientErr)
			continue
		}

		if clientErr != nil {
			log.Println(clientErr)
			continue
		}
	}
}

func (s *monitorStorage) sendCounterData() {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:8080/update/counter/PollCount/%d", pollCount), http.NoBody)

	if err != nil {
		log.Println(err)
	}

	req.Header.Add("Content-Type", "text/plain")

	resp, clientErr := httpClient.Do(req)

	closeErr := resp.Body.Close()
	if closeErr != nil {
		log.Println(clientErr)
		return
	}

	if clientErr != nil {
		log.Println(clientErr)
	}
}
