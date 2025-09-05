package services

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkBuildRequest(b *testing.B) {
	ctx := context.Background()
	url := "http://localhost:8080/update/gauge/X/1.23"
	data := []byte(`{"id":"X","type":"gauge","value":1.23}`)
	hash := "abc123deadbeef"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, err := buildRequest(ctx, url, data, hash)
		if err != nil {
			b.Fatalf("buildRequest error: %v", err)
		}
		if req.Body != nil {
			_, _ = io.ReadAll(req.Body)
			_ = req.Body.Close()
		}
	}
}

func BenchmarkSendRequest(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`OK`))
	}))
	defer srv.Close()

	client := srv.Client()
	ctx := context.Background()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		srv.URL+"/update/metric",
		http.NoBody,
	)
	if err != nil {
		b.Fatalf("request create error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Клонируем запрос с новым контекстом, чтобы не переиспользовать Body/Cancel
		r := req.Clone(ctx)
		res, err := sendRequest(client, r)
		if err != nil {
			b.Fatalf("sendRequest error: %v", err)
		}
		if res.Code != http.StatusOK {
			b.Fatalf("unexpected code: %d", res.Code)
		}
	}
}
