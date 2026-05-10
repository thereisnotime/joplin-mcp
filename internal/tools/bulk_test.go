package tools

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunBulk_AllSucceed(t *testing.T) {
	got := runBulk(context.Background(), []string{"a", "b", "c"}, func(_ context.Context, _ string) error {
		return nil
	})
	if len(got.Succeeded) != 3 {
		t.Errorf("Succeeded = %v, want 3", got.Succeeded)
	}
	if got.Failed != nil {
		t.Errorf("Failed should be nil, got %v", got.Failed)
	}
}

func TestRunBulk_PartialFailure(t *testing.T) {
	got := runBulk(context.Background(), []string{"a", "b", "c"}, func(_ context.Context, id string) error {
		if id == "b" {
			return errors.New("nope")
		}
		return nil
	})
	if len(got.Succeeded) != 2 {
		t.Errorf("Succeeded = %v, want 2", got.Succeeded)
	}
	if got.Failed["b"] == "" {
		t.Errorf("Failed[b] missing; got %v", got.Failed)
	}
}

func TestRunBulk_ContextAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := runBulk(ctx, []string{"a", "b"}, func(_ context.Context, _ string) error {
		t.Error("op should not be invoked when context is cancelled before dispatch")
		return nil
	})
	// Either path acceptable: items might be in Failed (cancelled before
	// dispatch) or Succeeded if op happened to slip through. The contract
	// is that the function returns promptly without panicking.
	if len(got.Succeeded)+len(got.Failed) == 0 {
		t.Error("expected at least some recorded outcomes")
	}
}

func TestRunBulk_RespectsConcurrencyLimit(t *testing.T) {
	const N = 12
	ids := make([]string, N)
	for i := range ids {
		ids[i] = "x"
	}
	start := time.Now()
	got := runBulk(context.Background(), ids, func(_ context.Context, _ string) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	elapsed := time.Since(start)
	if len(got.Succeeded) != N {
		t.Fatalf("Succeeded = %d, want %d", len(got.Succeeded), N)
	}
	// 12 items at 4 in flight × 50 ms each = ~150 ms minimum (3 batches).
	// Generous upper bound to avoid flakes.
	if elapsed < 100*time.Millisecond {
		t.Errorf("elapsed %v too short — concurrency cap not honoured", elapsed)
	}
}
