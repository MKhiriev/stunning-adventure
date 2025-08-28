package server

import (
	"net/http"

	"github.com/MKhiriev/stunning-adventure/internal/config"
)

type Server struct {
	server *http.Server
}

func (s *Server) ServerRun(handler http.Handler, cfg *config.ServerConfig) error {
	s.server = &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: handler,
	}

	return s.server.ListenAndServe()
}
