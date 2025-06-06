package server

import (
	"net/http"
)

type Server struct {
	server *http.Server
}

func (s *Server) ServerRun(handler http.Handler, port string) error {
	s.server = &http.Server{
		Addr:    "localhost:" + port,
		Handler: handler,
	}

	return s.server.ListenAndServe()
}
