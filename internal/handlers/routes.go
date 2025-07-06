package handlers

import (
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type Handler struct {
	MemStorage  *store.MemStorage
	FileStorage *store.FileStorage
	Logger      *zerolog.Logger
}

func NewHandler(memStorage *store.MemStorage, fileStorage *store.FileStorage, logger *zerolog.Logger) *Handler {
	return &Handler{
		MemStorage:  memStorage,
		FileStorage: fileStorage,
		Logger:      logger,
	}
}

func (h *Handler) Init() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer, h.WithLogging, WithGzipCompression)

	router.Post("/update/", h.UpdateMetricJSON)
	router.Post("/value/", h.GetMetricJSON)
	router.Post("/update/{metricType}/{metricName}/{metricValue}", h.MetricHandler)
	router.Get("/value/{metricType}/{metricName}", h.GetMetricValue)
	router.Get("/", h.GetAllMetrics)

	router.MethodNotAllowed(CheckHTTPMethod(router))

	return router
}
