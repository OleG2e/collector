package services

import (
	"log/slog"
	"sync"
	"testing"
)

func newBenchLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(ioDiscard{}, nil))
}

func BenchmarkMonitor_getStatForms_SmallMap(b *testing.B) {
	m := &Monitor{
		logger: newBenchLogger(),
		mx:     new(sync.RWMutex),
	}

	m.memStats.Store("Alloc", 100)
	m.memStats.Store("RandomValue", 42)
	m.memStats.Store("CPUutilization1", 12.34)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.getStatForms()
	}
}

func BenchmarkMonitor_PollCount(b *testing.B) {
	m := &Monitor{
		logger: newBenchLogger(),
		mx:     new(sync.RWMutex),
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.incrementPollCount()
		_ = m.getPollCount()
	}
}

func BenchmarkMonitor_getPollCountForm(b *testing.B) {
	m := &Monitor{
		logger: newBenchLogger(),
		mx:     new(sync.RWMutex),
	}
	for i := 0; i < 1000; i++ {
		m.incrementPollCount()
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = m.getPollCountForm()
	}
}
