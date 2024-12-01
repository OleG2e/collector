package storage

import (
	"runtime"
	"testing"
)

func TestRunMonitor(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RunMonitor()
		})
	}
}

func Test_initMonitor(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMonitor()
		})
	}
}

func Test_monitorStorage_initSendTicker(t *testing.T) {
	type fields struct {
		Stats        map[string]any
		RuntimeStats *runtime.MemStats
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &monitorStorage{
				Stats:        tt.fields.Stats,
				RuntimeStats: tt.fields.RuntimeStats,
			}
			s.initSendTicker()
		})
	}
}

func Test_monitorStorage_refreshStats(t *testing.T) {
	type fields struct {
		Stats        map[string]any
		RuntimeStats *runtime.MemStats
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &monitorStorage{
				Stats:        tt.fields.Stats,
				RuntimeStats: tt.fields.RuntimeStats,
			}
			s.refreshStats()
		})
	}
}

func Test_monitorStorage_seedGauge(t *testing.T) {
	type fields struct {
		Stats        map[string]any
		RuntimeStats *runtime.MemStats
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &monitorStorage{
				Stats:        tt.fields.Stats,
				RuntimeStats: tt.fields.RuntimeStats,
			}
			s.seedGauge()
		})
	}
}

func Test_monitorStorage_sendCounterData(t *testing.T) {
	type fields struct {
		Stats        map[string]any
		RuntimeStats *runtime.MemStats
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &monitorStorage{
				Stats:        tt.fields.Stats,
				RuntimeStats: tt.fields.RuntimeStats,
			}
			s.sendCounterData()
		})
	}
}

func Test_monitorStorage_sendGaugeData(t *testing.T) {
	type fields struct {
		Stats        map[string]any
		RuntimeStats *runtime.MemStats
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &monitorStorage{
				Stats:        tt.fields.Stats,
				RuntimeStats: tt.fields.RuntimeStats,
			}
			s.sendGaugeData()
		})
	}
}
