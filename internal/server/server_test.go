package server

import (
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/config"
)

func TestHTTPServerRun_ShutdownGracefully(t *testing.T) {
	s := &Server{}

	// минимальный сервер, просто отвечает 200 OK
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := &config.ServerConfig{
		ServerAddress: "127.0.0.1:0", // :0 -> случайный порт
	}
	s.HTTPServer(handler, cfg)

	// запускаем сервер в отдельной горутине
	done := make(chan error, 1)
	go func() {
		err := s.ServerRun()
		done <- err
	}()

	// ждём немного, чтобы сервер успел стартовать
	time.Sleep(100 * time.Millisecond)

	// имитируем сигнал завершения
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(syscall.SIGINT)

	// ждём завершения ServerRun
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shutdown in time")
	}
}
