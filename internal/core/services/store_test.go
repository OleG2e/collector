package services

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"collector/internal/adapters/store"
	"collector/internal/core/domain"
)

type fakeStore struct {
	restoreErr error
	closeErr   error
	saveErr    error
	restoreCnt int
	closeCnt   int
	saveCnt    int
	metrics    *domain.Metrics
}

func (f *fakeStore) Restore(_ context.Context) error { f.restoreCnt++; return f.restoreErr }
func (f *fakeStore) Close() error                    { f.closeCnt++; return f.closeErr }
func (f *fakeStore) Save(_ context.Context) error    { f.saveCnt++; return f.saveErr }
func (f *fakeStore) GetMetrics() *domain.Metrics {
	if f.metrics == nil {
		f.metrics = domain.NewMetrics()
	}
	return f.metrics
}
func (f *fakeStore) SetMetrics(m *domain.Metrics) { f.metrics = m }
func (f *fakeStore) GetStoreType() store.Type     { return "fake" }

func TestStoreService_Restore_Save_Close_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fs := &fakeStore{}
	ss := &StoreService{logger: logger, store: fs}

	if err := ss.Restore(context.Background()); err != nil {
		t.Fatalf("unexpected restore error: %v", err)
	}
	if err := ss.Save(context.Background()); err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}
	if err := ss.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}

	if fs.restoreCnt == 0 || fs.saveCnt == 0 || fs.closeCnt == 0 {
		t.Fatalf("store methods were not invoked: %+v", fs)
	}
}

func TestStoreService_ErrorsPropagate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fs := &fakeStore{
		restoreErr: errors.New("restore fail"),
		saveErr:    errors.New("save fail"),
		closeErr:   errors.New("close fail"),
	}
	ss := &StoreService{logger: logger, store: fs}

	if err := ss.Restore(context.Background()); err == nil {
		t.Fatalf("expected restore error")
	}
	if err := ss.Save(context.Background()); err == nil {
		t.Fatalf("expected save error")
	}
	if err := ss.Close(); err == nil {
		t.Fatalf("expected close error")
	}
}

func TestStoreService_InitFlushStorageTicker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fs := &fakeStore{}
	ss := &StoreService{logger: logger, store: fs}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ss.InitFlushStorageTicker(ctx, 10*time.Millisecond)
	time.Sleep(25 * time.Millisecond)
	cancel()
	time.Sleep(10 * time.Millisecond)

	if fs.saveCnt == 0 {
		t.Fatalf("expected at least one save tick, got %d", fs.saveCnt)
	}
}
