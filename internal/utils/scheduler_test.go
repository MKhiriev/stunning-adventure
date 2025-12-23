package utils

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestRunWithTicker_InvokesFunction(t *testing.T) {
	var called int32

	RunWithTicker(func() {
		atomic.AddInt32(&called, 1)
	}, 10*time.Millisecond)

	// ждём разумное время, чтобы тикер успел сработать
	time.Sleep(30 * time.Millisecond)

	if atomic.LoadInt32(&called) == 0 {
		t.Fatal("expected function to be called at least once")
	}
}

func TestRunWithTicker_CalledMultipleTimes(t *testing.T) {
	var called int32

	RunWithTicker(func() {
		atomic.AddInt32(&called, 1)
	}, 5*time.Millisecond)

	time.Sleep(40 * time.Millisecond)

	count := atomic.LoadInt32(&called)
	if count < 3 {
		t.Fatalf("expected function to be called multiple times, got %d", count)
	}
}

func TestRunWithTicker_AsyncExecution(t *testing.T) {
	start := time.Now()

	RunWithTicker(func() {
		// no-op
	}, 20*time.Millisecond)

	// функция должна вернуться сразу, не блокируясь
	elapsed := time.Since(start)

	if elapsed > 5*time.Millisecond {
		t.Fatalf("RunWithTicker appears to block, elapsed=%v", elapsed)
	}
}
