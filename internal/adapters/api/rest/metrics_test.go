package rest

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"collector/internal/config"
	"collector/internal/core/domain"
	"collector/pkg/network"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func newTestResponder() *network.Response {
	return network.NewResponse(newTestLogger(), &config.ServerConfig{})
}

func TestHasAllowedMetricByURLPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"contains gauge", "/update/gauge/Alloc/123", true},
		{"contains counter", "/update/counter/PollCount/1", true},
		{"no allowed type", "/update/metric/something", false},
		{"empty path", "", false},
		{"gauge as part of another word", "/update/supergaugeX", true},
		{"counter as part of another word", "/api/v1/counterparty", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := hasAllowedMetricByURLPath(tt.path); got != tt.want {
				t.Fatalf("hasAllowedMetricByURLPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestAllowedMetricsOnly_ByPath_AllowsNext(t *testing.T) {
	resp := newTestResponder()
	logger := newTestLogger()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	mw := AllowedMetricsOnly(resp, logger)(next)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/some", nil)
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	if !nextCalled {
		t.Fatal("expected next handler to be called when path contains allowed metric type")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestAllowedMetricsOnly_DecodeError_BadRequest(t *testing.T) {
	resp := newTestResponder()
	logger := newTestLogger()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not be called on decode error")
	})

	mw := AllowedMetricsOnly(resp, logger)(next)

	req := httptest.NewRequest(http.MethodPost, "/update/metric", bytes.NewBufferString("not-json"))
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestAllowedMetricsOnly_AllowedByBody_Gauge(t *testing.T) {
	resp := newTestResponder()
	logger := newTestLogger()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	mw := AllowedMetricsOnly(resp, logger)(next)

	body := `{"id":"HeapAlloc","type":"` + string(domain.MetricTypeGauge) + `","value":123.45}`
	req := httptest.NewRequest(http.MethodPost, "/update/metric", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	if !nextCalled {
		t.Fatal("expected next handler to be called when body type is gauge")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestAllowedMetricsOnly_AllowedByBody_Counter(t *testing.T) {
	resp := newTestResponder()
	logger := newTestLogger()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	mw := AllowedMetricsOnly(resp, logger)(next)

	body := `{"id":"PollCount","type":"` + string(domain.MetricTypeCounter) + `","delta":10}`
	req := httptest.NewRequest(http.MethodPost, "/update/metric", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	if !nextCalled {
		t.Fatal("expected next handler to be called when body type is counter")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestAllowedMetricsOnly_NotAllowedType_BadRequest(t *testing.T) {
	resp := newTestResponder()
	logger := newTestLogger()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not be called for not allowed metric type")
	})

	mw := AllowedMetricsOnly(resp, logger)(next)

	body := `{"id":"Something","type":"other"}`
	req := httptest.NewRequest(http.MethodPost, "/update/metric", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
}
