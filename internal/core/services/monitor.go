package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"collector/internal/config"
	"collector/internal/core/domain"
	"collector/pkg/hashing"
	"collector/pkg/retry"
	"golang.org/x/sync/errgroup"
)

type Monitor struct {
	memStats     sync.Map
	pollCount    atomic.Int64
	runtimeStats *runtime.MemStats
	logger       *slog.Logger
	agentConfig  *config.AgentConfig
	mx           *sync.RWMutex
	httpClient   *http.Client
}

func NewMonitor(logger *slog.Logger, agentConfig *config.AgentConfig) *Monitor {
	return &Monitor{
		mx:           new(sync.RWMutex),
		runtimeStats: new(runtime.MemStats),
		logger:       logger,
		httpClient:   http.DefaultClient,
		agentConfig:  agentConfig,
	}
}

func (s *Monitor) Run(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	s.initSendTicker(gCtx, g)
	s.initRefreshStatsTicker(gCtx, g)

	err := g.Wait()
	if err != nil {
		return fmt.Errorf("monitor run error: %w", err)
	}

	return nil
}

func (s *Monitor) refreshStats() {
	s.incrementPollCount()
	runtime.ReadMemStats(s.runtimeStats)
	s.seedMemStats()
}

func (s *Monitor) resetPollCount() {
	s.pollCount.Store(0)
}

func (s *Monitor) incrementPollCount() {
	s.pollCount.Add(1)
}

func (s *Monitor) initSendTicker(ctx context.Context, g *errgroup.Group) {
	ticker := time.NewTicker(s.agentConfig.GetReportIntervalDuration())

	g.Go(func() error {
		for {
			select {
			case <-ticker.C:
				sendDataErr := s.sendData(ctx)
				if sendDataErr != nil {
					return fmt.Errorf("send stats ticker data error: %w", sendDataErr)
				}

				s.resetPollCount()
			case <-ctx.Done():
				if ctx.Err() != nil {
					return fmt.Errorf("send stats ticker ctx error: %w", ctx.Err())
				}

				return nil
			}
		}
	})
}

func (s *Monitor) initRefreshStatsTicker(ctx context.Context, g *errgroup.Group) {
	ticker := time.NewTicker(s.agentConfig.GetPollIntervalDuration())

	g.Go(func() error {
		for {
			select {
			case <-ticker.C:
				s.refreshStats()
			case <-ctx.Done():
				if ctx.Err() != nil {
					return fmt.Errorf("poll stats ticker ctx error: %w", ctx.Err())
				}

				return nil
			}
		}
	})
}

func (s *Monitor) seedMemStats() {
	s.memStats.Store("Alloc", s.runtimeStats.Alloc)
	s.memStats.Store("BuckHashSys", s.runtimeStats.BuckHashSys)
	s.memStats.Store("Frees", s.runtimeStats.Frees)
	s.memStats.Store("GCCPUFraction", s.runtimeStats.GCCPUFraction)
	s.memStats.Store("GCSys", s.runtimeStats.GCSys)
	s.memStats.Store("HeapAlloc", s.runtimeStats.HeapAlloc)
	s.memStats.Store("HeapIdle", s.runtimeStats.HeapIdle)
	s.memStats.Store("HeapInuse", s.runtimeStats.HeapInuse)
	s.memStats.Store("HeapObjects", s.runtimeStats.HeapObjects)
	s.memStats.Store("HeapReleased", s.runtimeStats.HeapReleased)
	s.memStats.Store("HeapSys", s.runtimeStats.HeapSys)
	s.memStats.Store("LastGC", s.runtimeStats.LastGC)
	s.memStats.Store("Lookups", s.runtimeStats.Lookups)
	s.memStats.Store("MCacheInuse", s.runtimeStats.MCacheInuse)
	s.memStats.Store("MCacheSys", s.runtimeStats.MCacheSys)
	s.memStats.Store("MSpanInuse", s.runtimeStats.MSpanInuse)
	s.memStats.Store("MSpanSys", s.runtimeStats.MSpanSys)
	s.memStats.Store("Mallocs", s.runtimeStats.Mallocs)
	s.memStats.Store("NextGC", s.runtimeStats.NextGC)
	s.memStats.Store("NumForcedGC", s.runtimeStats.NumForcedGC)
	s.memStats.Store("NumGC", s.runtimeStats.NumGC)
	s.memStats.Store("OtherSys", s.runtimeStats.OtherSys)
	s.memStats.Store("PauseTotalNs", s.runtimeStats.PauseTotalNs)
	s.memStats.Store("StackInuse", s.runtimeStats.StackInuse)
	s.memStats.Store("StackSys", s.runtimeStats.StackSys)
	s.memStats.Store("Sys", s.runtimeStats.Sys)
	s.memStats.Store("TotalAlloc", s.runtimeStats.TotalAlloc)
	s.memStats.Store("RandomValue", rand.Int63())
}

func (s *Monitor) getStatForms() []*domain.MetricForm {
	var forms []*domain.MetricForm

	s.memStats.Range(func(key, value any) bool {
		valConverted, _ := strconv.ParseFloat(fmt.Sprintf("%v", value), 64)
		form := &domain.MetricForm{
			ID:    key.(string),
			MType: domain.MetricTypeGauge,
			Value: &valConverted,
		}
		forms = append(forms, form)

		return true
	})

	return forms
}

func (s *Monitor) getPollCountForm() *domain.MetricForm {
	delta := s.getPollCount()

	return &domain.MetricForm{ID: "PollCount", MType: domain.MetricTypeCounter, Delta: &delta}
}

func (s *Monitor) getPollCount() int64 {
	return s.pollCount.Load()
}

func (s *Monitor) sendData(ctx context.Context) error {
	address := s.agentConfig.GetAddress()
	url := "http://" + address + "/updates/"

	stats := s.getStatForms()

	data := make([]*domain.MetricForm, len(stats)+1)
	data = append(data, stats...)
	data = append(data, s.getPollCountForm())

	dataMarshalled, marshErr := json.Marshal(data)
	if marshErr != nil {
		return fmt.Errorf("marshall data error: %w", marshErr)
	}

	hash := ""
	if s.agentConfig.HasHashKey() {
		hash = hashing.HashByKey(string(dataMarshalled), s.agentConfig.GetHashKey())
	}

	tryErr := retry.Try(func() error {
		req, reqErr := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			url,
			bytes.NewReader(dataMarshalled),
		)

		if reqErr != nil {
			return fmt.Errorf("request error: %w", reqErr)
		}

		req.Header.Add("Content-Type", "application/json")

		if s.agentConfig.HasHashKey() {
			req.Header.Add(domain.HashHeader, hash)
		}

		reqCloseErr := req.Body.Close()
		if reqCloseErr != nil {
			return fmt.Errorf("request close error: %w", reqCloseErr)
		}

		resp, respErr := s.httpClient.Do(req)

		if respErr != nil {
			return fmt.Errorf("response error: %w", respErr)
		}

		_, bodyErr := io.ReadAll(resp.Body)

		if bodyErr != nil {
			return fmt.Errorf("read body error: %w", bodyErr)
		}

		respCloseErr := resp.Body.Close()
		if respCloseErr != nil {
			return fmt.Errorf("response close body error: %w", respCloseErr)
		}

		return nil
	})

	return tryErr
}
