package handlers

import (
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

type Handler struct {
	MemStorage *store.MemStorage
	Routes     []chi.Route
}

func NewHandler() *Handler {
	return &Handler{MemStorage: store.NewMemStorage()}
}

func (h *Handler) Init() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)

	router.Post("/update/{metricType}/{metricName}/{metricValue}", h.MetricHandler)
	router.Get("/value/{metricType}/{metricName}", h.GetMetricValue)
	router.Get("/", h.GetAllMetrics)

	h.Routes = router.Routes()

	router.MethodNotAllowed(CheckHTTPMethod(router))

	return router
}

func CheckHTTPMethod(router *chi.Mux) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestedURL := r.URL.Path
		requestedHTTPMethod := r.Method

		allRoutes := router.Routes()
		var foundRoute chi.Route
		for _, route := range allRoutes {
			if route.Pattern == requestedURL {
				foundRoute = route
				break
			}
		}

		if _, ok := foundRoute.Handlers[requestedHTTPMethod]; !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		router.ServeHTTP(w, r)
	}
}
