package handlers

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_WithHashing(t *testing.T) {
	logger := zerolog.Nop()
	hashKey := "test-secret-key"

	// Initialize hasher pool
	utils.InitHasherPool(hashKey)

	tests := []struct {
		name                 string
		hashKey              string
		requestBody          []byte
		requestHash          string
		calculateRequestHash bool
		expectedStatus       int
		expectedResponseBody string
		checkResponseHash    bool
		expectedResponseHash string
	}{
		{
			name:                 "valid hash - request and response",
			hashKey:              hashKey,
			requestBody:          []byte(`{"type":"counter","id":"test","delta":10}`),
			calculateRequestHash: true,
			expectedStatus:       http.StatusOK,
			expectedResponseBody: "processed",
			checkResponseHash:    true,
		},
		{
			name:                 "no hash key configured - skip hashing",
			hashKey:              "",
			requestBody:          []byte(`{"type":"counter"}`),
			requestHash:          "invalid-hash",
			expectedStatus:       http.StatusOK,
			expectedResponseBody: "processed",
			checkResponseHash:    false,
		},
		{
			name:                 "invalid hash - mismatch",
			hashKey:              hashKey,
			requestBody:          []byte(`{"type":"counter"}`),
			requestHash:          "0000000000000000000000000000000000000000000000000000000000000000",
			calculateRequestHash: false,
			expectedStatus:       http.StatusBadRequest,
		},
		{
			name:                 "empty request body with valid hash",
			hashKey:              hashKey,
			requestBody:          []byte{},
			calculateRequestHash: true,
			expectedStatus:       http.StatusOK,
			expectedResponseBody: "processed",
			checkResponseHash:    true,
		},
		{
			name:                 "large request body with valid hash",
			hashKey:              hashKey,
			requestBody:          bytes.Repeat([]byte("test data "), 1000),
			calculateRequestHash: true,
			expectedStatus:       http.StatusOK,
			expectedResponseBody: "processed",
			checkResponseHash:    true,
		},
		{
			name:                 "JSON request with valid hash",
			hashKey:              hashKey,
			requestBody:          []byte(`{"name":"test","values":[1,2,3,4,5]}`),
			calculateRequestHash: true,
			expectedStatus:       http.StatusOK,
			expectedResponseBody: `{"result":"success"}`,
			checkResponseHash:    true,
		},
		{
			name:                 "special characters in body with valid hash",
			hashKey:              hashKey,
			requestBody:          []byte(`{"text":"special: !@#$%^&*()_+-=[]{}|;:',.<>?/~"}`),
			calculateRequestHash: true,
			expectedStatus:       http.StatusOK,
			expectedResponseBody: "processed",
			checkResponseHash:    true,
		},
		{
			name:                 "unicode characters with valid hash",
			hashKey:              hashKey,
			requestBody:          []byte(`{"message":"Привет мир 🌍"}`),
			calculateRequestHash: true,
			expectedStatus:       http.StatusOK,
			expectedResponseBody: "processed",
			checkResponseHash:    true,
		},
		{
			name:                 "hash with wrong case - should fail",
			hashKey:              hashKey,
			requestBody:          []byte(`{"type":"counter"}`),
			requestHash:          "ABCDEF1234567890",
			calculateRequestHash: false,
			expectedStatus:       http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler := &Handler{
				logger:  &logger,
				hashKey: tt.hashKey,
			}

			// Create next handler
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify body can be read (should be restored by middleware)
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err, "failed to read body in next handler")
				assert.Equal(t, tt.requestBody, body, "request body should match")

				w.WriteHeader(tt.expectedStatus)
				if tt.expectedResponseBody != "" {
					w.Write([]byte(tt.expectedResponseBody))
				}
			})

			// Create middleware
			middleware := handler.WithHashing(nextHandler)

			// Calculate request hash if needed
			var requestHash string
			if tt.calculateRequestHash {
				hash := utils.Hash(tt.requestBody)
				requestHash = hex.EncodeToString(hash)
			} else {
				requestHash = tt.requestHash
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(tt.requestBody))
			if requestHash != "" {
				req.Header.Set("HashSHA256", requestHash)
			}

			rr := httptest.NewRecorder()

			// Execute middleware
			middleware.ServeHTTP(rr, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, rr.Code, "unexpected status code")

			// Assert response body
			if tt.expectedResponseBody != "" && tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedResponseBody, rr.Body.String(), "unexpected response body")
			}

			// Check response hash
			if tt.checkResponseHash && tt.expectedStatus == http.StatusOK && tt.expectedResponseBody != "" {
				responseHash := rr.Header().Get("HashSHA256")
				assert.NotEmpty(t, responseHash, "response should have HashSHA256 header")

				// Verify response hash is correct
				expectedHash := hex.EncodeToString(utils.Hash([]byte(tt.expectedResponseBody)))
				assert.Equal(t, expectedHash, responseHash, "response hash should match")
			}
		})
	}
}

func TestHandler_WithHashing_MultipleRequests(t *testing.T) {
	logger := zerolog.Nop()
	hashKey := "test-secret-key"
	utils.InitHasherPool(hashKey)

	handler := &Handler{
		logger:  &logger,
		hashKey: hashKey,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Response: " + string(body)))
	})

	middleware := handler.WithHashing(nextHandler)

	// Test multiple sequential requests
	for i := 0; i < 5; i++ {
		requestBody := []byte(`{"request":` + string(rune('0'+i)) + `}`)
		requestHash := hex.EncodeToString(utils.Hash(requestBody))

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(requestBody))
		req.Header.Set("HashSHA256", requestHash)

		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "request %d failed", i)

		responseHash := rr.Header().Get("HashSHA256")
		assert.NotEmpty(t, responseHash, "request %d: response should have hash", i)

		expectedResponseBody := "Response: " + string(requestBody)
		expectedHash := hex.EncodeToString(utils.Hash([]byte(expectedResponseBody)))
		assert.Equal(t, expectedHash, responseHash, "request %d: hash mismatch", i)
	}
}

func TestHandler_WithHashing_ConcurrentRequests(t *testing.T) {
	logger := zerolog.Nop()
	hashKey := "test-secret-key"
	utils.InitHasherPool(hashKey)

	handler := &Handler{
		logger:  &logger,
		hashKey: hashKey,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	middleware := handler.WithHashing(nextHandler)

	const numGoroutines = 50
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			requestBody := []byte(`{"id":` + string(rune('0'+id%10)) + `}`)
			requestHash := hex.EncodeToString(utils.Hash(requestBody))

			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(requestBody))
			req.Header.Set("HashSHA256", requestHash)

			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)

			responseHash := rr.Header().Get("HashSHA256")
			expectedHash := hex.EncodeToString(utils.Hash(requestBody))
			assert.Equal(t, expectedHash, responseHash)

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestHandler_WithHashing_BodyReadTwice(t *testing.T) {
	logger := zerolog.Nop()
	hashKey := "test-secret-key"
	utils.InitHasherPool(hashKey)

	handler := &Handler{
		logger:  &logger,
		hashKey: hashKey,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read body first time
		body1, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		// Try to read body second time (should be empty)
		body2, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("first: " + string(body1) + ", second: " + string(body2)))
	})

	middleware := handler.WithHashing(nextHandler)

	requestBody := []byte("test data")
	requestHash := hex.EncodeToString(utils.Hash(requestBody))

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(requestBody))
	req.Header.Set("HashSHA256", requestHash)

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "first: test data, second: ", rr.Body.String())
}

func TestHashingResponseWriter_WriteHeader(t *testing.T) {
	logger := zerolog.Nop()
	hashKey := "test-secret-key"
	utils.InitHasherPool(hashKey)

	handler := &Handler{
		logger:  &logger,
		hashKey: hashKey,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	})

	middleware := handler.WithHashing(nextHandler)

	requestBody := []byte("test")
	requestHash := hex.EncodeToString(utils.Hash(requestBody))

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(requestBody))
	req.Header.Set("HashSHA256", requestHash)

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.NotEmpty(t, rr.Header().Get("HashSHA256"))
}

func TestHandler_WithHashing_EmptyHashKey(t *testing.T) {
	logger := zerolog.Nop()

	handler := &Handler{
		logger:  &logger,
		hashKey: "",
	}

	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := handler.WithHashing(nextHandler)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("test")))
	req.Header.Set("HashSHA256", "some-hash")

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	assert.True(t, nextCalled, "next handler should be called")
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Header().Get("HashSHA256"), "no hash should be set when hashKey is empty")
}
