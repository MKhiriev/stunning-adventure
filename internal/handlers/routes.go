package handlers

import (
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type Handler struct {
	MemStorage *store.MemStorage
	Logger     *zerolog.Logger
}

func NewHandler(logger *zerolog.Logger) *Handler {
	return &Handler{
		MemStorage: store.NewMemStorage(),
		Logger:     logger,
	}
}

func (h *Handler) Init() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer, h.WithLogging)

	router.Route("/update", func(r chi.Router) {
		r.Post("/", h.JSONMetricHandler)
		r.Post("/{metricType}/{metricName}/{metricValue}", h.MetricHandler)
	})
	router.Route("/value", func(r chi.Router) {
		r.Post("/", h.JSONGetMetricValue)
		r.Get("/{metricType}/{metricName}", h.GetMetricValue)
	})
	router.Get("/", h.GetAllMetrics)

	router.MethodNotAllowed(CheckHTTPMethod(router))

	return router
}
