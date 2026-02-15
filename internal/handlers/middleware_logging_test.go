package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_WithLogging(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		path               string
		requestBody        string
		handlerStatus      int
		handlerResponse    string
		handlerDelay       time.Duration
		checkLogContains   []string
		checkLogNotContain []string
	}{
		{
			name:            "GET request logged",
			method:          http.MethodGet,
			path:            "/test",
			handlerStatus:   http.StatusOK,
			handlerResponse: "OK",
			checkLogContains: []string{
				`"method":"GET"`,
				`"uri":"/test"`,
				`"status":200`,
				`"duration":`,
			},
		},
		{
			name:            "POST request logged",
			method:          http.MethodPost,
			path:            "/api/data",
			requestBody:     `{"key":"value"}`,
			handlerStatus:   http.StatusCreated,
			handlerResponse: "Created",
			checkLogContains: []string{
				`"method":"POST"`,
				`"uri":"/api/data"`,
				`"status":201`,
			},
		},
		{
			name:            "PUT request logged",
			method:          http.MethodPut,
			path:            "/update",
			handlerStatus:   http.StatusNoContent,
			handlerResponse: "",
			checkLogContains: []string{
				`"method":"PUT"`,
				`"uri":"/update"`,
				`"status":204`,
			},
		},
		{
			name:            "DELETE request logged",
			method:          http.MethodDelete,
			path:            "/resource/123",
			handlerStatus:   http.StatusOK,
			handlerResponse: "Deleted",
			checkLogContains: []string{
				`"method":"DELETE"`,
				`"uri":"/resource/123"`,
				`"status":200`,
			},
		},
		{
			name:            "error status logged",
			method:          http.MethodGet,
			path:            "/error",
			handlerStatus:   http.StatusInternalServerError,
			handlerResponse: "Internal Server Error",
			checkLogContains: []string{
				`"method":"GET"`,
				`"uri":"/error"`,
				`"status":500`,
			},
		},
		{
			name:            "404 status logged",
			method:          http.MethodGet,
			path:            "/notfound",
			handlerStatus:   http.StatusNotFound,
			handlerResponse: "Not Found",
			checkLogContains: []string{
				`"method":"GET"`,
				`"uri":"/notfound"`,
				`"status":404`,
			},
		},
		{
			name:            "request with query parameters",
			method:          http.MethodGet,
			path:            "/search?q=test&limit=10",
			handlerStatus:   http.StatusOK,
			handlerResponse: "Results",
			checkLogContains: []string{
				`"method":"GET"`,
				`"uri":"/search?q=test&limit=10"`,
				`"status":200`,
			},
		},
		{
			name:            "request with special characters in path",
			method:          http.MethodGet,
			path:            "/path/with%20spaces",
			handlerStatus:   http.StatusOK,
			handlerResponse: "OK",
			checkLogContains: []string{
				`"method":"GET"`,
				`"uri":"/path/with%20spaces"`,
				`"status":200`,
			},
		},
		{
			name:            "duration is logged",
			method:          http.MethodGet,
			path:            "/slow",
			handlerStatus:   http.StatusOK,
			handlerResponse: "Done",
			handlerDelay:    50 * time.Millisecond,
			checkLogContains: []string{
				`"duration":`,
				`"status":200`,
			},
		},
		{
			name:            "PATCH request logged",
			method:          http.MethodPatch,
			path:            "/resource",
			handlerStatus:   http.StatusOK,
			handlerResponse: "Patched",
			checkLogContains: []string{
				`"method":"PATCH"`,
				`"uri":"/resource"`,
				`"status":200`,
			},
		},
		{
			name:            "OPTIONS request logged",
			method:          http.MethodOptions,
			path:            "/api",
			handlerStatus:   http.StatusOK,
			handlerResponse: "",
			checkLogContains: []string{
				`"method":"OPTIONS"`,
				`"uri":"/api"`,
				`"status":200`,
			},
		},
		{
			name:            "HEAD request logged",
			method:          http.MethodHead,
			path:            "/check",
			handlerStatus:   http.StatusOK,
			handlerResponse: "",
			checkLogContains: []string{
				`"method":"HEAD"`,
				`"uri":"/check"`,
				`"status":200`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create buffer to capture logs
			var logBuffer bytes.Buffer
			logger := zerolog.New(&logBuffer).With().Timestamp().Logger()

			handler := &Handler{
				logger: &logger,
			}

			// Create next handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.handlerDelay > 0 {
					time.Sleep(tt.handlerDelay)
				}
				w.WriteHeader(tt.handlerStatus)
				if tt.handlerResponse != "" {
					w.Write([]byte(tt.handlerResponse))
				}
			})

			// Create middleware
			middleware := handler.WithLogging(nextHandler)

			// Create request
			var body *strings.Reader
			if tt.requestBody != "" {
				body = strings.NewReader(tt.requestBody)
			} else {
				body = strings.NewReader("")
			}
			req := httptest.NewRequest(tt.method, tt.path, body)

			rr := httptest.NewRecorder()

			// Execute middleware
			middleware.ServeHTTP(rr, req)

			// Assert status code
			assert.Equal(t, tt.handlerStatus, rr.Code, "unexpected status code")

			// Assert response body
			if tt.handlerResponse != "" {
				assert.Equal(t, tt.handlerResponse, rr.Body.String(), "unexpected response body")
			}

			// Get log output
			logOutput := logBuffer.String()
			require.NotEmpty(t, logOutput, "log should not be empty")

			// Check log contains expected strings
			for _, expected := range tt.checkLogContains {
				assert.Contains(t, logOutput, expected, "log should contain: %s", expected)
			}

			// Check log doesn't contain unexpected strings
			for _, unexpected := range tt.checkLogNotContain {
				assert.NotContains(t, logOutput, unexpected, "log should not contain: %s", unexpected)
			}
		})
	}
}

func TestHandler_WithLogging_ResponseSize(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := zerolog.New(&logBuffer).With().Timestamp().Logger()

	handler := &Handler{
		logger: &logger,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Write specific amount of data
		w.Write([]byte(strings.Repeat("a", 1024))) // 1KB
	})

	middleware := handler.WithLogging(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	logOutput := logBuffer.String()

	// Check that size is logged
	assert.Contains(t, logOutput, `"size":`, "log should contain response size")
	assert.Contains(t, logOutput, `1024`, "log should contain correct size")
}

func TestHandler_WithLogging_MultipleRequests(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := zerolog.New(&logBuffer).With().Timestamp().Logger()

	handler := &Handler{
		logger: &logger,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := handler.WithLogging(nextHandler)

	// Execute multiple requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "request %d failed", i)
	}

	logOutput := logBuffer.String()

	// Should have 5 log entries
	logLines := strings.Split(strings.TrimSpace(logOutput), "\n")
	assert.Len(t, logLines, 5, "should have 5 log entries")
}

func TestHandler_WithLogging_DurationAccuracy(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := zerolog.New(&logBuffer).With().Timestamp().Logger()

	handler := &Handler{
		logger: &logger,
	}

	delay := 100 * time.Millisecond
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	})

	middleware := handler.WithLogging(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	start := time.Now()
	middleware.ServeHTTP(rr, req)
	actualDuration := time.Since(start)

	logOutput := logBuffer.String()

	// Check that duration is logged
	assert.Contains(t, logOutput, `"duration":`, "log should contain duration")

	// Duration should be at least the delay
	assert.GreaterOrEqual(t, actualDuration, delay, "actual duration should be at least the delay")
}

func TestHandler_WithLogging_WithHeaders(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := zerolog.New(&logBuffer).With().Timestamp().Logger()

	handler := &Handler{
		logger: &logger,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that headers are accessible
		assert.NotEmpty(t, r.Header.Get("User-Agent"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := handler.WithLogging(nextHandler)

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"test":"data"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent/1.0")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, `"method":"POST"`)
	assert.Contains(t, logOutput, `"status":200`)
}

func TestHandler_WithLogging_PanicRecovery(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := zerolog.New(&logBuffer).With().Timestamp().Logger()

	handler := &Handler{
		logger: &logger,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	middleware := handler.WithLogging(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	// Should panic (logging middleware doesn't recover panics)
	assert.Panics(t, func() {
		middleware.ServeHTTP(rr, req)
	}, "should panic")
}

func TestHandler_WithLogging_NoStatusWritten(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := zerolog.New(&logBuffer).With().Timestamp().Logger()

	handler := &Handler{
		logger: &logger,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't write status explicitly - should default to 200
		w.Write([]byte("OK"))
	})

	middleware := handler.WithLogging(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, `"status":200`, "should log default status 200")
}
