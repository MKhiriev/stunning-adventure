package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/config"
	"google.golang.org/grpc"
)

const timeout = 10 * time.Second

type Server struct {
	server          *http.Server
	gRPCServer      *grpc.Server
	gRPCNetListener net.Listener
}

func (s *Server) ServerRun() error {
	if s.server == nil && s.gRPCServer == nil {
		fmt.Println("nothing to run!")
		return nil
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

		// finish HTTP server
		s.shutdownHTTPServer()

		// finish gRPC server
		s.shutdownGRPCServer()

		close(idleConnsClosed)
	}()

	if s.server != nil {
		fmt.Println("Launching HTTP Server")
		go s.launchHTTPServer()
	}
	if s.gRPCServer != nil {
		fmt.Println("Launching GRPC Server")
		go s.launchGRPCServer()
	}

	<-idleConnsClosed
	fmt.Println("Server Shutdown gracefully")

	return nil
}

func (s *Server) HTTPServer(handler http.Handler, cfg *config.ServerConfig) {
	s.server = &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      handler,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}
}

func (s *Server) GRPCServer(gRPCServer *grpc.Server, gRPCNetListener net.Listener) {
	s.gRPCServer = gRPCServer
	s.gRPCNetListener = gRPCNetListener
}

func (s *Server) launchHTTPServer() {
	if err := s.server.ListenAndServe(); err != nil {
		fmt.Printf("HTTP server ListenAndServe: %v\n", err)
	}
}

func (s *Server) launchGRPCServer() {
	if err := s.gRPCServer.Serve(s.gRPCNetListener); err != nil {
		fmt.Printf("gRPC server Serve: %v\n", err)
	}
}

func (s *Server) shutdownHTTPServer() {
	if err := s.server.Shutdown(context.Background()); s.server != nil && err != nil {
		// ошибки закрытия Listener
		fmt.Printf("HTTP server Shutdown: %v\n", err)
	}
}

func (s *Server) shutdownGRPCServer() {
	if s.gRPCServer != nil {
		s.gRPCServer.GracefulStop()
	}
}
