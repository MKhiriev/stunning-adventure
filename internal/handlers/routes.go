package handlers

import (
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/service"
	"github.com/MKhiriev/stunning-adventure/internal/validators"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type Handler struct {
	logger          *zerolog.Logger
	metricsService  service.MetricsService
	dbPingService   service.PingService
	metricValidator validators.Validator
	hashKey         string
}

func NewHandler(metricsService service.MetricsService, dbPingService service.PingService, cfg *config.ServerConfig, logger *zerolog.Logger) *Handler {
	return &Handler{
		logger:          logger,
		metricsService:  metricsService,
		dbPingService:   dbPingService,
		metricValidator: validators.NewMetricsValidator(),
		hashKey:         cfg.HashKey,
	}
}

func (h *Handler) Init() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer, h.WithLogging, WithContext)
	router.Group(func(r chi.Router) {
		r.Use(GZip, h.WithHashing)
		r.Post("/updates/", h.BatchUpdateMetricJSON)
		r.Post("/update/", h.UpdateMetricJSON)
		r.Post("/value/", h.GetMetricJSON)
		r.Get("/", h.GetAllMetrics)
	})

	router.Group(func(r chi.Router) {
		r.Post("/update/{metricType}/{metricName}/{metricValue}", h.MetricHandler)
		r.Get("/value/{metricType}/{metricName}", h.GetMetricValue)
	})

	router.Group(func(r chi.Router) {
		r.Use(h.DatabaseConnectionCheck)
		r.Get("/ping", h.Ping)
	})

	router.MethodNotAllowed(CheckHTTPMethod(router))

	return router
}
