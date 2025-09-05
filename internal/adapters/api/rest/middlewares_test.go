package rest

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"collector/internal/config"
	"collector/internal/core/domain"
	"collector/pkg/hashing"
)

func TestIsSupportedContentType(t *testing.T) {
	tests := []struct {
		name string
		ct   string
		want bool
	}{
		{"json", "application/json; charset=utf-8", true},
		{"html", "text/html", true},
		{"empty -> default html", "", true},
		{"xml", "application/xml", false},
		{"plain", "text/plain", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := isSupportedContentType(tt.ct); got != tt.want {
				t.Fatalf("isSupportedContentType(%q) = %v, want %v", tt.ct, got, tt.want)
			}
		})
	}
}

func TestLoggerMiddleware_PassesThroughAndLogs(t *testing.T) {
	logger := newTestLogger()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})

	mw := LoggerMiddleware(logger)(next)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if rr.Body.String() != "ok" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}

func TestGzipMiddleware_InboundGzipIsDecompressed(t *testing.T) {
	logger := newTestLogger()

	const payload = `{"a":1}`
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write([]byte(payload))
	_ = gz.Close()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		b, _ := io.ReadAll(r.Body)
		if string(b) != payload {
			t.Fatalf("unexpected decompressed body: %q", string(b))
		}
		w.WriteHeader(http.StatusOK)
	})

	mw := GzipMiddleware(logger)(next)

	req := httptest.NewRequest(http.MethodPost, "/in", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestGzipMiddleware_OutboundCompressedWhenAcceptedAndSupported(t *testing.T) {
	logger := newTestLogger()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	mw := GzipMiddleware(logger)(next)

	req := httptest.NewRequest(http.MethodGet, "/out", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip encoding")
	}

	gr, err := gzip.NewReader(bytes.NewReader(rr.Body.Bytes()))
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()
	data, _ := io.ReadAll(gr)
	if got := string(data); got != `{"ok":true}` {
		t.Fatalf("unexpected decompressed body: %q", got)
	}
}

func TestGzipMiddleware_NoCompressionWhenNotAccepted(t *testing.T) {
	logger := newTestLogger()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("x"))
	})
	mw := GzipMiddleware(logger)(next)

	req := httptest.NewRequest(http.MethodGet, "/nozip", nil)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Fatalf("did not expect gzip encoding")
	}
	if rr.Body.String() != "x" {
		t.Fatalf("unexpected body: %q", rr.Body.String())
	}
}

func TestRecoverMiddleware_HandlesPanic(t *testing.T) {
	logger := newTestLogger()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	mw := RecoverMiddleware(logger)(next)

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

func TestCheckSignMiddleware_PassesWhenNoKeyOrNoHeader(t *testing.T) {
	logger := newTestLogger()
	conf := &config.ServerConfig{}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	mw := CheckSignMiddleware(conf, logger)(next)

	req := httptest.NewRequest(http.MethodPost, "/sign", strings.NewReader(`{"x":1}`))
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}

	conf.BaseConfig.HashKey = "k"
	req2 := httptest.NewRequest(http.MethodPost, "/sign", strings.NewReader(`{"x":1}`))
	rr2 := httptest.NewRecorder()
	mw.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusAccepted {
		t.Fatalf("expected 202 when no header, got %d", rr2.Code)
	}
}

func TestCheckSignMiddleware_RejectsOnWrongHash_AcceptsOnCorrect(t *testing.T) {
	logger := newTestLogger()
	conf := &config.ServerConfig{BaseConfig: config.BaseConfig{HashKey: "secret"}}

	body := `{"n":42}`
	correct := hashing.HashByKey(body, conf.GetHashKey())
	wrong := "deadbeef"

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		if string(data) != body {
			t.Fatalf("body not preserved: %q", string(data))
		}
		w.WriteHeader(http.StatusOK)
	})

	mw := CheckSignMiddleware(conf, logger)(next)

	reqBad := httptest.NewRequest(http.MethodPost, "/sign", strings.NewReader(body))
	reqBad.Header.Set(domain.HashHeader, wrong)
	rrBad := httptest.NewRecorder()
	mw.ServeHTTP(rrBad, reqBad)
	if rrBad.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for wrong hash, got %d", rrBad.Code)
	}

	reqOk := httptest.NewRequest(http.MethodPost, "/sign", strings.NewReader(body))
	reqOk.Header.Set(domain.HashHeader, correct)
	rrOk := httptest.NewRecorder()
	mw.ServeHTTP(rrOk, reqOk)
	if rrOk.Code != http.StatusOK {
		t.Fatalf("expected 200 for correct hash, got %d", rrOk.Code)
	}
}

func TestRequestIDMiddleware_SetsIDInContext(t *testing.T) {
	captured := ""

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v, ok := r.Context().Value(RequestIDKey).(string); ok {
			captured = v
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mw := RequestIDMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/rid", nil)
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if captured == "" {
		t.Fatalf("request id was not set in context")
	}
}

func TestLoggerMiddleware_DurationFieldPresent(t *testing.T) {
	logger := newTestLogger()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	mw := LoggerMiddleware(logger)(next)

	req := httptest.NewRequest(http.MethodGet, "/sleep", nil)
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}
