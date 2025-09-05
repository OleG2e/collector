package logging

import (
	"context"
	"log/slog"
	"testing"

	"collector/internal/adapters/api/rest"
)

type captureHandler struct {
	enabled bool
	attrs   []slog.Attr
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return h.enabled
}

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	var acc []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		acc = append(acc, a)
		return true
	})
	h.attrs = acc
	return nil
}

func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &captureHandler{
		enabled: h.enabled,
		attrs:   append(append([]slog.Attr{}, h.attrs...), attrs...),
	}
}

func (h *captureHandler) WithGroup(name string) slog.Handler {
	return &captureHandler{enabled: h.enabled, attrs: h.attrs}
}

func findAttr(attrs []slog.Attr, key string) (slog.Attr, bool) {
	for _, a := range attrs {
		if a.Key == key {
			return a, true
		}
	}
	return slog.Attr{}, false
}

func TestRequestIDHandler_InsertsRequestID_WhenPresent(t *testing.T) {
	base := &captureHandler{enabled: true}
	h := NewRequestIDHandler(base)
	logger := slog.New(h)

	ctx := context.WithValue(context.Background(), rest.RequestIDKey, "req-123")
	logger.InfoContext(ctx, "hello", slog.String("k", "v"))

	reqKey := string(rest.RequestIDKey)

	if _, ok := findAttr(base.attrs, "k"); !ok {
		t.Fatalf("missing user attr 'k'")
	}
	attr, ok := findAttr(base.attrs, reqKey)
	if !ok {
		t.Fatalf("missing request id attr %q", reqKey)
	}
	if attr.Value.String() != "req-123" {
		t.Fatalf("unexpected request id value: %s", attr.Value.String())
	}
}

func TestRequestIDHandler_SkipsWhenNoRequestID(t *testing.T) {
	base := &captureHandler{enabled: true}
	h := NewRequestIDHandler(base)
	logger := slog.New(h)

	logger.Info("no ctx id", slog.String("x", "y"))

	reqKey := string(rest.RequestIDKey)
	if _, ok := findAttr(base.attrs, reqKey); ok {
		t.Fatalf("unexpected request id attr %q when not provided in context", reqKey)
	}
	if _, ok := findAttr(base.attrs, "x"); !ok {
		t.Fatalf("missing user attr 'x'")
	}
}
