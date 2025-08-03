package logging

import (
	"collector/internal/adapters/api/rest"
	"context"
	"log/slog"
	"os"
)

func NewLogger(lvl slog.Level) *slog.Logger {
	return slog.New(NewRequestIDHandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	})))
}

type requestIDHandler struct {
	handler slog.Handler
}

func NewRequestIDHandler(handler slog.Handler) slog.Handler {
	return &requestIDHandler{handler: handler}
}

func (c *requestIDHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return c.handler.Enabled(ctx, level)
}

func (c *requestIDHandler) Handle(ctx context.Context, r slog.Record) error {
	if requestID, ok := ctx.Value(rest.RequestIDKey).(string); ok {
		r.AddAttrs(slog.String(string(rest.RequestIDKey), requestID))
	}

	return c.handler.Handle(ctx, r)
}

func (c *requestIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &requestIDHandler{handler: c.handler.WithAttrs(attrs)}
}

func (c *requestIDHandler) WithGroup(name string) slog.Handler {
	return &requestIDHandler{handler: c.handler.WithGroup(name)}
}
