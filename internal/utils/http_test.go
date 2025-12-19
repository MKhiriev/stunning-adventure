package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	timeout := 5 * time.Second

	client := NewHTTPClient(timeout)
	if client == nil {
		t.Fatal("expected non-nil resty client")
	}

	if client.GetClient().Timeout != timeout {
		t.Fatalf(
			"unexpected timeout: want %v, got %v",
			timeout,
			client.GetClient().Timeout,
		)
	}
}

func TestNewHTTPClient_ZeroTimeout(t *testing.T) {
	client := NewHTTPClient(0)
	if client == nil {
		t.Fatal("expected non-nil resty client")
	}

	if client.GetClient().Timeout != 0 {
		t.Fatalf("expected zero timeout, got %v", client.GetClient().Timeout)
	}
}

func TestNewHTTPClient_HTTPRequestSuccess(t *testing.T) {
	// test server returning immediately
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	client := NewHTTPClient(2 * time.Second)

	resp, err := client.R().
		SetContext(context.Background()).
		Get(srv.URL)

	if err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode())
	}

	if string(resp.Body()) != "ok" {
		t.Fatalf("unexpected response body: %q", resp.Body())
	}
}

func TestNewHTTPClient_TimeoutExceeded(t *testing.T) {
	// server that intentionally delays response
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewHTTPClient(50 * time.Millisecond)

	start := time.Now()
	_, err := client.R().Get(srv.URL)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	// resty wraps timeout errors, so we check timing instead of error text
	if elapsed > 150*time.Millisecond {
		t.Fatalf("request did not timeout fast enough: elapsed=%v", elapsed)
	}
}

func TestNewHTTPClient_ContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewHTTPClient(5 * time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.R().
		SetContext(ctx).
		Get(srv.URL)

	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}
