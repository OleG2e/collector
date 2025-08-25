package rest

import (
	"log/slog"
	"net/http"
	"testing"

	"collector/pkg/network"
)

func TestAllowedMetricsOnly(t *testing.T) {
	type args struct {
		resp   *network.Response
		logger *slog.Logger
	}
	tests := []struct {
		name string
		args args
		want func(next http.Handler) http.Handler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Add checks
		})
	}
}

func Test_hasAllowedMetricByURLPath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasAllowedMetricByURLPath(tt.args.path); got != tt.want {
				t.Errorf("hasAllowedMetricByURLPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
