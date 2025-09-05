package domain

import (
	"sync"
	"testing"
)

func TestMetrics_GaugeSetGet(t *testing.T) {
	m := NewMetrics()
	m.SetGaugeValue("X", 1.5)

	if got, ok := m.GetGaugeValue("X"); !ok || got != 1.5 {
		t.Fatalf("unexpected gauge: %v %v", got, ok)
	}
	if _, ok := m.GetGaugeValue("Y"); ok {
		t.Fatalf("unexpected gauge found for Y")
	}
}

func TestMetrics_CounterAddGet(t *testing.T) {
	m := NewMetrics()
	m.AddCounterValue("C", 10)
	m.AddCounterValue("C", 5)

	if got, ok := m.GetCounterValue("C"); !ok || got != 15 {
		t.Fatalf("unexpected counter: %v %v", got, ok)
	}
	if _, ok := m.GetCounterValue("Z"); ok {
		t.Fatalf("unexpected counter found for Z")
	}
}

func TestMetrics_GetCopies(t *testing.T) {
	m := NewMetrics()
	m.SetGaugeValue("G", 2.2)
	m.AddCounterValue("C", 3)

	g := m.GetGauges()
	c := m.GetCounters()

	g["G"] = 100
	c["C"] = 200

	if got, _ := m.GetGaugeValue("G"); got != 2.2 {
		t.Fatalf("original gauge mutated: %v", got)
	}
	if got, _ := m.GetCounterValue("C"); got != 3 {
		t.Fatalf("original counter mutated: %v", got)
	}
}

func TestMetrics_ConcurrentAccess(t *testing.T) {
	m := NewMetrics()
	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.SetGaugeValue("G", float64(i))
			m.AddCounterValue("C", int64(1))
		}(i)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.GetGaugeValue("G")
			m.GetCounterValue("C")
			m.GetGauges()
			m.GetCounters()
		}()
	}

	wg.Wait()
}
