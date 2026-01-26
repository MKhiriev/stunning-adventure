package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPingService is a mock implementation of PingService
type MockPingService struct {
	mock.Mock
}

func (m *MockPingService) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestHandler_DatabaseConnectionCheck(t *testing.T) {
	tests := []struct {
		name              string
		pingError         error
		expectedStatus    int
		expectedBody      string
		nextHandlerCalled bool
	}{
		{
			name:              "database connected - success",
			pingError:         nil,
			expectedStatus:    http.StatusOK,
			expectedBody:      "pong",
			nextHandlerCalled: true,
		},
		{
			name:              "database not connected - connection error",
			pingError:         errors.New("connection refused"),
			expectedStatus:    http.StatusInternalServerError,
			expectedBody:      "Internal Server Error\n",
			nextHandlerCalled: false,
		},
		{
			name:              "database not connected - timeout error",
			pingError:         errors.New("context deadline exceeded"),
			expectedStatus:    http.StatusInternalServerError,
			expectedBody:      "Internal Server Error\n",
			nextHandlerCalled: false,
		},
		{
			name:              "database not connected - nil pointer",
			pingError:         errors.New("DB connection is nil"),
			expectedStatus:    http.StatusInternalServerError,
			expectedBody:      "Internal Server Error\n",
			nextHandlerCalled: false,
		},
		{
			name:              "database not connected - network error",
			pingError:         errors.New("network unreachable"),
			expectedStatus:    http.StatusInternalServerError,
			expectedBody:      "Internal Server Error\n",
			nextHandlerCalled: false,
		},
		{
			name:              "database not connected - authentication error",
			pingError:         errors.New("authentication failed"),
			expectedStatus:    http.StatusInternalServerError,
			expectedBody:      "Internal Server Error\n",
			nextHandlerCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zerolog.Nop()

			// Create mock ping service
			mockPingService := new(MockPingService)
			mockPingService.On("Ping", mock.Anything).Return(tt.pingError)

			handler := &Handler{
				logger:        &logger,
				dbPingService: mockPingService,
			}

			// Track if next handler was called
			nextHandlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextHandlerCalled = true
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("pong"))
			})

			// Create middleware
			middleware := handler.DatabaseConnectionCheck(nextHandler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			rr := httptest.NewRecorder()

			// Execute middleware
			middleware.ServeHTTP(rr, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, rr.Code, "unexpected status code")

			// Assert response body
			assert.Equal(t, tt.expectedBody, rr.Body.String(), "unexpected response body")

			// Assert next handler was called or not
			assert.Equal(t, tt.nextHandlerCalled, nextHandlerCalled, "next handler call mismatch")

			// Verify mock was called
			mockPingService.AssertExpectations(t)
		})
	}
}

func TestHandler_DatabaseConnectionCheck_ContextPropagation(t *testing.T) {
	logger := zerolog.Nop()

	mockPingService := new(MockPingService)
	mockPingService.On("Ping", mock.Anything).Return(nil)

	handler := &Handler{
		logger:        &logger,
		dbPingService: mockPingService,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context is still available
		assert.NotNil(t, r.Context(), "context should not be nil")
		w.WriteHeader(http.StatusOK)
	})

	middleware := handler.DatabaseConnectionCheck(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockPingService.AssertExpectations(t)
}

func TestHandler_DatabaseConnectionCheck_MultipleRequests(t *testing.T) {
	logger := zerolog.Nop()

	mockPingService := new(MockPingService)
	// Setup mock to return success for all calls
	mockPingService.On("Ping", mock.Anything).Return(nil)

	handler := &Handler{
		logger:        &logger,
		dbPingService: mockPingService,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	middleware := handler.DatabaseConnectionCheck(nextHandler)

	// Execute multiple requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "request %d failed", i)
		assert.Equal(t, "pong", rr.Body.String(), "request %d body mismatch", i)
	}

	// Verify Ping was called 5 times
	mockPingService.AssertNumberOfCalls(t, "Ping", 5)
}

func TestHandler_DatabaseConnectionCheck_ConcurrentRequests(t *testing.T) {
	logger := zerolog.Nop()

	mockPingService := new(MockPingService)
	mockPingService.On("Ping", mock.Anything).Return(nil)

	handler := &Handler{
		logger:        &logger,
		dbPingService: mockPingService,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	middleware := handler.DatabaseConnectionCheck(nextHandler)

	const numGoroutines = 20
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			rr := httptest.NewRecorder()

			middleware.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "request %d failed", id)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify Ping was called correct number of times
	mockPingService.AssertNumberOfCalls(t, "Ping", numGoroutines)
}

func TestHandler_DatabaseConnectionCheck_AlternatingSuccess(t *testing.T) {
	logger := zerolog.Nop()

	mockPingService := new(MockPingService)

	handler := &Handler{
		logger:        &logger,
		dbPingService: mockPingService,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	middleware := handler.DatabaseConnectionCheck(nextHandler)

	testCases := []struct {
		pingError      error
		expectedStatus int
	}{
		{nil, http.StatusOK},
		{errors.New("connection error"), http.StatusInternalServerError},
		{nil, http.StatusOK},
		{errors.New("timeout"), http.StatusInternalServerError},
		{nil, http.StatusOK},
	}

	for i, tc := range testCases {
		// Setup mock for this specific call
		mockPingService.On("Ping", mock.Anything).Return(tc.pingError).Once()

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		assert.Equal(t, tc.expectedStatus, rr.Code, "request %d status mismatch", i)
	}

	mockPingService.AssertExpectations(t)
}

func TestHandler_DatabaseConnectionCheck_DifferentHTTPMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			logger := zerolog.Nop()

			mockPingService := new(MockPingService)
			mockPingService.On("Ping", mock.Anything).Return(nil)

			handler := &Handler{
				logger:        &logger,
				dbPingService: mockPingService,
			}

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("pong"))
			})

			middleware := handler.DatabaseConnectionCheck(nextHandler)

			req := httptest.NewRequest(method, "/ping", nil)
			rr := httptest.NewRecorder()

			middleware.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "method %s failed", method)
			mockPingService.AssertExpectations(t)
		})
	}
}

func TestHandler_DatabaseConnectionCheck_WithHeaders(t *testing.T) {
	logger := zerolog.Nop()

	mockPingService := new(MockPingService)
	mockPingService.On("Ping", mock.Anything).Return(nil)

	handler := &Handler{
		logger:        &logger,
		dbPingService: mockPingService,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers are accessible
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-agent", r.Header.Get("User-Agent"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	middleware := handler.DatabaseConnectionCheck(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockPingService.AssertExpectations(t)
}

func TestHandler_DatabaseConnectionCheck_ErrorDoesNotAffectResponse(t *testing.T) {
	logger := zerolog.Nop()

	mockPingService := new(MockPingService)
	mockPingService.On("Ping", mock.Anything).Return(errors.New("database error"))

	handler := &Handler{
		logger:        &logger,
		dbPingService: mockPingService,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should NOT be called
		t.Error("next handler should not be called when ping fails")
		w.WriteHeader(http.StatusOK)
	})

	middleware := handler.DatabaseConnectionCheck(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "Internal Server Error\n", rr.Body.String())
	mockPingService.AssertExpectations(t)
}

func TestHandler_DatabaseConnectionCheck_NilPingService(t *testing.T) {
	logger := zerolog.Nop()

	handler := &Handler{
		logger:        &logger,
		dbPingService: nil, // nil service
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := handler.DatabaseConnectionCheck(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	// Should panic or handle nil gracefully
	assert.Panics(t, func() {
		middleware.ServeHTTP(rr, req)
	}, "should panic when dbPingService is nil")
}

func TestHandler_DatabaseConnectionCheck_ContextCancellation(t *testing.T) {
	logger := zerolog.Nop()

	mockPingService := new(MockPingService)
	mockPingService.On("Ping", mock.Anything).Return(context.Canceled)

	handler := &Handler{
		logger:        &logger,
		dbPingService: mockPingService,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called when context is cancelled")
		w.WriteHeader(http.StatusOK)
	})

	middleware := handler.DatabaseConnectionCheck(nextHandler)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := httptest.NewRequest(http.MethodGet, "/ping", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockPingService.AssertExpectations(t)
}
