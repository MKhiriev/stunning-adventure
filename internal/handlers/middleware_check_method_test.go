package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestCheckHTTPMethod(t *testing.T) {
	tests := []struct {
		name              string
		setupRoutes       func(r *chi.Mux)
		requestMethod     string
		requestPath       string
		expectedStatus    int
		expectedBody      string
		shouldCallHandler bool
	}{
		{
			name: "valid GET request to registered route",
			setupRoutes: func(r *chi.Mux) {
				r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("GET OK"))
				})
			},
			requestMethod:     http.MethodGet,
			requestPath:       "/test",
			expectedStatus:    http.StatusOK,
			expectedBody:      "GET OK",
			shouldCallHandler: true,
		},
		{
			name: "valid POST request to registered route",
			setupRoutes: func(r *chi.Mux) {
				r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("POST OK"))
				})
			},
			requestMethod:     http.MethodPost,
			requestPath:       "/test",
			expectedStatus:    http.StatusCreated,
			expectedBody:      "POST OK",
			shouldCallHandler: true,
		},
		{
			name: "wrong method for registered route - POST instead of GET",
			setupRoutes: func(r *chi.Mux) {
				r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("GET OK"))
				})
			},
			requestMethod:     http.MethodPost,
			requestPath:       "/test",
			expectedStatus:    http.StatusNotFound,
			expectedBody:      "",
			shouldCallHandler: false,
		},
		{
			name: "wrong method for registered route - GET instead of POST",
			setupRoutes: func(r *chi.Mux) {
				r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("POST OK"))
				})
			},
			requestMethod:     http.MethodGet,
			requestPath:       "/test",
			expectedStatus:    http.StatusNotFound,
			expectedBody:      "",
			shouldCallHandler: false,
		},
		{
			name: "PUT method on GET-only route",
			setupRoutes: func(r *chi.Mux) {
				r.Get("/api/resource", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("GET OK"))
				})
			},
			requestMethod:     http.MethodPut,
			requestPath:       "/api/resource",
			expectedStatus:    http.StatusNotFound,
			expectedBody:      "",
			shouldCallHandler: false,
		},
		{
			name: "DELETE method on POST-only route",
			setupRoutes: func(r *chi.Mux) {
				r.Post("/api/resource", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("POST OK"))
				})
			},
			requestMethod:     http.MethodDelete,
			requestPath:       "/api/resource",
			expectedStatus:    http.StatusNotFound,
			expectedBody:      "",
			shouldCallHandler: false,
		},
		{
			name: "route with multiple methods - valid GET",
			setupRoutes: func(r *chi.Mux) {
				r.Get("/multi", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("GET OK"))
				})
				r.Post("/multi", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("POST OK"))
				})
			},
			requestMethod:     http.MethodGet,
			requestPath:       "/multi",
			expectedStatus:    http.StatusOK,
			expectedBody:      "GET OK",
			shouldCallHandler: true,
		},
		{
			name: "route with multiple methods - valid POST",
			setupRoutes: func(r *chi.Mux) {
				r.Get("/multi", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("GET OK"))
				})
				r.Post("/multi", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("POST OK"))
				})
			},
			requestMethod:     http.MethodPost,
			requestPath:       "/multi",
			expectedStatus:    http.StatusCreated,
			expectedBody:      "POST OK",
			shouldCallHandler: true,
		},
		{
			name: "route with multiple methods - invalid PUT",
			setupRoutes: func(r *chi.Mux) {
				r.Get("/multi", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("GET OK"))
				})
				r.Post("/multi", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("POST OK"))
				})
			},
			requestMethod:     http.MethodPut,
			requestPath:       "/multi",
			expectedStatus:    http.StatusNotFound,
			expectedBody:      "",
			shouldCallHandler: false,
		},
		{
			name: "unregistered route",
			setupRoutes: func(r *chi.Mux) {
				r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("GET OK"))
				})
			},
			requestMethod:     http.MethodGet,
			requestPath:       "/nonexistent",
			expectedStatus:    http.StatusNotFound,
			expectedBody:      "404 page not found\n",
			shouldCallHandler: false,
		},
		{
			name: "route with path parameters - valid method",
			setupRoutes: func(r *chi.Mux) {
				r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("User ID: " + chi.URLParam(r, "id")))
				})
			},
			requestMethod:     http.MethodGet,
			requestPath:       "/users/123",
			expectedStatus:    http.StatusOK,
			expectedBody:      "User ID: 123",
			shouldCallHandler: true,
		},
		{
			name: "PATCH method on route",
			setupRoutes: func(r *chi.Mux) {
				r.Patch("/resource", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("PATCH OK"))
				})
			},
			requestMethod:     http.MethodPatch,
			requestPath:       "/resource",
			expectedStatus:    http.StatusOK,
			expectedBody:      "PATCH OK",
			shouldCallHandler: true,
		},
		{
			name: "OPTIONS method",
			setupRoutes: func(r *chi.Mux) {
				r.Options("/resource", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("OPTIONS OK"))
				})
			},
			requestMethod:     http.MethodOptions,
			requestPath:       "/resource",
			expectedStatus:    http.StatusOK,
			expectedBody:      "OPTIONS OK",
			shouldCallHandler: true,
		},
		{
			name: "HEAD method",
			setupRoutes: func(r *chi.Mux) {
				r.Head("/resource", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				})
			},
			requestMethod:     http.MethodHead,
			requestPath:       "/resource",
			expectedStatus:    http.StatusOK,
			expectedBody:      "",
			shouldCallHandler: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router and setup routes
			router := chi.NewRouter()
			tt.setupRoutes(router)

			// Set MethodNotAllowed handler
			router.MethodNotAllowed(CheckHTTPMethod(router))

			// Create test request
			req := httptest.NewRequest(tt.requestMethod, tt.requestPath, nil)
			rr := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(rr, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, rr.Code, "unexpected status code")

			// Assert response body
			assert.Equal(t, tt.expectedBody, rr.Body.String(), "unexpected response body")
		})
	}
}

func TestCheckHTTPMethod_EmptyRouter(t *testing.T) {
	router := chi.NewRouter()
	router.MethodNotAllowed(CheckHTTPMethod(router))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "expected 404 for empty router")
}

func TestCheckHTTPMethod_ComplexRouting(t *testing.T) {
	router := chi.NewRouter()

	// Setup complex routing structure
	router.Route("/api", func(r chi.Router) {
		r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("GET users"))
		})
		r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("POST users"))
		})
	})

	router.MethodNotAllowed(CheckHTTPMethod(router))

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid GET to /api/users",
			method:         http.MethodGet,
			path:           "/api/users",
			expectedStatus: http.StatusOK,
			expectedBody:   "GET users",
		},
		{
			name:           "valid POST to /api/users",
			method:         http.MethodPost,
			path:           "/api/users",
			expectedStatus: http.StatusCreated,
			expectedBody:   "POST users",
		},
		{
			name:           "invalid DELETE to /api/users",
			method:         http.MethodDelete,
			path:           "/api/users",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
		})
	}
}
