package handlers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_WithDecryption(t *testing.T) {
	logger := zerolog.Nop()

	// Generate RSA key pair for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "failed to generate private key")
	publicKey := &privateKey.PublicKey

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedStatus int
		expectedBody   string
		checkBody      func(t *testing.T, body string)
	}{
		{
			name: "successfully decrypt encrypted body",
			setupRequest: func() *http.Request {
				plaintext := []byte(`{"type":"counter","id":"test","delta":10}`)
				encrypted, err := utils.EncryptData(plaintext, publicKey)
				require.NoError(t, err, "failed to encrypt data")

				req := httptest.NewRequest(http.MethodPost, "/test", encrypted)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"type":"counter","id":"test","delta":10}`,
		},
		{
			name: "pass through empty body",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte{}))
				return req
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "empty body received",
		},
		{
			name: "pass through nil body",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				return req
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "empty body received",
		},
		{
			name: "decrypt large encrypted body",
			setupRequest: func() *http.Request {
				// Create a large payload that requires chunking
				plaintext := bytes.Repeat([]byte("test data "), 100)
				encrypted, err := utils.EncryptData(plaintext, publicKey)
				require.NoError(t, err, "failed to encrypt large data")

				req := httptest.NewRequest(http.MethodPost, "/test", encrypted)
				return req
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				expected := string(bytes.Repeat([]byte("test data "), 100))
				assert.Equal(t, expected, body, "decrypted large body should match")
			},
		},
		{
			name: "decrypt JSON with special characters",
			setupRequest: func() *http.Request {
				plaintext := []byte(`{"name":"test","value":"special chars: !@#$%^&*()"}`)
				encrypted, err := utils.EncryptData(plaintext, publicKey)
				require.NoError(t, err, "failed to encrypt data")

				req := httptest.NewRequest(http.MethodPost, "/test", encrypted)
				return req
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"name":"test","value":"special chars: !@#$%^&*()"}`,
		},
		{
			name: "decrypt unicode characters",
			setupRequest: func() *http.Request {
				plaintext := []byte(`{"message":"Привет мир 🌍"}`)
				encrypted, err := utils.EncryptData(plaintext, publicKey)
				require.NoError(t, err, "failed to encrypt data")

				req := httptest.NewRequest(http.MethodPost, "/test", encrypted)
				return req
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Привет мир 🌍"}`,
		},
		{
			name: "fail on corrupted encrypted data",
			setupRequest: func() *http.Request {
				plaintext := []byte(`{"type":"counter"}`)
				encrypted, err := utils.EncryptData(plaintext, publicKey)
				require.NoError(t, err, "failed to encrypt data")

				// Corrupt the encrypted data
				corruptedData := encrypted.Bytes()
				if len(corruptedData) > 10 {
					corruptedData[10] ^= 0xFF // Flip bits
				}

				req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(corruptedData))
				return req
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name: "fail on random data with correct size",
			setupRequest: func() *http.Request {
				// Create random data with correct chunk size (256 bytes for 2048-bit key)
				randomData := make([]byte, 256)
				_, err := rand.Read(randomData)
				require.NoError(t, err)

				req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(randomData))
				return req
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error\n",
		},
		{
			name: "decrypt minimal payload",
			setupRequest: func() *http.Request {
				plaintext := []byte("a")
				encrypted, err := utils.EncryptData(plaintext, publicKey)
				require.NoError(t, err, "failed to encrypt data")

				req := httptest.NewRequest(http.MethodPost, "/test", encrypted)
				return req
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler := &Handler{
				logger:     &logger,
				privateKey: privateKey,
			}

			// Create next handler that reads the decrypted body
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("failed to read body in next handler: %v", err)
				}

				if len(body) == 0 {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("empty body received"))
					return
				}

				w.WriteHeader(http.StatusOK)
				w.Write(body)
			})

			// Create middleware
			middleware := handler.WithDecryption(nextHandler)

			// Setup request
			req := tt.setupRequest()
			rr := httptest.NewRecorder()

			// Execute middleware
			middleware.ServeHTTP(rr, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, rr.Code, "unexpected status code")

			// Assert response body
			if tt.checkBody != nil {
				tt.checkBody(t, rr.Body.String())
			} else {
				assert.Equal(t, tt.expectedBody, rr.Body.String(), "unexpected response body")
			}
		})
	}
}

func TestHandler_WithDecryption_NilPrivateKey(t *testing.T) {
	logger := zerolog.Nop()

	handler := &Handler{
		logger:     &logger,
		privateKey: nil,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := handler.WithDecryption(nextHandler)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("test")))
	rr := httptest.NewRecorder()

	// This should panic or fail because privateKey is nil
	assert.Panics(t, func() {
		middleware.ServeHTTP(rr, req)
	}, "expected panic when privateKey is nil")
}

func TestHandler_WithDecryption_MultipleRequests(t *testing.T) {
	logger := zerolog.Nop()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey

	handler := &Handler{
		logger:     &logger,
		privateKey: privateKey,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	middleware := handler.WithDecryption(nextHandler)

	// Test multiple sequential requests
	for i := 0; i < 5; i++ {
		plaintext := []byte(`{"request":` + string(rune('0'+i)) + `}`)
		encrypted, err := utils.EncryptData(plaintext, publicKey)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/test", encrypted)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "request %d failed", i)
		assert.Equal(t, string(plaintext), rr.Body.String(), "request %d body mismatch", i)
	}
}

func TestHandler_WithDecryption_ChainedMiddleware(t *testing.T) {
	logger := zerolog.Nop()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey

	handler := &Handler{
		logger:     &logger,
		privateKey: privateKey,
	}

	// Create a chain of middleware
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("processed: " + string(body)))
	})

	// Add another middleware before decryption
	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Just pass through
			next.ServeHTTP(w, r)
		})
	}

	chain := loggingMiddleware(handler.WithDecryption(finalHandler))

	plaintext := []byte("test data")
	encrypted, err := utils.EncryptData(plaintext, publicKey)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/test", encrypted)
	rr := httptest.NewRecorder()

	chain.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "processed: test data", rr.Body.String())
}

func TestHandler_WithDecryption_BodyCanBeReadMultipleTimes(t *testing.T) {
	logger := zerolog.Nop()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey

	handler := &Handler{
		logger:     &logger,
		privateKey: privateKey,
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read body first time
		body1, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		// Try to read body second time (should be empty since we consumed it)
		body2, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("first: " + string(body1) + ", second: " + string(body2)))
	})

	middleware := handler.WithDecryption(nextHandler)

	plaintext := []byte("test")
	encrypted, err := utils.EncryptData(plaintext, publicKey)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/test", encrypted)
	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "first: test, second: ", rr.Body.String())
}

func TestHandler_WithDecryption_DifferentKeySizes(t *testing.T) {
	logger := zerolog.Nop()

	keySizes := []int{1024, 2048, 4096}

	for _, keySize := range keySizes {
		t.Run(fmt.Sprintf("key_size_%d", keySize), func(t *testing.T) {
			privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
			require.NoError(t, err)
			publicKey := &privateKey.PublicKey

			handler := &Handler{
				logger:     &logger,
				privateKey: privateKey,
			}

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				w.WriteHeader(http.StatusOK)
				w.Write(body)
			})

			middleware := handler.WithDecryption(nextHandler)

			plaintext := []byte("test data for different key sizes")
			encrypted, err := utils.EncryptData(plaintext, publicKey)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/test", encrypted)
			rr := httptest.NewRecorder()

			middleware.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, string(plaintext), rr.Body.String())
		})
	}
}
