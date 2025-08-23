package services

import (
	"context"
	"fmt"
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
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"golang.org/x/sync/errgroup"
)

type Monitor struct {
	memStats    sync.Map
	pollCount   atomic.Int64
	logger      *slog.Logger
	agentConfig *config.AgentConfig
	mx          *sync.RWMutex
	httpClient  *http.Client
}

func NewMonitor(logger *slog.Logger, agentConfig *config.AgentConfig) *Monitor {
	return &Monitor{
		mx:          new(sync.RWMutex),
		logger:      logger,
		httpClient:  http.DefaultClient,
		agentConfig: agentConfig,
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

	runtimeStats := new(runtime.MemStats)
	runtime.ReadMemStats(runtimeStats)
	s.seedMemStats(runtimeStats)

	extraStats, err := mem.VirtualMemory()
	if err != nil {
		s.logger.Error("failed to get memory extraStats", slog.Any("error", err))
		extraStats = new(mem.VirtualMemoryStat)
	}
	s.seedExtraStats(extraStats)
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
				stats := append(s.getStatForms(), s.getPollCountForm())
				sendDataErr := sendData(ctx, s.httpClient, s.agentConfig, stats)
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

func (s *Monitor) seedMemStats(runtimeStats *runtime.MemStats) {
	s.memStats.Store("Alloc", runtimeStats.Alloc)
	s.memStats.Store("BuckHashSys", runtimeStats.BuckHashSys)
	s.memStats.Store("Frees", runtimeStats.Frees)
	s.memStats.Store("GCCPUFraction", runtimeStats.GCCPUFraction)
	s.memStats.Store("GCSys", runtimeStats.GCSys)
	s.memStats.Store("HeapAlloc", runtimeStats.HeapAlloc)
	s.memStats.Store("HeapIdle", runtimeStats.HeapIdle)
	s.memStats.Store("HeapInuse", runtimeStats.HeapInuse)
	s.memStats.Store("HeapObjects", runtimeStats.HeapObjects)
	s.memStats.Store("HeapReleased", runtimeStats.HeapReleased)
	s.memStats.Store("HeapSys", runtimeStats.HeapSys)
	s.memStats.Store("LastGC", runtimeStats.LastGC)
	s.memStats.Store("Lookups", runtimeStats.Lookups)
	s.memStats.Store("MCacheInuse", runtimeStats.MCacheInuse)
	s.memStats.Store("MCacheSys", runtimeStats.MCacheSys)
	s.memStats.Store("MSpanInuse", runtimeStats.MSpanInuse)
	s.memStats.Store("MSpanSys", runtimeStats.MSpanSys)
	s.memStats.Store("Mallocs", runtimeStats.Mallocs)
	s.memStats.Store("NextGC", runtimeStats.NextGC)
	s.memStats.Store("NumForcedGC", runtimeStats.NumForcedGC)
	s.memStats.Store("NumGC", runtimeStats.NumGC)
	s.memStats.Store("OtherSys", runtimeStats.OtherSys)
	s.memStats.Store("PauseTotalNs", runtimeStats.PauseTotalNs)
	s.memStats.Store("StackInuse", runtimeStats.StackInuse)
	s.memStats.Store("StackSys", runtimeStats.StackSys)
	s.memStats.Store("Sys", runtimeStats.Sys)
	s.memStats.Store("TotalAlloc", runtimeStats.TotalAlloc)
	s.memStats.Store("RandomValue", rand.Int63())
}

func (s *Monitor) seedExtraStats(extraStat *mem.VirtualMemoryStat) {
	s.memStats.Store("TotalMemory", extraStat.Total)
	s.memStats.Store("FreeMemory", extraStat.Free)

	var firstCoreUtilization float64
	percentages, cpuPercentErr := cpu.Percent(0, true)
	if cpuPercentErr != nil || len(percentages) == 0 {
		firstCoreUtilization = 0
	}
	if len(percentages) > 0 {
		firstCoreUtilization = percentages[0]
	}

	s.memStats.Store("CPUutilization1", firstCoreUtilization)
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
