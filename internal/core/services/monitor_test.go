package services

import (
	"log/slog"
	"sync"
	"testing"

	"collector/internal/core/domain"
)

func newNullLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(ioDiscard{}, nil))
}

func TestMonitor_PollCount(t *testing.T) {
	m := &Monitor{
		logger: newNullLogger(),
		mx:     new(sync.RWMutex),
	}
	if got := m.getPollCount(); got != 0 {
		t.Fatalf("unexpected initial poll count: %d", got)
	}
	m.incrementPollCount()
	m.incrementPollCount()
	if got := m.getPollCount(); got != 2 {
		t.Fatalf("unexpected poll count after increment: %d", got)
	}
	m.resetPollCount()
	if got := m.getPollCount(); got != 0 {
		t.Fatalf("unexpected poll count after reset: %d", got)
	}
}

func TestMonitor_getPollCountForm(t *testing.T) {
	m := &Monitor{
		logger: newNullLogger(),
		mx:     new(sync.RWMutex),
	}
	m.incrementPollCount()
	form := m.getPollCountForm()

	if form.ID != "PollCount" {
		t.Fatalf("unexpected id: %s", form.ID)
	}
	if !form.IsCounterType() {
		t.Fatalf("expected counter type")
	}
	if form.Delta == nil || *form.Delta != 1 {
		t.Fatalf("unexpected delta: %v", form.Delta)
	}
}

func TestMonitor_getStatForms_FromSeededMap(t *testing.T) {
	m := &Monitor{
		logger: newNullLogger(),
		mx:     new(sync.RWMutex),
	}
	m.memStats.Store("Alloc", int64(100))
	m.memStats.Store("RandomValue", int64(42))
	m.memStats.Store("CPUutilization1", float64(12.34))

	forms := m.getStatForms()
	got := map[string]*domain.MetricForm{}
	for _, f := range forms {
		got[f.ID] = f
	}

	if f, ok := got["Alloc"]; !ok || !f.IsGaugeType() || f.Value == nil || *f.Value != 100 {
		t.Fatalf("unexpected Alloc form: %+v", f)
	}
	if f, ok := got["RandomValue"]; !ok || !f.IsGaugeType() || f.Value == nil || *f.Value != 42 {
		t.Fatalf("unexpected RandomValue form: %+v", f)
	}
	if f, ok := got["CPUutilization1"]; !ok || !f.IsGaugeType() || f.Value == nil ||
		*f.Value != 12.34 {
		t.Fatalf("unexpected CPUutilization1 form: %+v", f)
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
