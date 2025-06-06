package handlers

import (
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"net/http"
)

type Handler struct {
	*store.MemStorage
}

func NewHandler() *Handler {
	return &Handler{MemStorage: store.NewMemStorage()}
}

func (h *Handler) Init() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", h.MetricHandler) // {metric_type}/{metric_name}/{metric_value}
	return mux
}
