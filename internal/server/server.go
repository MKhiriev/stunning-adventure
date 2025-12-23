package server

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
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

	idleConnsClosed := make(chan struct{})
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()
	go func() {
		<-ctx.Done()

		if err := s.server.Shutdown(context.Background()); err != nil {
			// ошибки закрытия Listener
			fmt.Printf("HTTP server Shutdown: %v\n", err)
		}

		close(idleConnsClosed)
	}()

	if err := s.server.ListenAndServe(); err != nil {
		fmt.Printf("HTTP server ListenAndServe: %v\n", err)
	}

	<-idleConnsClosed
	fmt.Println("Server Shutdown gracefully")

	return nil
}
