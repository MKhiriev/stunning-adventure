package handlers

import (
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type Handler struct {
	memStorage  *store.MemStorage
	fileStorage *store.FileStorage
	logger      *zerolog.Logger
	db          *store.DB
}

func NewHandler(memStorage *store.MemStorage, fileStorage *store.FileStorage, db *store.DB, logger *zerolog.Logger) *Handler {
	return &Handler{
		memStorage:  memStorage,
		fileStorage: fileStorage,
		logger:      logger,
		db:          db,
	}
}

func (h *Handler) Init() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer, h.WithLogging, GZip)

	router.Post("/update/", h.UpdateMetricJSON)
	router.Post("/value/", h.GetMetricJSON)
	router.Post("/update/{metricType}/{metricName}/{metricValue}", h.MetricHandler)
	router.Get("/value/{metricType}/{metricName}", h.GetMetricValue)
	router.Get("/", h.GetAllMetrics)

	router.Get("/ping", h.Ping)

	router.MethodNotAllowed(CheckHTTPMethod(router))

	return router
}
