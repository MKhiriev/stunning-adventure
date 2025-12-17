package handlers

import (
	"crypto/rsa"

	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/internal/service"
	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/MKhiriev/stunning-adventure/internal/validators"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type Handler struct {
	logger         *zerolog.Logger
	metricsService service.MetricsService
	dbPingService  service.PingService
	auditService   service.AuditPublisher

	metricValidator validators.Validator
	hashKey         string

	privateKey *rsa.PrivateKey // key for decrypting messages
}

func NewHandler(metricsService service.MetricsService, dbPingService service.PingService, auditService service.AuditPublisher, privateKey *rsa.PrivateKey, cfg *config.ServerConfig, logger *zerolog.Logger) *Handler {
	handler := &Handler{
		logger:         logger,
		metricsService: metricsService,
		dbPingService:  dbPingService,
		auditService:   auditService,

		metricValidator: validators.NewMetricsValidator(),
		hashKey:         cfg.HashKey,
	}

	if privateKey != nil {
		handler.privateKey = privateKey
	}

	utils.InitHasherPool(cfg.HashKey)

	return handler
}

func (h *Handler) Init() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(h.WithLogging)
	if h.privateKey != nil {
		router.Use(h.WithDecryption)
	}

	router.Mount("/debug", middleware.Profiler())

	// metrics handlers group
	router.Group(func(r chi.Router) {
		r.Use(GZip, h.WithHashing)

		r.Post("/value/", h.GetMetricJSON)
		r.Get("/", h.GetAllMetrics)
		r.Get("/value/{metricType}/{metricName}", h.GetMetricValue)

		r.Group(func(audit chi.Router) {
			if h.auditService != nil {
				audit.Use(h.Audit)
			}
			audit.Post("/updates/", h.BatchUpdateMetricJSON)
			audit.Post("/update/", h.UpdateMetricJSON)
			audit.Post("/update/{metricType}/{metricName}/{metricValue}", h.MetricHandler)
		})
	})

	// database ping group
	router.Group(func(r chi.Router) {
		r.Use(h.DatabaseConnectionCheck)
		r.Get("/ping", h.Ping)
	})

	router.MethodNotAllowed(CheckHTTPMethod(router))

	return router
}
