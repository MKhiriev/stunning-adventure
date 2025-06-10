package handlers

import (
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	MemStorage *store.MemStorage
}

func NewHandler() *Handler {
	return &Handler{MemStorage: store.NewMemStorage()}
}

func (h *Handler) Init() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/update/{metricType}/{metricName}/{metricValue}", h.MetricHandler)
	router.Get("/value/{metricType}/{metricName}", h.GetMetricValue)
	router.Get("/", h.GetAllMetrics)

	return router
}
