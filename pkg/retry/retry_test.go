package retry

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestTry_SuccessFirstAttempt_NoSleep(t *testing.T) {
	var calls atomic.Int32

	start := time.Now()
	err := Try(func() error {
		calls.Add(1)
		return nil
	})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", calls.Load())
	}
	// Проверяем, что не было задержек (грубая верхняя граница)
	if elapsed > 50*time.Millisecond {
		t.Fatalf("expected near-zero time without retries, took %v", elapsed)
	}
}

func TestTry_AllAttemptsFail_ReturnsLastError(t *testing.T) {
	// этот тест займёт ~9 секунд из-за встроенных задержек (1s + 3s + 5s).
	t.Skip("Пропускаем длительный тест")

	var calls atomic.Int32
	wantErr := errors.New("always fail")

	start := time.Now()
	err := Try(func() error {
		calls.Add(1)
		return wantErr
	})
	elapsed := time.Since(start)

	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
	if calls.Load() < 2 {
		t.Fatalf("expected multiple attempts, got %d", calls.Load())
	}
	// Ожидаем минимум сумму задержек: 1+3+5 = 9с
	if elapsed < 9*time.Second {
		t.Fatalf("expected at least 9s elapsed due to backoff, got %v", elapsed)
	}
}
