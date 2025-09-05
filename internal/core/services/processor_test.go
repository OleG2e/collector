package services

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"collector/internal/core/domain"
)

func TestBuildRequest_HeadersAndMethod(t *testing.T) {
	data := []byte(`{"id":"A"}`)
	req, err := buildRequest(context.Background(), "http://example.com/x", data, "hash123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Method != http.MethodPost {
		t.Fatalf("unexpected method: %s", req.Method)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content-type: %s", got)
	}
	if got := req.Header.Get(domain.HashHeader); got != "hash123" {
		t.Fatalf("unexpected hash header: %s", got)
	}
	b, _ := io.ReadAll(req.Body)
	if string(b) != string(data) {
		t.Fatalf("unexpected body: %s", string(b))
	}
}

func TestBuildRequest_NoHash(t *testing.T) {
	data := []byte(`{"id":"B"}`)
	req, err := buildRequest(context.Background(), "http://example.com/y", data, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := req.Header.Get(domain.HashHeader); got != "" {
		t.Fatalf("unexpected hash header: %s", got)
	}
}

func TestSendRequest_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"id"`) {
			t.Fatalf("unexpected body: %s", string(body))
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL, strings.NewReader(`{"id":"X"}`))
	client := srv.Client()

	res, err := sendRequest(client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Code != http.StatusCreated {
		t.Fatalf("unexpected code: %d", res.Code)
	}
	if !strings.Contains(res.Status, "201") {
		t.Fatalf("unexpected status: %s", res.Status)
	}
}

func TestSendRequest_DoError(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:1", nil)
	client := &http.Client{}

	_, err := sendRequest(client, req)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
