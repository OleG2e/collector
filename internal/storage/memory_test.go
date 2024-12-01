package storage

import (
	"reflect"
	"testing"
)

func TestGetStorage(t *testing.T) {
	tests := []struct {
		name string
		want *MemStorage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_AddCounterValue(t *testing.T) {
	type fields struct {
		Metrics Metrics
	}
	type args struct {
		metricName string
		value      int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := MemStorage{
				Metrics: tt.fields.Metrics,
			}
			s.AddCounterValue(tt.args.metricName, tt.args.value)
		})
	}
}

func TestMemStorage_GetCounterValue(t *testing.T) {
	type fields struct {
		Metrics Metrics
	}
	type args struct {
		metricName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int64
		want1  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := MemStorage{
				Metrics: tt.fields.Metrics,
			}
			got, got1 := s.GetCounterValue(tt.args.metricName)
			if got != tt.want {
				t.Errorf("GetCounterValue() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetCounterValue() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestMemStorage_GetGaugeValue(t *testing.T) {
	type fields struct {
		Metrics Metrics
	}
	type args struct {
		metricName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
		want1  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := MemStorage{
				Metrics: tt.fields.Metrics,
			}
			got, got1 := s.GetGaugeValue(tt.args.metricName)
			if got != tt.want {
				t.Errorf("GetGaugeValue() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetGaugeValue() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestMemStorage_SetGaugeValue(t *testing.T) {
	type fields struct {
		Metrics Metrics
	}
	type args struct {
		metricName string
		value      float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := MemStorage{
				Metrics: tt.fields.Metrics,
			}
			s.SetGaugeValue(tt.args.metricName, tt.args.value)
		})
	}
}

func TestMemStorage_setCounterValue(t *testing.T) {
	type fields struct {
		Metrics Metrics
	}
	type args struct {
		metricName string
		value      int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := MemStorage{
				Metrics: tt.fields.Metrics,
			}
			s.setCounterValue(tt.args.metricName, tt.args.value)
		})
	}
}

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want MemStorage
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}
