package server

import (
	"net/http"
)

type Server struct {
	server *http.Server
}

func (s *Server) ServerRun(handler http.Handler, serverAddress string) error {
	s.server = &http.Server{
		Addr:    serverAddress,
		Handler: handler,
	}

	return s.server.ListenAndServe()
}
