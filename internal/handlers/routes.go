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
	router.Group(func(r chi.Router) {
		r.Post("/update/", h.UpdateMetricJSON)
		r.Post("/value/", h.GetMetricJSON)
		r.Post("/update/{metricType}/{metricName}/{metricValue}", h.MetricHandler)
		r.Get("/value/{metricType}/{metricName}", h.GetMetricValue)
		r.Get("/", h.GetAllMetrics)
	})

	router.Group(func(r chi.Router) {
		r.Use(h.DatabaseConnectionCheck)
		r.Get("/ping", h.Ping)
	})

	router.MethodNotAllowed(CheckHTTPMethod(router))

	return router
}
