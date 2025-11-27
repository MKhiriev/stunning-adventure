package server

import (
	"net/http"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/config"
)

const timeout = 10 * time.Second

type Server struct {
	server *http.Server
}

func (s *Server) ServerRun(handler http.Handler, cfg *config.ServerConfig) error {
	s.server = &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      handler,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	return s.server.ListenAndServe()
}
