package network

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewRequestInfo_ReadsFieldsAndPreservesBody(t *testing.T) {
	origBody := `{"x":1,"y":"z"}`
	req, err := http.NewRequest(
		http.MethodPost,
		"http://example.com/api/v1/test?ok=1",
		io.NopCloser(strings.NewReader(origBody)),
	)
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}

	ri := NewRequestInfo(req)

	if ri.Method != http.MethodPost {
		t.Fatalf("unexpected method: %s", ri.Method)
	}
	if ri.URL != "http://example.com/api/v1/test?ok=1" {
		t.Fatalf("unexpected url: %s", ri.URL)
	}
	if ri.Body != origBody {
		t.Fatalf("unexpected body captured: %q", ri.Body)
	}

	b2, _ := io.ReadAll(req.Body)
	if string(b2) != origBody {
		t.Fatalf("request body was not preserved, got: %q", string(b2))
	}
}

func TestRequestInfo_StringFormat(t *testing.T) {
	req, err := http.NewRequest(
		http.MethodGet,
		"http://localhost/status",
		io.NopCloser(strings.NewReader("hello")),
	)
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}

	ri := NewRequestInfo(req)
	got := ri.String()

	wantPrefix := "GET http://localhost/status "
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("unexpected prefix in String(): %q", got)
	}
	if !strings.HasSuffix(got, "hello") {
		t.Fatalf("unexpected suffix (body) in String(): %q", got)
	}
}
