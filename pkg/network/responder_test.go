package network

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"collector/internal/config"
	"collector/internal/core/domain"
	"collector/pkg/hashing"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestResponse_Success_SetsDefaultHeaders(t *testing.T) {
	resp := NewResponse(newTestLogger(), &config.ServerConfig{})
	rr := httptest.NewRecorder()

	resp.Success(rr)

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content-type: %q", got)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestResponse_BadRequestError_Status400(t *testing.T) {
	resp := NewResponse(newTestLogger(), &config.ServerConfig{})
	rr := httptest.NewRecorder()

	resp.BadRequestError(rr, "bad request")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestResponse_ServerError_Status500(t *testing.T) {
	resp := NewResponse(newTestLogger(), &config.ServerConfig{})
	rr := httptest.NewRecorder()

	resp.ServerError(rr, "server error")

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestResponse_Send_WritesJSON_StatusAndHeaders_NoHash(t *testing.T) {
	conf := &config.ServerConfig{}
	resp := NewResponse(newTestLogger(), conf)
	rr := httptest.NewRecorder()

	payload := map[string]any{"ok": true, "n": 42}
	resp.Send(context.Background(), rr, http.StatusCreated, payload)

	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type: %q", ct)
	}
	if got := rr.Header().Get(domain.HashHeader); got != "" {
		t.Fatalf("unexpected hash header: %q", got)
	}

	var got map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got["ok"] != true || int(got["n"].(float64)) != 42 {
		t.Fatalf("unexpected body: %v", got)
	}
}

func TestResponse_Send_AddsHashHeader_WhenKeyProvided(t *testing.T) {
	conf := &config.ServerConfig{
		BaseConfig: config.BaseConfig{HashKey: "test_key"},
	}
	resp := NewResponse(newTestLogger(), conf)
	rr := httptest.NewRecorder()

	payload := map[string]any{"msg": "hello"}
	resp.Send(context.Background(), rr, http.StatusOK, payload)

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type: %q", ct)
	}

	expectedBody, _ := json.Marshal(payload)
	wantHash := hashing.HashByKey(string(expectedBody), conf.GetHashKey())
	if got := rr.Header().Get(domain.HashHeader); got != wantHash {
		t.Fatalf("unexpected hash header: got %q, want %q", got, wantHash)
	}

	var got map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got["msg"] != "hello" {
		t.Fatalf("unexpected body: %v", got)
	}
}
